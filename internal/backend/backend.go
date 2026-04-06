package backend

import (
	"context"

	"github.com/sennadevos/kosa/internal/domain"
)

// Backend abstracts all database operations. The application layer depends
// only on this interface — it never talks to a specific database or API directly.
type Backend interface {
	// Accounts
	ListAccounts(ctx context.Context, opts domain.AccountFilter) ([]domain.Account, error)
	GetAccount(ctx context.Context, id string) (*domain.Account, error)
	CreateAccount(ctx context.Context, a domain.AccountInput) (*domain.Account, error)
	UpdateAccount(ctx context.Context, id string, a domain.AccountInput) (*domain.Account, error)

	// Transactions
	ListTransactions(ctx context.Context, opts domain.TransactionFilter) ([]domain.Transaction, error)
	GetTransaction(ctx context.Context, id string) (*domain.Transaction, error)
	CreateTransaction(ctx context.Context, t domain.TransactionInput) (*domain.Transaction, error)
	UpdateTransaction(ctx context.Context, id string, t domain.TransactionInput) (*domain.Transaction, error)
	DeleteTransaction(ctx context.Context, id string) error

	// Recurring Rules
	ListRecurringRules(ctx context.Context, opts domain.RecurringRuleFilter) ([]domain.RecurringRule, error)
	GetRecurringRule(ctx context.Context, id string) (*domain.RecurringRule, error)
	CreateRecurringRule(ctx context.Context, r domain.RecurringRuleInput) (*domain.RecurringRule, error)
	UpdateRecurringRule(ctx context.Context, id string, r domain.RecurringRuleInput) (*domain.RecurringRule, error)

	// Loans
	ListLoans(ctx context.Context, opts domain.LoanFilter) ([]domain.Loan, error)
	GetLoan(ctx context.Context, id string) (*domain.Loan, error)
	CreateLoan(ctx context.Context, l domain.LoanInput) (*domain.Loan, error)
	UpdateLoan(ctx context.Context, id string, l domain.LoanInput) (*domain.Loan, error)

	// Loan Payments
	ListLoanPayments(ctx context.Context, loanID string) ([]domain.LoanPayment, error)
	CreateLoanPayment(ctx context.Context, p domain.LoanPaymentInput) (*domain.LoanPayment, error)

	// Balance Snapshots
	ListSnapshots(ctx context.Context, accountID string) ([]domain.BalanceSnapshot, error)
	LatestSnapshot(ctx context.Context, accountID string) (*domain.BalanceSnapshot, error)
	CreateSnapshot(ctx context.Context, s domain.SnapshotInput) (*domain.BalanceSnapshot, error)

	// Categories
	ListCategories(ctx context.Context) ([]domain.Category, error)
	GetCategoryByName(ctx context.Context, name string) (*domain.Category, error)

	// Tags
	ListTags(ctx context.Context) ([]domain.Tag, error)
	GetTagByName(ctx context.Context, name string) (*domain.Tag, error)
	GetOrCreateTag(ctx context.Context, name string) (*domain.Tag, error)
}
