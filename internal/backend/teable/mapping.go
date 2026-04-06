package teable

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/sennadevos/kosa/internal/domain"
)

// FieldMap maps domain field names to Teable field IDs.
type FieldMap map[string]string

// Reverse returns an inverted map: Teable field ID -> domain field name.
func (fm FieldMap) Reverse() FieldMap {
	rev := make(FieldMap, len(fm))
	for k, v := range fm {
		rev[v] = k
	}
	return rev
}

// Helper functions for reading Teable field values

func getString(fields map[string]interface{}, key string) string {
	v, ok := fields[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func getFloat(fields map[string]interface{}, key string) float64 {
	v, ok := fields[key]
	if !ok || v == nil {
		return 0
	}
	f, ok := v.(float64)
	if !ok {
		return 0
	}
	return f
}

func getAmount(fields map[string]interface{}, key string) domain.Amount {
	v, ok := fields[key]
	if !ok || v == nil {
		return domain.ZeroAmount()
	}
	switch val := v.(type) {
	case float64:
		return domain.Amount{Decimal: decimal.NewFromFloat(val)}
	case string:
		a, err := domain.NewAmount(val)
		if err != nil {
			return domain.ZeroAmount()
		}
		return a
	default:
		return domain.ZeroAmount()
	}
}

func getBool(fields map[string]interface{}, key string) bool {
	v, ok := fields[key]
	if !ok || v == nil {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

func getInt(fields map[string]interface{}, key string) int {
	v, ok := fields[key]
	if !ok || v == nil {
		return 0
	}
	f, ok := v.(float64)
	if !ok {
		return 0
	}
	return int(f)
}

func getTime(fields map[string]interface{}, key string) time.Time {
	s := getString(fields, key)
	if s == "" {
		return time.Time{}
	}
	// try common formats
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05.000Z", "2006-01-02"} {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t
		}
	}
	return time.Time{}
}

func getTimePtr(fields map[string]interface{}, key string) *time.Time {
	t := getTime(fields, key)
	if t.IsZero() {
		return nil
	}
	return &t
}

// linkSingle extracts the first record ID from a Teable link field.
// Teable stores link fields as arrays: [{"id": "rec_xxx"}] or ["rec_xxx"].
func linkSingle(v interface{}) string {
	if v == nil {
		return ""
	}
	arr, ok := v.([]interface{})
	if !ok || len(arr) == 0 {
		// might be a direct string
		if s, ok := v.(string); ok {
			return s
		}
		return ""
	}
	// each element could be a map with "id" or a direct string
	first := arr[0]
	if s, ok := first.(string); ok {
		return s
	}
	if m, ok := first.(map[string]interface{}); ok {
		if id, ok := m["id"].(string); ok {
			return id
		}
	}
	return ""
}

// linkMulti extracts multiple record IDs from a Teable link field.
func linkMulti(v interface{}) []string {
	if v == nil {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	var ids []string
	for _, item := range arr {
		if s, ok := item.(string); ok {
			ids = append(ids, s)
		} else if m, ok := item.(map[string]interface{}); ok {
			if id, ok := m["id"].(string); ok {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// wrapLink wraps a single ID for a Teable link field write.
func wrapLink(id string) interface{} {
	if id == "" {
		return nil
	}
	return []map[string]interface{}{{"id": id}}
}

// wrapLinks wraps multiple IDs for a Teable link field write.
func wrapLinks(ids []string) interface{} {
	if len(ids) == 0 {
		return nil
	}
	out := make([]map[string]interface{}, len(ids))
	for i, id := range ids {
		out[i] = map[string]interface{}{"id": id}
	}
	return out
}

// setIfNotZero sets a field only if the value is non-zero.
func setIfNotZero(fields map[string]interface{}, fldID string, v interface{}) {
	if fldID == "" {
		return
	}
	switch val := v.(type) {
	case string:
		if val != "" {
			fields[fldID] = val
		}
	case float64:
		if val != 0 {
			fields[fldID] = val
		}
	case int:
		if val != 0 {
			fields[fldID] = val
		}
	case bool:
		if val {
			fields[fldID] = val
		}
	case domain.Amount:
		if !val.IsZero() {
			f, _ := val.Decimal.Float64()
			fields[fldID] = f
		}
	default:
		if val != nil {
			fields[fldID] = val
		}
	}
}

// Mapping functions: domain <-> Teable fields

func mapFieldsToAccount(raw map[string]interface{}, fm FieldMap) domain.Account {
	rev := fm.Reverse()
	f := make(map[string]interface{}, len(raw))
	for fid, v := range raw {
		if name, ok := rev[fid]; ok {
			f[name] = v
		}
	}
	return domain.Account{
		Name:      getString(f, "name"),
		Type:      domain.AccountType(getString(f, "type")),
		Provider:  getString(f, "provider"),
		Currency:  getString(f, "currency"),
		IBAN:      getString(f, "iban"),
		IsDefault: getBool(f, "is_default"),
		Notes:     getString(f, "notes"),
	}
}

func mapAccountToFields(a domain.AccountInput, fm FieldMap) map[string]interface{} {
	fields := make(map[string]interface{})
	setIfNotZero(fields, fm["name"], a.Name)
	setIfNotZero(fields, fm["type"], string(a.Type))
	setIfNotZero(fields, fm["provider"], a.Provider)
	setIfNotZero(fields, fm["currency"], a.Currency)
	setIfNotZero(fields, fm["iban"], a.IBAN)
	if a.IsDefault {
		fields[fm["is_default"]] = true
	}
	setIfNotZero(fields, fm["notes"], a.Notes)
	return fields
}

func mapFieldsToTransaction(raw map[string]interface{}, fm FieldMap) domain.Transaction {
	rev := fm.Reverse()
	f := make(map[string]interface{}, len(raw))
	for fid, v := range raw {
		if name, ok := rev[fid]; ok {
			f[name] = v
		}
	}
	return domain.Transaction{
		Date:            getTime(f, "date"),
		Type:            domain.TransactionType(getString(f, "type")),
		Amount:          getAmount(f, "amount"),
		Description:     getString(f, "description"),
		CategoryID:      linkSingle(f["category"]),
		TagIDs:          linkMulti(f["tags"]),
		AccountID:       linkSingle(f["account"]),
		ToAccountID:     linkSingle(f["to_account"]),
		LoanID:          linkSingle(f["loan"]),
		RecurringRuleID: linkSingle(f["recurring_rule"]),
		RefundOfID:      linkSingle(f["refund_of"]),
		Cashback:        getAmount(f, "cashback"),
		Reference:       getString(f, "reference"),
		ForeignAmount:   getAmount(f, "foreign_amount"),
		ForeignCurrency: getString(f, "foreign_currency"),
		ExchangeRate:    getAmount(f, "exchange_rate"),
		Notes:           getString(f, "notes"),
	}
}

func mapTransactionToFields(t domain.TransactionInput, fm FieldMap) map[string]interface{} {
	fields := make(map[string]interface{})
	if !t.Date.IsZero() {
		fields[fm["date"]] = t.Date.Format("2006-01-02")
	}
	fields[fm["type"]] = string(t.Type)
	f, _ := t.Amount.Decimal.Float64()
	fields[fm["amount"]] = f
	setIfNotZero(fields, fm["description"], t.Description)
	if t.CategoryID != "" {
		fields[fm["category"]] = wrapLink(t.CategoryID)
	}
	if len(t.TagIDs) > 0 {
		fields[fm["tags"]] = wrapLinks(t.TagIDs)
	}
	if t.AccountID != "" {
		fields[fm["account"]] = wrapLink(t.AccountID)
	}
	if t.ToAccountID != "" {
		fields[fm["to_account"]] = wrapLink(t.ToAccountID)
	}
	if t.LoanID != "" {
		fields[fm["loan"]] = wrapLink(t.LoanID)
	}
	if t.RecurringRuleID != "" {
		fields[fm["recurring_rule"]] = wrapLink(t.RecurringRuleID)
	}
	if t.RefundOfID != "" {
		fields[fm["refund_of"]] = wrapLink(t.RefundOfID)
	}
	setIfNotZero(fields, fm["cashback"], t.Cashback)
	setIfNotZero(fields, fm["reference"], t.Reference)
	setIfNotZero(fields, fm["foreign_amount"], t.ForeignAmount)
	setIfNotZero(fields, fm["foreign_currency"], t.ForeignCurrency)
	setIfNotZero(fields, fm["exchange_rate"], t.ExchangeRate)
	setIfNotZero(fields, fm["notes"], t.Notes)
	return fields
}

func mapFieldsToLoan(raw map[string]interface{}, fm FieldMap) domain.Loan {
	rev := fm.Reverse()
	f := make(map[string]interface{}, len(raw))
	for fid, v := range raw {
		if name, ok := rev[fid]; ok {
			f[name] = v
		}
	}
	return domain.Loan{
		Type:             domain.LoanType(getString(f, "type")),
		CounterpartyName: getString(f, "counterparty_name"),
		CounterpartyURI:  getString(f, "counterparty_uri"),
		Description:      getString(f, "description"),
		OriginalAmount:   getAmount(f, "original_amount"),
		Currency:         getString(f, "currency"),
		DateCreated:      getTime(f, "date_created"),
		DueDate:          getTimePtr(f, "due_date"),
		InterestType:     domain.InterestType(getString(f, "interest_type")),
		InterestRate:     getAmount(f, "interest_rate"),
		InterestPeriod:   domain.InterestPeriod(getString(f, "interest_period")),
		IsSettled:        getBool(f, "is_settled"),
		Notes:            getString(f, "notes"),
	}
}

func mapLoanToFields(l domain.LoanInput, fm FieldMap) map[string]interface{} {
	fields := make(map[string]interface{})
	fields[fm["type"]] = string(l.Type)
	setIfNotZero(fields, fm["counterparty_name"], l.CounterpartyName)
	setIfNotZero(fields, fm["counterparty_uri"], l.CounterpartyURI)
	setIfNotZero(fields, fm["description"], l.Description)
	f, _ := l.OriginalAmount.Decimal.Float64()
	fields[fm["original_amount"]] = f
	setIfNotZero(fields, fm["currency"], l.Currency)
	if !l.DateCreated.IsZero() {
		fields[fm["date_created"]] = l.DateCreated.Format("2006-01-02")
	}
	if l.DueDate != nil {
		fields[fm["due_date"]] = l.DueDate.Format("2006-01-02")
	}
	setIfNotZero(fields, fm["interest_type"], string(l.InterestType))
	setIfNotZero(fields, fm["interest_rate"], l.InterestRate)
	setIfNotZero(fields, fm["interest_period"], string(l.InterestPeriod))
	if l.IsSettled {
		fields[fm["is_settled"]] = true
	}
	setIfNotZero(fields, fm["notes"], l.Notes)
	return fields
}

func mapFieldsToLoanPayment(raw map[string]interface{}, fm FieldMap) domain.LoanPayment {
	rev := fm.Reverse()
	f := make(map[string]interface{}, len(raw))
	for fid, v := range raw {
		if name, ok := rev[fid]; ok {
			f[name] = v
		}
	}
	return domain.LoanPayment{
		LoanID:    linkSingle(f["loan"]),
		Date:      getTime(f, "date"),
		Amount:    getAmount(f, "amount"),
		AccountID: linkSingle(f["account"]),
		Notes:     getString(f, "notes"),
	}
}

func mapLoanPaymentToFields(p domain.LoanPaymentInput, fm FieldMap) map[string]interface{} {
	fields := make(map[string]interface{})
	if p.LoanID != "" {
		fields[fm["loan"]] = wrapLink(p.LoanID)
	}
	if !p.Date.IsZero() {
		fields[fm["date"]] = p.Date.Format("2006-01-02")
	}
	f, _ := p.Amount.Decimal.Float64()
	fields[fm["amount"]] = f
	if p.AccountID != "" {
		fields[fm["account"]] = wrapLink(p.AccountID)
	}
	setIfNotZero(fields, fm["notes"], p.Notes)
	return fields
}

func mapFieldsToSnapshot(raw map[string]interface{}, fm FieldMap) domain.BalanceSnapshot {
	rev := fm.Reverse()
	f := make(map[string]interface{}, len(raw))
	for fid, v := range raw {
		if name, ok := rev[fid]; ok {
			f[name] = v
		}
	}
	return domain.BalanceSnapshot{
		AccountID: linkSingle(f["account"]),
		Date:      getTime(f, "date"),
		Balance:   getAmount(f, "balance"),
		Source:    domain.SnapshotSource(getString(f, "source")),
		Notes:     getString(f, "notes"),
	}
}

func mapSnapshotToFields(s domain.SnapshotInput, fm FieldMap) map[string]interface{} {
	fields := make(map[string]interface{})
	if s.AccountID != "" {
		fields[fm["account"]] = wrapLink(s.AccountID)
	}
	if !s.Date.IsZero() {
		fields[fm["date"]] = s.Date.Format("2006-01-02")
	}
	f, _ := s.Balance.Decimal.Float64()
	fields[fm["balance"]] = f
	setIfNotZero(fields, fm["source"], string(s.Source))
	setIfNotZero(fields, fm["notes"], s.Notes)
	return fields
}

func mapFieldsToRecurringRule(raw map[string]interface{}, fm FieldMap) domain.RecurringRule {
	rev := fm.Reverse()
	f := make(map[string]interface{}, len(raw))
	for fid, v := range raw {
		if name, ok := rev[fid]; ok {
			f[name] = v
		}
	}
	return domain.RecurringRule{
		Name:       getString(f, "name"),
		Type:       domain.TransactionType(getString(f, "type")),
		Amount:     getAmount(f, "amount"),
		CategoryID: linkSingle(f["category"]),
		TagIDs:     linkMulti(f["tags"]),
		AccountID:  linkSingle(f["account"]),
		Frequency:  domain.Frequency(getString(f, "frequency")),
		DayOfMonth: getInt(f, "day_of_month"),
		StartDate:  getTime(f, "start_date"),
		EndDate:    getTimePtr(f, "end_date"),
		IsActive:   getBool(f, "is_active"),
		Notes:      getString(f, "notes"),
	}
}

func mapRecurringRuleToFields(r domain.RecurringRuleInput, fm FieldMap) map[string]interface{} {
	fields := make(map[string]interface{})
	setIfNotZero(fields, fm["name"], r.Name)
	fields[fm["type"]] = string(r.Type)
	f, _ := r.Amount.Decimal.Float64()
	fields[fm["amount"]] = f
	if r.CategoryID != "" {
		fields[fm["category"]] = wrapLink(r.CategoryID)
	}
	if len(r.TagIDs) > 0 {
		fields[fm["tags"]] = wrapLinks(r.TagIDs)
	}
	if r.AccountID != "" {
		fields[fm["account"]] = wrapLink(r.AccountID)
	}
	fields[fm["frequency"]] = string(r.Frequency)
	if r.DayOfMonth > 0 {
		fields[fm["day_of_month"]] = r.DayOfMonth
	}
	if !r.StartDate.IsZero() {
		fields[fm["start_date"]] = r.StartDate.Format("2006-01-02")
	}
	if r.EndDate != nil {
		fields[fm["end_date"]] = r.EndDate.Format("2006-01-02")
	}
	fields[fm["is_active"]] = r.IsActive
	setIfNotZero(fields, fm["notes"], r.Notes)
	return fields
}

func mapFieldsToCategory(raw map[string]interface{}, fm FieldMap) domain.Category {
	rev := fm.Reverse()
	f := make(map[string]interface{}, len(raw))
	for fid, v := range raw {
		if name, ok := rev[fid]; ok {
			f[name] = v
		}
	}
	return domain.Category{
		Name: getString(f, "name"),
		Type: domain.CategoryType(getString(f, "type")),
	}
}

func mapFieldsToTag(raw map[string]interface{}, fm FieldMap) domain.Tag {
	rev := fm.Reverse()
	f := make(map[string]interface{}, len(raw))
	for fid, v := range raw {
		if name, ok := rev[fid]; ok {
			f[name] = v
		}
	}
	return domain.Tag{
		Name: getString(f, "name"),
	}
}
