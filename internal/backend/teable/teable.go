package teable

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sennadevos/kosa/internal/backend"
	"github.com/sennadevos/kosa/internal/config"
	"github.com/sennadevos/kosa/internal/domain"
)

// Backend implements backend.Backend using the Teable REST API.
type Backend struct {
	client *Client
	cfg    config.TeableConfig
}

func New(cfg config.TeableConfig) *Backend {
	return &Backend{
		client: NewClient(cfg.URL, cfg.Token),
		cfg:    cfg,
	}
}

func (b *Backend) tableID(name string) string {
	return b.cfg.Tables[name]
}

func (b *Backend) fieldMap(table string) FieldMap {
	fm := b.cfg.Fields[table]
	if fm == nil {
		return FieldMap{}
	}
	return FieldMap(fm)
}

// filter builder helper
func filterJSON(conditions []filterCondition) string {
	if len(conditions) == 0 {
		return ""
	}
	f := filterExpr{
		Conjunction: "and",
		Conditions:  conditions,
	}
	data, _ := json.Marshal(f)
	return string(data)
}

type filterExpr struct {
	Conjunction string            `json:"conjunction"`
	Conditions  []filterCondition `json:"filterSet"`
}

type filterCondition struct {
	FieldID  string      `json:"fieldId"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// Accounts

func (b *Backend) ListAccounts(ctx context.Context, opts domain.AccountFilter) ([]domain.Account, error) {
	var conditions []filterCondition
	if opts.Type != nil {
		conditions = append(conditions, filterCondition{
			FieldID:  b.fieldMap("accounts")["type"],
			Operator: "is",
			Value:    string(*opts.Type),
		})
	}

	records, err := b.client.ListRecords(ctx, b.tableID("accounts"), filterJSON(conditions), "", 0)
	if err != nil {
		return nil, err
	}

	out := make([]domain.Account, len(records))
	fm := b.fieldMap("accounts")
	for i, r := range records {
		out[i] = mapFieldsToAccount(r.Fields, fm)
		out[i].ID = r.ID
	}
	return out, nil
}

func (b *Backend) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	rec, err := b.client.GetRecord(ctx, b.tableID("accounts"), id)
	if err != nil {
		return nil, err
	}
	a := mapFieldsToAccount(rec.Fields, b.fieldMap("accounts"))
	a.ID = rec.ID
	return &a, nil
}

func (b *Backend) CreateAccount(ctx context.Context, a domain.AccountInput) (*domain.Account, error) {
	fields := mapAccountToFields(a, b.fieldMap("accounts"))
	rec, err := b.client.CreateRecord(ctx, b.tableID("accounts"), fields)
	if err != nil {
		return nil, err
	}
	acc := mapFieldsToAccount(rec.Fields, b.fieldMap("accounts"))
	acc.ID = rec.ID
	return &acc, nil
}

func (b *Backend) UpdateAccount(ctx context.Context, id string, a domain.AccountInput) (*domain.Account, error) {
	fields := mapAccountToFields(a, b.fieldMap("accounts"))
	rec, err := b.client.UpdateRecord(ctx, b.tableID("accounts"), id, fields)
	if err != nil {
		return nil, err
	}
	acc := mapFieldsToAccount(rec.Fields, b.fieldMap("accounts"))
	acc.ID = rec.ID
	return &acc, nil
}

// Transactions

func (b *Backend) ListTransactions(ctx context.Context, opts domain.TransactionFilter) ([]domain.Transaction, error) {
	fm := b.fieldMap("transactions")
	var conditions []filterCondition

	if opts.AccountID != "" {
		conditions = append(conditions, filterCondition{
			FieldID: fm["account"], Operator: "is", Value: opts.AccountID,
		})
	}
	if opts.Type != nil {
		conditions = append(conditions, filterCondition{
			FieldID: fm["type"], Operator: "is", Value: string(*opts.Type),
		})
	}
	if opts.LoanID != "" {
		conditions = append(conditions, filterCondition{
			FieldID: fm["loan"], Operator: "is", Value: opts.LoanID,
		})
	}
	if opts.DateFrom != nil {
		conditions = append(conditions, filterCondition{
			FieldID: fm["date"], Operator: "isAfter", Value: opts.DateFrom.Format("2006-01-02"),
		})
	}
	if opts.DateTo != nil {
		conditions = append(conditions, filterCondition{
			FieldID: fm["date"], Operator: "isBefore", Value: opts.DateTo.Format("2006-01-02"),
		})
	}

	records, err := b.client.ListRecords(ctx, b.tableID("transactions"), filterJSON(conditions), "", opts.Limit)
	if err != nil {
		return nil, err
	}

	out := make([]domain.Transaction, len(records))
	for i, r := range records {
		out[i] = mapFieldsToTransaction(r.Fields, fm)
		out[i].ID = r.ID
	}
	return out, nil
}

func (b *Backend) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	rec, err := b.client.GetRecord(ctx, b.tableID("transactions"), id)
	if err != nil {
		return nil, err
	}
	t := mapFieldsToTransaction(rec.Fields, b.fieldMap("transactions"))
	t.ID = rec.ID
	return &t, nil
}

func (b *Backend) CreateTransaction(ctx context.Context, t domain.TransactionInput) (*domain.Transaction, error) {
	fields := mapTransactionToFields(t, b.fieldMap("transactions"))
	rec, err := b.client.CreateRecord(ctx, b.tableID("transactions"), fields)
	if err != nil {
		return nil, err
	}
	txn := mapFieldsToTransaction(rec.Fields, b.fieldMap("transactions"))
	txn.ID = rec.ID
	return &txn, nil
}

func (b *Backend) UpdateTransaction(ctx context.Context, id string, t domain.TransactionInput) (*domain.Transaction, error) {
	fields := mapTransactionToFields(t, b.fieldMap("transactions"))
	rec, err := b.client.UpdateRecord(ctx, b.tableID("transactions"), id, fields)
	if err != nil {
		return nil, err
	}
	txn := mapFieldsToTransaction(rec.Fields, b.fieldMap("transactions"))
	txn.ID = rec.ID
	return &txn, nil
}

func (b *Backend) DeleteTransaction(ctx context.Context, id string) error {
	return b.client.DeleteRecord(ctx, b.tableID("transactions"), id)
}

// Recurring Rules

func (b *Backend) ListRecurringRules(ctx context.Context, opts domain.RecurringRuleFilter) ([]domain.RecurringRule, error) {
	fm := b.fieldMap("recurring_rules")
	var conditions []filterCondition
	if opts.ActiveOnly {
		conditions = append(conditions, filterCondition{
			FieldID: fm["is_active"], Operator: "is", Value: true,
		})
	}

	records, err := b.client.ListRecords(ctx, b.tableID("recurring_rules"), filterJSON(conditions), "", 0)
	if err != nil {
		return nil, err
	}

	out := make([]domain.RecurringRule, len(records))
	for i, r := range records {
		out[i] = mapFieldsToRecurringRule(r.Fields, fm)
		out[i].ID = r.ID
	}
	return out, nil
}

func (b *Backend) GetRecurringRule(ctx context.Context, id string) (*domain.RecurringRule, error) {
	rec, err := b.client.GetRecord(ctx, b.tableID("recurring_rules"), id)
	if err != nil {
		return nil, err
	}
	r := mapFieldsToRecurringRule(rec.Fields, b.fieldMap("recurring_rules"))
	r.ID = rec.ID
	return &r, nil
}

func (b *Backend) CreateRecurringRule(ctx context.Context, r domain.RecurringRuleInput) (*domain.RecurringRule, error) {
	fields := mapRecurringRuleToFields(r, b.fieldMap("recurring_rules"))
	rec, err := b.client.CreateRecord(ctx, b.tableID("recurring_rules"), fields)
	if err != nil {
		return nil, err
	}
	rule := mapFieldsToRecurringRule(rec.Fields, b.fieldMap("recurring_rules"))
	rule.ID = rec.ID
	return &rule, nil
}

func (b *Backend) UpdateRecurringRule(ctx context.Context, id string, r domain.RecurringRuleInput) (*domain.RecurringRule, error) {
	fields := mapRecurringRuleToFields(r, b.fieldMap("recurring_rules"))
	rec, err := b.client.UpdateRecord(ctx, b.tableID("recurring_rules"), id, fields)
	if err != nil {
		return nil, err
	}
	rule := mapFieldsToRecurringRule(rec.Fields, b.fieldMap("recurring_rules"))
	rule.ID = rec.ID
	return &rule, nil
}

// Loans

func (b *Backend) ListLoans(ctx context.Context, opts domain.LoanFilter) ([]domain.Loan, error) {
	fm := b.fieldMap("loans")
	var conditions []filterCondition
	if opts.Type != nil {
		conditions = append(conditions, filterCondition{
			FieldID: fm["type"], Operator: "is", Value: string(*opts.Type),
		})
	}
	if opts.Settled != nil {
		conditions = append(conditions, filterCondition{
			FieldID: fm["is_settled"], Operator: "is", Value: *opts.Settled,
		})
	}

	records, err := b.client.ListRecords(ctx, b.tableID("loans"), filterJSON(conditions), "", 0)
	if err != nil {
		return nil, err
	}

	out := make([]domain.Loan, len(records))
	for i, r := range records {
		out[i] = mapFieldsToLoan(r.Fields, fm)
		out[i].ID = r.ID
	}
	return out, nil
}

func (b *Backend) GetLoan(ctx context.Context, id string) (*domain.Loan, error) {
	rec, err := b.client.GetRecord(ctx, b.tableID("loans"), id)
	if err != nil {
		return nil, err
	}
	l := mapFieldsToLoan(rec.Fields, b.fieldMap("loans"))
	l.ID = rec.ID
	return &l, nil
}

func (b *Backend) CreateLoan(ctx context.Context, l domain.LoanInput) (*domain.Loan, error) {
	fields := mapLoanToFields(l, b.fieldMap("loans"))
	rec, err := b.client.CreateRecord(ctx, b.tableID("loans"), fields)
	if err != nil {
		return nil, err
	}
	loan := mapFieldsToLoan(rec.Fields, b.fieldMap("loans"))
	loan.ID = rec.ID
	return &loan, nil
}

func (b *Backend) UpdateLoan(ctx context.Context, id string, l domain.LoanInput) (*domain.Loan, error) {
	fields := mapLoanToFields(l, b.fieldMap("loans"))
	rec, err := b.client.UpdateRecord(ctx, b.tableID("loans"), id, fields)
	if err != nil {
		return nil, err
	}
	loan := mapFieldsToLoan(rec.Fields, b.fieldMap("loans"))
	loan.ID = rec.ID
	return &loan, nil
}

// Loan Payments

func (b *Backend) ListLoanPayments(ctx context.Context, loanID string) ([]domain.LoanPayment, error) {
	fm := b.fieldMap("loan_payments")
	conditions := []filterCondition{{
		FieldID: fm["loan"], Operator: "is", Value: loanID,
	}}

	records, err := b.client.ListRecords(ctx, b.tableID("loan_payments"), filterJSON(conditions), "", 0)
	if err != nil {
		return nil, err
	}

	out := make([]domain.LoanPayment, len(records))
	for i, r := range records {
		out[i] = mapFieldsToLoanPayment(r.Fields, fm)
		out[i].ID = r.ID
	}
	return out, nil
}

func (b *Backend) CreateLoanPayment(ctx context.Context, p domain.LoanPaymentInput) (*domain.LoanPayment, error) {
	fields := mapLoanPaymentToFields(p, b.fieldMap("loan_payments"))
	rec, err := b.client.CreateRecord(ctx, b.tableID("loan_payments"), fields)
	if err != nil {
		return nil, err
	}
	payment := mapFieldsToLoanPayment(rec.Fields, b.fieldMap("loan_payments"))
	payment.ID = rec.ID
	return &payment, nil
}

// Balance Snapshots

func (b *Backend) ListSnapshots(ctx context.Context, accountID string) ([]domain.BalanceSnapshot, error) {
	fm := b.fieldMap("balance_snapshots")
	conditions := []filterCondition{{
		FieldID: fm["account"], Operator: "is", Value: accountID,
	}}

	records, err := b.client.ListRecords(ctx, b.tableID("balance_snapshots"), filterJSON(conditions), "", 0)
	if err != nil {
		return nil, err
	}

	out := make([]domain.BalanceSnapshot, len(records))
	for i, r := range records {
		out[i] = mapFieldsToSnapshot(r.Fields, fm)
		out[i].ID = r.ID
	}
	return out, nil
}

func (b *Backend) LatestSnapshot(ctx context.Context, accountID string) (*domain.BalanceSnapshot, error) {
	snaps, err := b.ListSnapshots(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return nil, fmt.Errorf("no snapshots for account %s", accountID)
	}
	latest := snaps[0]
	for _, s := range snaps[1:] {
		if s.Date.After(latest.Date) {
			latest = s
		}
	}
	return &latest, nil
}

func (b *Backend) CreateSnapshot(ctx context.Context, s domain.SnapshotInput) (*domain.BalanceSnapshot, error) {
	fields := mapSnapshotToFields(s, b.fieldMap("balance_snapshots"))
	rec, err := b.client.CreateRecord(ctx, b.tableID("balance_snapshots"), fields)
	if err != nil {
		return nil, err
	}
	snap := mapFieldsToSnapshot(rec.Fields, b.fieldMap("balance_snapshots"))
	snap.ID = rec.ID
	return &snap, nil
}

// Categories

func (b *Backend) ListCategories(ctx context.Context) ([]domain.Category, error) {
	fm := b.fieldMap("categories")
	records, err := b.client.ListRecords(ctx, b.tableID("categories"), "", "", 0)
	if err != nil {
		return nil, err
	}

	out := make([]domain.Category, len(records))
	for i, r := range records {
		out[i] = mapFieldsToCategory(r.Fields, fm)
		out[i].ID = r.ID
	}
	return out, nil
}

func (b *Backend) GetCategoryByName(ctx context.Context, name string) (*domain.Category, error) {
	cats, err := b.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range cats {
		if strings.EqualFold(c.Name, name) {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("category %q not found", name)
}

// Tags

func (b *Backend) ListTags(ctx context.Context) ([]domain.Tag, error) {
	fm := b.fieldMap("tags")
	records, err := b.client.ListRecords(ctx, b.tableID("tags"), "", "", 0)
	if err != nil {
		return nil, err
	}

	out := make([]domain.Tag, len(records))
	for i, r := range records {
		out[i] = mapFieldsToTag(r.Fields, fm)
		out[i].ID = r.ID
	}
	return out, nil
}

func (b *Backend) GetTagByName(ctx context.Context, name string) (*domain.Tag, error) {
	tags, err := b.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range tags {
		if strings.EqualFold(t.Name, name) {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("tag %q not found", name)
}

func (b *Backend) GetOrCreateTag(ctx context.Context, name string) (*domain.Tag, error) {
	t, err := b.GetTagByName(ctx, name)
	if err == nil {
		return t, nil
	}
	// create it
	fm := b.fieldMap("tags")
	fields := map[string]interface{}{
		fm["name"]: name,
	}
	rec, err := b.client.CreateRecord(ctx, b.tableID("tags"), fields)
	if err != nil {
		return nil, err
	}
	tag := mapFieldsToTag(rec.Fields, fm)
	tag.ID = rec.ID
	return &tag, nil
}

// compile-time check
var _ backend.Backend = (*Backend)(nil)
