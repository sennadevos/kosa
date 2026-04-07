package teable

import (
	"context"
	"fmt"
)

// SetupResult holds all table and field IDs created during setup.
type SetupResult struct {
	Tables map[string]string            // table name -> table ID
	Fields map[string]map[string]string // table name -> field name -> field ID
}

// Setup creates all kosa tables and fields in a Teable base.
// Tables are created in dependency order so link fields can reference them.
func Setup(ctx context.Context, client *Client, baseID string) (*SetupResult, error) {
	result := &SetupResult{
		Tables: make(map[string]string),
		Fields: make(map[string]map[string]string),
	}

	// Phase 1: create tables without link fields (order doesn't matter)
	if err := createCategories(ctx, client, baseID, result); err != nil {
		return nil, fmt.Errorf("categories: %w", err)
	}
	if err := createTags(ctx, client, baseID, result); err != nil {
		return nil, fmt.Errorf("tags: %w", err)
	}
	if err := createAccounts(ctx, client, baseID, result); err != nil {
		return nil, fmt.Errorf("accounts: %w", err)
	}

	// Phase 2: create tables that link to phase 1 tables
	if err := createTransactions(ctx, client, baseID, result); err != nil {
		return nil, fmt.Errorf("transactions: %w", err)
	}
	if err := createRecurringRules(ctx, client, baseID, result); err != nil {
		return nil, fmt.Errorf("recurring_rules: %w", err)
	}
	if err := createLoans(ctx, client, baseID, result); err != nil {
		return nil, fmt.Errorf("loans: %w", err)
	}
	if err := createBalanceSnapshots(ctx, client, baseID, result); err != nil {
		return nil, fmt.Errorf("balance_snapshots: %w", err)
	}

	// Phase 3: tables that link to phase 2 tables
	if err := createLoanPayments(ctx, client, baseID, result); err != nil {
		return nil, fmt.Errorf("loan_payments: %w", err)
	}

	// Phase 4: add link fields that reference tables created later (e.g., transaction -> loan, transaction -> recurring_rule, refund_of -> transaction)
	if err := addTransactionLinks(ctx, client, result); err != nil {
		return nil, fmt.Errorf("transaction links: %w", err)
	}

	return result, nil
}

func saveField(result *SetupResult, table, name, id string) {
	if result.Fields[table] == nil {
		result.Fields[table] = make(map[string]string)
	}
	result.Fields[table][name] = id
}

func createCategories(ctx context.Context, client *Client, baseID string, result *SetupResult) error {
	resp, err := client.CreateTable(ctx, baseID, CreateTableRequest{
		Name: "Categories",
		Fields: []CreateFieldRequest{
			{Name: "name", Type: "singleLineText"},
			{Name: "type", Type: "singleSelect", Options: SelectFieldOptions{
				Choices: []SelectChoice{
					{Name: "income"}, {Name: "expense"}, {Name: "neutral"},
				},
			}},
		},
	})
	if err != nil {
		return err
	}
	result.Tables["categories"] = resp.ID
	for _, f := range resp.Fields {
		saveField(result, "categories", f.Name, f.ID)
	}
	return nil
}

func createTags(ctx context.Context, client *Client, baseID string, result *SetupResult) error {
	resp, err := client.CreateTable(ctx, baseID, CreateTableRequest{
		Name: "Tags",
		Fields: []CreateFieldRequest{
			{Name: "name", Type: "singleLineText"},
		},
	})
	if err != nil {
		return err
	}
	result.Tables["tags"] = resp.ID
	for _, f := range resp.Fields {
		saveField(result, "tags", f.Name, f.ID)
	}
	return nil
}

func createAccounts(ctx context.Context, client *Client, baseID string, result *SetupResult) error {
	resp, err := client.CreateTable(ctx, baseID, CreateTableRequest{
		Name: "Accounts",
		Fields: []CreateFieldRequest{
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
			{Name: "is_default", Type: "checkbox"},
			{Name: "notes", Type: "longText"},
		},
	})
	if err != nil {
		return err
	}
	result.Tables["accounts"] = resp.ID
	for _, f := range resp.Fields {
		saveField(result, "accounts", f.Name, f.ID)
	}
	return nil
}

func createTransactions(ctx context.Context, client *Client, baseID string, result *SetupResult) error {
	// create table with non-link fields first
	resp, err := client.CreateTable(ctx, baseID, CreateTableRequest{
		Name: "Transactions",
		Fields: []CreateFieldRequest{
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
	})
	if err != nil {
		return err
	}
	result.Tables["transactions"] = resp.ID
	for _, f := range resp.Fields {
		saveField(result, "transactions", f.Name, f.ID)
	}

	// add link fields to already-created tables
	for _, link := range []struct {
		name         string
		foreignTable string
		relationship string
	}{
		{"category", "categories", "manyOne"},
		{"tags", "tags", "manyMany"},
		{"account", "accounts", "manyOne"},
		{"to_account", "accounts", "manyOne"},
	} {
		f, err := client.CreateField(ctx, resp.ID, CreateFieldRequest{
			Name: link.name,
			Type: "link",
			Options: LinkFieldOptions{
				Relationship:   link.relationship,
				ForeignTableID: result.Tables[link.foreignTable],
				IsOneWay:       true,
			},
		})
		if err != nil {
			return fmt.Errorf("link field %s: %w", link.name, err)
		}
		saveField(result, "transactions", link.name, f.ID)
	}

	return nil
}

func createRecurringRules(ctx context.Context, client *Client, baseID string, result *SetupResult) error {
	resp, err := client.CreateTable(ctx, baseID, CreateTableRequest{
		Name: "Recurring Rules",
		Fields: []CreateFieldRequest{
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
	})
	if err != nil {
		return err
	}
	result.Tables["recurring_rules"] = resp.ID
	for _, f := range resp.Fields {
		saveField(result, "recurring_rules", f.Name, f.ID)
	}

	// link fields
	for _, link := range []struct {
		name         string
		foreignTable string
		relationship string
	}{
		{"category", "categories", "manyOne"},
		{"tags", "tags", "manyMany"},
		{"account", "accounts", "manyOne"},
	} {
		f, err := client.CreateField(ctx, resp.ID, CreateFieldRequest{
			Name: link.name,
			Type: "link",
			Options: LinkFieldOptions{
				Relationship:   link.relationship,
				ForeignTableID: result.Tables[link.foreignTable],
				IsOneWay:       true,
			},
		})
		if err != nil {
			return fmt.Errorf("link field %s: %w", link.name, err)
		}
		saveField(result, "recurring_rules", link.name, f.ID)
	}

	return nil
}

func createLoans(ctx context.Context, client *Client, baseID string, result *SetupResult) error {
	resp, err := client.CreateTable(ctx, baseID, CreateTableRequest{
		Name: "Loans",
		Fields: []CreateFieldRequest{
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
	})
	if err != nil {
		return err
	}
	result.Tables["loans"] = resp.ID
	for _, f := range resp.Fields {
		saveField(result, "loans", f.Name, f.ID)
	}
	return nil
}

func createBalanceSnapshots(ctx context.Context, client *Client, baseID string, result *SetupResult) error {
	resp, err := client.CreateTable(ctx, baseID, CreateTableRequest{
		Name: "Balance Snapshots",
		Fields: []CreateFieldRequest{
			{Name: "date", Type: "date"},
			{Name: "balance", Type: "number"},
			{Name: "source", Type: "singleSelect", Options: SelectFieldOptions{
				Choices: []SelectChoice{
					{Name: "manual"}, {Name: "bank_import"}, {Name: "reconciliation"},
				},
			}},
			{Name: "notes", Type: "longText"},
		},
	})
	if err != nil {
		return err
	}
	result.Tables["balance_snapshots"] = resp.ID
	for _, f := range resp.Fields {
		saveField(result, "balance_snapshots", f.Name, f.ID)
	}

	// link to accounts
	f, err := client.CreateField(ctx, resp.ID, CreateFieldRequest{
		Name: "account",
		Type: "link",
		Options: LinkFieldOptions{
			Relationship:   "manyOne",
			ForeignTableID: result.Tables["accounts"],
			IsOneWay:       true,
		},
	})
	if err != nil {
		return fmt.Errorf("link field account: %w", err)
	}
	saveField(result, "balance_snapshots", "account", f.ID)
	return nil
}

func createLoanPayments(ctx context.Context, client *Client, baseID string, result *SetupResult) error {
	resp, err := client.CreateTable(ctx, baseID, CreateTableRequest{
		Name: "Loan Payments",
		Fields: []CreateFieldRequest{
			{Name: "date", Type: "date"},
			{Name: "amount", Type: "number"},
			{Name: "notes", Type: "longText"},
		},
	})
	if err != nil {
		return err
	}
	result.Tables["loan_payments"] = resp.ID
	for _, f := range resp.Fields {
		saveField(result, "loan_payments", f.Name, f.ID)
	}

	// link fields
	for _, link := range []struct {
		name         string
		foreignTable string
	}{
		{"loan", "loans"},
		{"account", "accounts"},
	} {
		f, err := client.CreateField(ctx, resp.ID, CreateFieldRequest{
			Name: link.name,
			Type: "link",
			Options: LinkFieldOptions{
				Relationship:   "manyOne",
				ForeignTableID: result.Tables[link.foreignTable],
				IsOneWay:       true,
			},
		})
		if err != nil {
			return fmt.Errorf("link field %s: %w", link.name, err)
		}
		saveField(result, "loan_payments", link.name, f.ID)
	}

	return nil
}

// addTransactionLinks adds link fields to Transactions that reference
// tables created after it (loans, recurring_rules, and self-reference for refund_of).
func addTransactionLinks(ctx context.Context, client *Client, result *SetupResult) error {
	txnTable := result.Tables["transactions"]

	for _, link := range []struct {
		name         string
		foreignTable string
		relationship string
	}{
		{"loan", "loans", "manyOne"},
		{"recurring_rule", "recurring_rules", "manyOne"},
		{"refund_of", "transactions", "manyOne"},
	} {
		f, err := client.CreateField(ctx, txnTable, CreateFieldRequest{
			Name: link.name,
			Type: "link",
			Options: LinkFieldOptions{
				Relationship:   link.relationship,
				ForeignTableID: result.Tables[link.foreignTable],
				IsOneWay:       true,
			},
		})
		if err != nil {
			return fmt.Errorf("link field %s: %w", link.name, err)
		}
		saveField(result, "transactions", link.name, f.ID)
	}

	return nil
}
