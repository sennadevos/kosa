package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sennadevos/kosa/internal/domain"
)

type SplitInput struct {
	TotalAmount domain.Amount
	Description string
	Category    string
	Account     string
	// Friends maps counterparty name to their share amount.
	// If nil/empty with FriendNames set, splits equally.
	Friends     map[string]domain.Amount
	FriendNames []string
	MyShare     *domain.Amount // explicit personal share; nil = auto-calculate
}

type SplitResult struct {
	PersonalExpense *domain.Transaction
	Loans           []*domain.Loan
	LoanExpenses    []*domain.Transaction
}

func (a *App) Split(ctx context.Context, in SplitInput) (*SplitResult, error) {
	accountID, err := a.resolveAccountID(ctx, in.Account)
	if err != nil {
		return nil, err
	}
	catID, err := a.resolveCategoryID(ctx, in.Category)
	if err != nil {
		return nil, err
	}

	// determine shares
	friends := in.Friends
	if friends == nil || len(friends) == 0 {
		if len(in.FriendNames) == 0 {
			return nil, fmt.Errorf("no friends specified for split")
		}
		// equal split
		totalPeople := len(in.FriendNames) + 1 // +1 for you
		perPerson := in.TotalAmount.Div(domain.NewAmountFromInt(int64(totalPeople)))
		friends = make(map[string]domain.Amount, len(in.FriendNames))
		for _, name := range in.FriendNames {
			friends[name] = perPerson
		}
	}

	// calculate my share
	friendsTotal := domain.ZeroAmount()
	for _, share := range friends {
		friendsTotal = friendsTotal.Add(share)
	}

	myShare := in.TotalAmount.Sub(friendsTotal)
	if in.MyShare != nil {
		myShare = *in.MyShare
	}

	now := time.Now()
	result := &SplitResult{}

	// 1. create personal expense (my share)
	if myShare.IsPositive() {
		txn, err := a.Backend.CreateTransaction(ctx, domain.TransactionInput{
			Date:        now,
			Type:        domain.TransactionExpense,
			Amount:      myShare,
			Description: in.Description,
			CategoryID:  catID,
			AccountID:   accountID,
		})
		if err != nil {
			return nil, fmt.Errorf("creating personal expense: %w", err)
		}
		result.PersonalExpense = txn
	}

	// 2. for each friend: create loan + loan-linked expense
	for name, share := range friends {
		loan, err := a.Backend.CreateLoan(ctx, domain.LoanInput{
			Type:             domain.LoanReceivable,
			CounterpartyName: name,
			Description:      in.Description,
			OriginalAmount:   share,
			DateCreated:      now,
			InterestType:     domain.InterestNone,
		})
		if err != nil {
			return nil, fmt.Errorf("creating loan for %s: %w", name, err)
		}
		result.Loans = append(result.Loans, loan)

		txn, err := a.Backend.CreateTransaction(ctx, domain.TransactionInput{
			Date:        now,
			Type:        domain.TransactionExpense,
			Amount:      share,
			Description: in.Description,
			CategoryID:  catID,
			AccountID:   accountID,
			LoanID:      loan.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("creating loan expense for %s: %w", name, err)
		}
		result.LoanExpenses = append(result.LoanExpenses, txn)
	}

	return result, nil
}
