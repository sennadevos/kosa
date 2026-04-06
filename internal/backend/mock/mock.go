package mock

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sennadevos/kosa/internal/backend"
	"github.com/sennadevos/kosa/internal/domain"
)

// Backend is an in-memory implementation of backend.Backend for testing.
type Backend struct {
	mu            sync.Mutex
	nextID        int
	accounts      []domain.Account
	transactions  []domain.Transaction
	recurringRules []domain.RecurringRule
	loans         []domain.Loan
	loanPayments  []domain.LoanPayment
	snapshots     []domain.BalanceSnapshot
	categories    []domain.Category
	tags          []domain.Tag
}

func New() *Backend {
	return &Backend{}
}

func (b *Backend) genID() string {
	b.nextID++
	return fmt.Sprintf("mock_%d", b.nextID)
}

// Accounts

func (b *Backend) ListAccounts(_ context.Context, opts domain.AccountFilter) ([]domain.Account, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []domain.Account
	for _, a := range b.accounts {
		if opts.Type != nil && a.Type != *opts.Type {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

func (b *Backend) GetAccount(_ context.Context, id string) (*domain.Account, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, a := range b.accounts {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("account %s not found", id)
}

func (b *Backend) CreateAccount(_ context.Context, a domain.AccountInput) (*domain.Account, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	acc := domain.Account{
		ID:        b.genID(),
		Name:      a.Name,
		Type:      a.Type,
		Provider:  a.Provider,
		Currency:  a.Currency,
		IBAN:      a.IBAN,
		IsDefault: a.IsDefault,
		Notes:     a.Notes,
	}
	b.accounts = append(b.accounts, acc)
	return &acc, nil
}

func (b *Backend) UpdateAccount(_ context.Context, id string, a domain.AccountInput) (*domain.Account, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, acc := range b.accounts {
		if acc.ID == id {
			b.accounts[i].Name = a.Name
			b.accounts[i].Type = a.Type
			b.accounts[i].Provider = a.Provider
			b.accounts[i].Currency = a.Currency
			b.accounts[i].IBAN = a.IBAN
			b.accounts[i].IsDefault = a.IsDefault
			b.accounts[i].Notes = a.Notes
			return &b.accounts[i], nil
		}
	}
	return nil, fmt.Errorf("account %s not found", id)
}

// Transactions

func (b *Backend) ListTransactions(_ context.Context, opts domain.TransactionFilter) ([]domain.Transaction, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []domain.Transaction
	for _, t := range b.transactions {
		if opts.AccountID != "" && t.AccountID != opts.AccountID {
			continue
		}
		if opts.CategoryID != "" && t.CategoryID != opts.CategoryID {
			continue
		}
		if opts.Type != nil && t.Type != *opts.Type {
			continue
		}
		if opts.LoanID != "" && t.LoanID != opts.LoanID {
			continue
		}
		if opts.DateFrom != nil && t.Date.Before(*opts.DateFrom) {
			continue
		}
		if opts.DateTo != nil && t.Date.After(*opts.DateTo) {
			continue
		}
		if opts.TagID != "" {
			found := false
			for _, tid := range t.TagIDs {
				if tid == opts.TagID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		out = append(out, t)
	}
	if opts.Limit > 0 && len(out) > opts.Limit {
		out = out[len(out)-opts.Limit:]
	}
	return out, nil
}

func (b *Backend) GetTransaction(_ context.Context, id string) (*domain.Transaction, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, t := range b.transactions {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("transaction %s not found", id)
}

func (b *Backend) CreateTransaction(_ context.Context, t domain.TransactionInput) (*domain.Transaction, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	txn := domain.Transaction{
		ID:              b.genID(),
		Date:            t.Date,
		Type:            t.Type,
		Amount:          t.Amount,
		Description:     t.Description,
		CategoryID:      t.CategoryID,
		TagIDs:          t.TagIDs,
		AccountID:       t.AccountID,
		ToAccountID:     t.ToAccountID,
		LoanID:          t.LoanID,
		RecurringRuleID: t.RecurringRuleID,
		RefundOfID:      t.RefundOfID,
		Cashback:        t.Cashback,
		Reference:       t.Reference,
		ForeignAmount:   t.ForeignAmount,
		ForeignCurrency: t.ForeignCurrency,
		ExchangeRate:    t.ExchangeRate,
		Notes:           t.Notes,
	}
	b.transactions = append(b.transactions, txn)
	return &txn, nil
}

func (b *Backend) UpdateTransaction(_ context.Context, id string, t domain.TransactionInput) (*domain.Transaction, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, txn := range b.transactions {
		if txn.ID == id {
			b.transactions[i].Date = t.Date
			b.transactions[i].Type = t.Type
			b.transactions[i].Amount = t.Amount
			b.transactions[i].Description = t.Description
			b.transactions[i].CategoryID = t.CategoryID
			b.transactions[i].TagIDs = t.TagIDs
			b.transactions[i].AccountID = t.AccountID
			b.transactions[i].ToAccountID = t.ToAccountID
			b.transactions[i].LoanID = t.LoanID
			b.transactions[i].RecurringRuleID = t.RecurringRuleID
			b.transactions[i].RefundOfID = t.RefundOfID
			b.transactions[i].Cashback = t.Cashback
			b.transactions[i].Reference = t.Reference
			b.transactions[i].ForeignAmount = t.ForeignAmount
			b.transactions[i].ForeignCurrency = t.ForeignCurrency
			b.transactions[i].ExchangeRate = t.ExchangeRate
			b.transactions[i].Notes = t.Notes
			return &b.transactions[i], nil
		}
	}
	return nil, fmt.Errorf("transaction %s not found", id)
}

func (b *Backend) DeleteTransaction(_ context.Context, id string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, t := range b.transactions {
		if t.ID == id {
			b.transactions = append(b.transactions[:i], b.transactions[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("transaction %s not found", id)
}

// Recurring Rules

func (b *Backend) ListRecurringRules(_ context.Context, opts domain.RecurringRuleFilter) ([]domain.RecurringRule, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []domain.RecurringRule
	for _, r := range b.recurringRules {
		if opts.ActiveOnly && !r.IsActive {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func (b *Backend) GetRecurringRule(_ context.Context, id string) (*domain.RecurringRule, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, r := range b.recurringRules {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("recurring rule %s not found", id)
}

func (b *Backend) CreateRecurringRule(_ context.Context, r domain.RecurringRuleInput) (*domain.RecurringRule, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	rule := domain.RecurringRule{
		ID:         b.genID(),
		Name:       r.Name,
		Type:       r.Type,
		Amount:     r.Amount,
		CategoryID: r.CategoryID,
		TagIDs:     r.TagIDs,
		AccountID:  r.AccountID,
		Frequency:  r.Frequency,
		DayOfMonth: r.DayOfMonth,
		StartDate:  r.StartDate,
		EndDate:    r.EndDate,
		IsActive:   r.IsActive,
		Notes:      r.Notes,
	}
	b.recurringRules = append(b.recurringRules, rule)
	return &rule, nil
}

func (b *Backend) UpdateRecurringRule(_ context.Context, id string, r domain.RecurringRuleInput) (*domain.RecurringRule, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, rule := range b.recurringRules {
		if rule.ID == id {
			b.recurringRules[i].Name = r.Name
			b.recurringRules[i].Type = r.Type
			b.recurringRules[i].Amount = r.Amount
			b.recurringRules[i].CategoryID = r.CategoryID
			b.recurringRules[i].TagIDs = r.TagIDs
			b.recurringRules[i].AccountID = r.AccountID
			b.recurringRules[i].Frequency = r.Frequency
			b.recurringRules[i].DayOfMonth = r.DayOfMonth
			b.recurringRules[i].StartDate = r.StartDate
			b.recurringRules[i].EndDate = r.EndDate
			b.recurringRules[i].IsActive = r.IsActive
			b.recurringRules[i].Notes = r.Notes
			return &b.recurringRules[i], nil
		}
	}
	return nil, fmt.Errorf("recurring rule %s not found", id)
}

// Loans

func (b *Backend) ListLoans(_ context.Context, opts domain.LoanFilter) ([]domain.Loan, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []domain.Loan
	for _, l := range b.loans {
		if opts.Type != nil && l.Type != *opts.Type {
			continue
		}
		if opts.Settled != nil {
			if *opts.Settled && !l.IsSettled {
				continue
			}
			if !*opts.Settled && l.IsSettled {
				continue
			}
		}
		if opts.CounterpartyName != "" && !strings.EqualFold(l.CounterpartyName, opts.CounterpartyName) {
			continue
		}
		out = append(out, l)
	}
	return out, nil
}

func (b *Backend) GetLoan(_ context.Context, id string) (*domain.Loan, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, l := range b.loans {
		if l.ID == id {
			return &l, nil
		}
	}
	return nil, fmt.Errorf("loan %s not found", id)
}

func (b *Backend) CreateLoan(_ context.Context, l domain.LoanInput) (*domain.Loan, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	loan := domain.Loan{
		ID:               b.genID(),
		Type:             l.Type,
		CounterpartyName: l.CounterpartyName,
		CounterpartyURI:  l.CounterpartyURI,
		Description:      l.Description,
		OriginalAmount:   l.OriginalAmount,
		Currency:         l.Currency,
		DateCreated:      l.DateCreated,
		DueDate:          l.DueDate,
		InterestType:     l.InterestType,
		InterestRate:     l.InterestRate,
		InterestPeriod:   l.InterestPeriod,
		Notes:            l.Notes,
	}
	b.loans = append(b.loans, loan)
	return &loan, nil
}

func (b *Backend) UpdateLoan(_ context.Context, id string, l domain.LoanInput) (*domain.Loan, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, loan := range b.loans {
		if loan.ID == id {
			b.loans[i].Type = l.Type
			b.loans[i].CounterpartyName = l.CounterpartyName
			b.loans[i].CounterpartyURI = l.CounterpartyURI
			b.loans[i].Description = l.Description
			b.loans[i].OriginalAmount = l.OriginalAmount
			b.loans[i].Currency = l.Currency
			b.loans[i].DateCreated = l.DateCreated
			b.loans[i].DueDate = l.DueDate
			b.loans[i].InterestType = l.InterestType
			b.loans[i].InterestRate = l.InterestRate
			b.loans[i].InterestPeriod = l.InterestPeriod
			b.loans[i].IsSettled = l.IsSettled
			b.loans[i].Notes = l.Notes
			return &b.loans[i], nil
		}
	}
	return nil, fmt.Errorf("loan %s not found", id)
}

// Loan Payments

func (b *Backend) ListLoanPayments(_ context.Context, loanID string) ([]domain.LoanPayment, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []domain.LoanPayment
	for _, p := range b.loanPayments {
		if p.LoanID == loanID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (b *Backend) CreateLoanPayment(_ context.Context, p domain.LoanPaymentInput) (*domain.LoanPayment, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	payment := domain.LoanPayment{
		ID:        b.genID(),
		LoanID:    p.LoanID,
		Date:      p.Date,
		Amount:    p.Amount,
		AccountID: p.AccountID,
		Notes:     p.Notes,
	}
	b.loanPayments = append(b.loanPayments, payment)
	return &payment, nil
}

// Balance Snapshots

func (b *Backend) ListSnapshots(_ context.Context, accountID string) ([]domain.BalanceSnapshot, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []domain.BalanceSnapshot
	for _, s := range b.snapshots {
		if s.AccountID == accountID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (b *Backend) LatestSnapshot(_ context.Context, accountID string) (*domain.BalanceSnapshot, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var latest *domain.BalanceSnapshot
	for i, s := range b.snapshots {
		if s.AccountID == accountID {
			if latest == nil || s.Date.After(latest.Date) {
				latest = &b.snapshots[i]
			}
		}
	}
	if latest == nil {
		return nil, fmt.Errorf("no snapshots for account %s", accountID)
	}
	return latest, nil
}

func (b *Backend) CreateSnapshot(_ context.Context, s domain.SnapshotInput) (*domain.BalanceSnapshot, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	snap := domain.BalanceSnapshot{
		ID:        b.genID(),
		AccountID: s.AccountID,
		Date:      s.Date,
		Balance:   s.Balance,
		Source:    s.Source,
		Notes:     s.Notes,
	}
	b.snapshots = append(b.snapshots, snap)
	return &snap, nil
}

// Categories

func (b *Backend) ListCategories(_ context.Context) ([]domain.Category, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]domain.Category{}, b.categories...), nil
}

func (b *Backend) GetCategoryByName(_ context.Context, name string) (*domain.Category, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, c := range b.categories {
		if strings.EqualFold(c.Name, name) {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("category %q not found", name)
}

// Tags

func (b *Backend) ListTags(_ context.Context) ([]domain.Tag, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]domain.Tag{}, b.tags...), nil
}

func (b *Backend) GetTagByName(_ context.Context, name string) (*domain.Tag, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, t := range b.tags {
		if strings.EqualFold(t.Name, name) {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("tag %q not found", name)
}

func (b *Backend) GetOrCreateTag(_ context.Context, name string) (*domain.Tag, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, t := range b.tags {
		if strings.EqualFold(t.Name, name) {
			return &t, nil
		}
	}
	tag := domain.Tag{ID: b.genID(), Name: name}
	b.tags = append(b.tags, tag)
	return &tag, nil
}

// Seed helpers for tests

func (b *Backend) SeedAccount(a domain.Account) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if a.ID == "" {
		a.ID = b.genID()
	}
	b.accounts = append(b.accounts, a)
}

func (b *Backend) SeedCategory(c domain.Category) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if c.ID == "" {
		c.ID = b.genID()
	}
	b.categories = append(b.categories, c)
}

func (b *Backend) SeedTag(t domain.Tag) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if t.ID == "" {
		t.ID = b.genID()
	}
	b.tags = append(b.tags, t)
}

func (b *Backend) SeedSnapshot(s domain.BalanceSnapshot) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if s.ID == "" {
		s.ID = b.genID()
	}
	b.snapshots = append(b.snapshots, s)
}

func (b *Backend) SeedTransaction(t domain.Transaction) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if t.ID == "" {
		t.ID = b.genID()
	}
	b.transactions = append(b.transactions, t)
}

func (b *Backend) SeedLoan(l domain.Loan) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if l.ID == "" {
		l.ID = b.genID()
	}
	b.loans = append(b.loans, l)
}

func (b *Backend) SeedLoanPayment(p domain.LoanPayment) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if p.ID == "" {
		p.ID = b.genID()
	}
	b.loanPayments = append(b.loanPayments, p)
}

// compile-time check
var _ backend.Backend = (*Backend)(nil)
