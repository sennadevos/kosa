package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

type LoanNewInput struct {
	Type             domain.LoanType
	CounterpartyName string
	CounterpartyURI  string
	Description      string
	Amount           domain.Amount
	Currency         string
	DateCreated      time.Time
	DueDate          *time.Time
	InterestType     domain.InterestType
	InterestRate     domain.Amount
	InterestPeriod   domain.InterestPeriod
	Notes            string
}

func (a *App) LoanNew(ctx context.Context, in LoanNewInput) (*domain.Loan, error) {
	interestType := in.InterestType
	if interestType == "" {
		interestType = domain.InterestNone
	}

	dateCreated := in.DateCreated
	if dateCreated.IsZero() {
		dateCreated = time.Now()
	}

	loan, err := a.Backend.CreateLoan(ctx, domain.LoanInput{
		Type:             in.Type,
		CounterpartyName: in.CounterpartyName,
		CounterpartyURI:  in.CounterpartyURI,
		Description:      in.Description,
		OriginalAmount:   in.Amount,
		Currency:         in.Currency,
		DateCreated:      dateCreated,
		DueDate:          in.DueDate,
		InterestType:     interestType,
		InterestRate:     in.InterestRate,
		InterestPeriod:   in.InterestPeriod,
		Notes:            in.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("creating loan: %w", err)
	}
	return loan, nil
}

type LoanPayInput struct {
	LoanID  string
	Amount  domain.Amount
	Account string
	Notes   string
}

func (a *App) LoanPay(ctx context.Context, in LoanPayInput) (*domain.LoanPayment, *domain.Transaction, error) {
	loan, err := a.Backend.GetLoan(ctx, in.LoanID)
	if err != nil {
		return nil, nil, fmt.Errorf("loan: %w", err)
	}

	accountID, err := a.ResolveAccountID(ctx, in.Account)
	if err != nil {
		return nil, nil, err
	}

	// create payment
	payment, err := a.Backend.CreateLoanPayment(ctx, domain.LoanPaymentInput{
		LoanID:    in.LoanID,
		Date:      time.Now(),
		Amount:    in.Amount,
		AccountID: accountID,
		Notes:     in.Notes,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("creating payment: %w", err)
	}

	// create corresponding transaction
	txnType := domain.TransactionExpense
	if loan.Type == domain.LoanReceivable {
		txnType = domain.TransactionIncome
	}

	txn, err := a.Backend.CreateTransaction(ctx, domain.TransactionInput{
		Date:        time.Now(),
		Type:        txnType,
		Amount:      in.Amount,
		Description: fmt.Sprintf("Loan payment: %s", loan.Description),
		AccountID:   accountID,
		LoanID:      in.LoanID,
	})
	if err != nil {
		return payment, nil, fmt.Errorf("creating loan transaction: %w", err)
	}

	// check if loan is settled
	payments, _ := a.Backend.ListLoanPayments(ctx, in.LoanID)
	status, _ := ReplayLoan(loan, payments)
	if status.Remaining.IsZero() {
		a.Backend.UpdateLoan(ctx, in.LoanID, domain.LoanInput{
			Type:             loan.Type,
			CounterpartyName: loan.CounterpartyName,
			CounterpartyURI:  loan.CounterpartyURI,
			Description:      loan.Description,
			OriginalAmount:   loan.OriginalAmount,
			Currency:         loan.Currency,
			DateCreated:      loan.DateCreated,
			DueDate:          loan.DueDate,
			InterestType:     loan.InterestType,
			InterestRate:     loan.InterestRate,
			InterestPeriod:   loan.InterestPeriod,
			IsSettled:        true,
			Notes:            loan.Notes,
		})
	}

	return payment, txn, nil
}

func (a *App) LoanShow(ctx context.Context, loanID string) (*domain.Loan, *domain.LoanStatus, []domain.LoanTimelineEntry, error) {
	loan, err := a.Backend.GetLoan(ctx, loanID)
	if err != nil {
		return nil, nil, nil, err
	}
	payments, err := a.Backend.ListLoanPayments(ctx, loanID)
	if err != nil {
		return nil, nil, nil, err
	}
	status, timeline := ReplayLoan(loan, payments)
	return loan, status, timeline, nil
}

func (a *App) Owe(ctx context.Context, amount domain.Amount, description, counterparty string, date time.Time) (*domain.Loan, error) {
	return a.LoanNew(ctx, LoanNewInput{
		Type:             domain.LoanPayable,
		CounterpartyName: counterparty,
		Description:      description,
		Amount:           amount,
		DateCreated:      date,
	})
}

func (a *App) Lent(ctx context.Context, amount domain.Amount, description, counterparty string, date time.Time) (*domain.Loan, error) {
	return a.LoanNew(ctx, LoanNewInput{
		Type:             domain.LoanReceivable,
		CounterpartyName: counterparty,
		Description:      description,
		Amount:           amount,
		DateCreated:      date,
	})
}
