package teable

// schema_def.go defines the expected schema for all kosa tables.
//
// This is the single source of truth for what the Teable schema should
// look like. Both `kosa init` (fresh setup) and `kosa schema sync`
// (incremental migration) use this definition.
//
// To evolve the schema:
//   1. Add/modify fields in the table definitions below
//   2. Run `kosa schema sync` to apply changes to an existing Teable base
//   3. The sync command will create missing fields and report what changed
//
// Limitations:
//   - Sync is additive only: it creates missing fields but never deletes,
//     renames, or changes the type of existing fields
//   - Destructive changes (rename, delete, type change) must be done
//     manually in the Teable UI, then update the config.toml field IDs
//   - Select choice additions are not yet synced — add them in the UI

// FieldDef describes a field that should exist on a table.
// Required-field enforcement is handled at the application layer (CLI
// validation), not in Teable — its API does not support notNull constraints.
type FieldDef struct {
	Name    string
	Type    string
	Options interface{} // nil, SelectFieldOptions, or LinkFieldOptions
}

// TableDef describes a table and its expected fields.
type TableDef struct {
	Name   string
	Fields []FieldDef
	// LinkFields are created after all tables exist, since they reference
	// other tables by ID. They are listed separately so that init can
	// create them in the right order.
	LinkFields []LinkFieldDef
}

// LinkFieldDef describes a link field that references another table.
type LinkFieldDef struct {
	Name           string
	ForeignTable   string // key in the tables map (e.g. "accounts")
	Relationship   string // "manyOne", "manyMany", "oneMany"
}

// ExpectedSchema returns the complete schema definition for all kosa tables.
// This is used by both `kosa init` and `kosa schema sync`.
func ExpectedSchema() map[string]TableDef {
	return map[string]TableDef{
		"categories": {
			Name: "Categories",
			Fields: []FieldDef{
				{Name: "name", Type: "singleLineText"},
				{Name: "type", Type: "singleSelect", Options: SelectFieldOptions{
					Choices: []SelectChoice{
						{Name: "income"}, {Name: "expense"}, {Name: "neutral"},
					},
				}},
			},
		},
		"tags": {
			Name: "Tags",
			Fields: []FieldDef{
				{Name: "name", Type: "singleLineText"},
			},
		},
		"accounts": {
			Name: "Accounts",
			Fields: []FieldDef{
				{Name: "name", Type: "singleLineText"},
				{Name: "type", Type: "singleSelect", Options: SelectFieldOptions{
					Choices: []SelectChoice{
						{Name: "checking"}, {Name: "savings"}, {Name: "investment"},
						{Name: "credit_card"}, {Name: "cash"},
					},
				}},
				{Name: "provider", Type: "singleLineText"},
				{Name: "currency", Type: "singleLineText"},
				{Name: "iban", Type: "singleLineText"},
				{Name: "notes", Type: "longText"},
			},
		},
		"transactions": {
			Name: "Transactions",
			Fields: []FieldDef{
				{Name: "date", Type: "date"},
				{Name: "type", Type: "singleSelect", Options: SelectFieldOptions{
					Choices: []SelectChoice{
						{Name: "income"}, {Name: "expense"}, {Name: "transfer"}, {Name: "refund"},
					},
				}},
				{Name: "amount", Type: "number"},
				{Name: "description", Type: "singleLineText"},
				{Name: "cashback", Type: "number"},
				{Name: "reference", Type: "singleLineText"},
				{Name: "foreign_amount", Type: "number"},
				{Name: "foreign_currency", Type: "singleLineText"},
				{Name: "exchange_rate", Type: "number"},
				{Name: "notes", Type: "longText"},
			},
			LinkFields: []LinkFieldDef{
				{Name: "category", ForeignTable: "categories", Relationship: "manyOne"},
				{Name: "tags", ForeignTable: "tags", Relationship: "manyMany"},
				{Name: "account", ForeignTable: "accounts", Relationship: "manyOne"},
				{Name: "to_account", ForeignTable: "accounts", Relationship: "manyOne"},
				{Name: "loan", ForeignTable: "loans", Relationship: "manyOne"},
				{Name: "recurring_rule", ForeignTable: "recurring_rules", Relationship: "manyOne"},
				{Name: "refund_of", ForeignTable: "transactions", Relationship: "manyOne"},
			},
		},
		"recurring_rules": {
			Name: "Recurring Rules",
			Fields: []FieldDef{
				{Name: "name", Type: "singleLineText"},
				{Name: "type", Type: "singleSelect", Options: SelectFieldOptions{
					Choices: []SelectChoice{{Name: "income"}, {Name: "expense"}},
				}},
				{Name: "amount", Type: "number"},
				{Name: "frequency", Type: "singleSelect", Options: SelectFieldOptions{
					Choices: []SelectChoice{
						{Name: "daily"}, {Name: "weekly"}, {Name: "biweekly"},
						{Name: "monthly"}, {Name: "quarterly"}, {Name: "yearly"},
					},
				}},
				{Name: "day_of_month", Type: "number"},
				{Name: "start_date", Type: "date"},
				{Name: "end_date", Type: "date"},
				{Name: "is_active", Type: "checkbox"},
				{Name: "notes", Type: "longText"},
			},
			LinkFields: []LinkFieldDef{
				{Name: "category", ForeignTable: "categories", Relationship: "manyOne"},
				{Name: "tags", ForeignTable: "tags", Relationship: "manyMany"},
				{Name: "account", ForeignTable: "accounts", Relationship: "manyOne"},
			},
		},
		"loans": {
			Name: "Loans",
			Fields: []FieldDef{
				{Name: "type", Type: "singleSelect", Options: SelectFieldOptions{
					Choices: []SelectChoice{{Name: "payable"}, {Name: "receivable"}},
				}},
				{Name: "counterparty_name", Type: "singleLineText"},
				{Name: "counterparty_uri", Type: "singleLineText"},
				{Name: "description", Type: "singleLineText"},
				{Name: "original_amount", Type: "number"},
				{Name: "currency", Type: "singleLineText"},
				{Name: "date_created", Type: "date"},
				{Name: "due_date", Type: "date"},
				{Name: "interest_type", Type: "singleSelect", Options: SelectFieldOptions{
					Choices: []SelectChoice{{Name: "none"}, {Name: "flat"}, {Name: "periodic"}},
				}},
				{Name: "interest_rate", Type: "number"},
				{Name: "interest_period", Type: "singleSelect", Options: SelectFieldOptions{
					Choices: []SelectChoice{
						{Name: "weekly"}, {Name: "monthly"}, {Name: "quarterly"}, {Name: "yearly"},
					},
				}},
				{Name: "is_settled", Type: "checkbox"},
				{Name: "notes", Type: "longText"},
			},
		},
		"loan_payments": {
			Name: "Loan Payments",
			Fields: []FieldDef{
				{Name: "date", Type: "date"},
				{Name: "amount", Type: "number"},
				{Name: "notes", Type: "longText"},
			},
			LinkFields: []LinkFieldDef{
				{Name: "loan", ForeignTable: "loans", Relationship: "manyOne"},
				{Name: "account", ForeignTable: "accounts", Relationship: "manyOne"},
			},
		},
		"balance_snapshots": {
			Name: "Balance Snapshots",
			Fields: []FieldDef{
				{Name: "date", Type: "date"},
				{Name: "balance", Type: "number"},
				{Name: "source", Type: "singleSelect", Options: SelectFieldOptions{
					Choices: []SelectChoice{
						{Name: "manual"}, {Name: "bank_import"}, {Name: "reconciliation"},
					},
				}},
				{Name: "notes", Type: "longText"},
			},
			LinkFields: []LinkFieldDef{
				{Name: "account", ForeignTable: "accounts", Relationship: "manyOne"},
			},
		},
	}
}
