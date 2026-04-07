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
// It reads the expected schema from ExpectedSchema() and creates tables
// in dependency order so link fields can reference each other.
//
// Fields are created without notNull (Teable doesn't support it during
// inline table creation), then required fields are updated via PATCH
// to set notNull afterward.
func Setup(ctx context.Context, client *Client, baseID string) (*SetupResult, error) {
	schema := ExpectedSchema()
	result := &SetupResult{
		Tables: make(map[string]string),
		Fields: make(map[string]map[string]string),
	}

	// Create tables in dependency order:
	// Phase 1: tables with no link field dependencies
	// Phase 2: tables that link to phase 1 tables
	// Phase 3: tables that link to phase 2 tables
	// Phase 4: add remaining link fields (cross-references between phase 2+ tables)
	// Phase 5: set notNull on required fields
	phases := [][]string{
		{"categories", "tags", "accounts"},           // no dependencies
		{"transactions", "recurring_rules", "loans"}, // link to phase 1
		{"loan_payments", "balance_snapshots"},        // link to phase 2
	}

	for _, phase := range phases {
		for _, tableKey := range phase {
			def, ok := schema[tableKey]
			if !ok {
				continue
			}
			if err := createTableFromDef(ctx, client, baseID, tableKey, def, result); err != nil {
				return nil, fmt.Errorf("%s: %w", tableKey, err)
			}
		}
	}

	// Phase 4: create link fields that couldn't be created during table creation
	for _, phase := range phases {
		for _, tableKey := range phase {
			def := schema[tableKey]
			if err := createMissingLinks(ctx, client, tableKey, def, result); err != nil {
				return nil, fmt.Errorf("%s links: %w", tableKey, err)
			}
		}
	}

	// Phase 5: set notNull on required fields via PATCH
	for _, phase := range phases {
		for _, tableKey := range phase {
			def := schema[tableKey]
			if err := setRequiredFields(ctx, client, tableKey, def, result); err != nil {
				return nil, fmt.Errorf("%s required: %w", tableKey, err)
			}
		}
	}

	return result, nil
}

func saveField(result *SetupResult, table, name, id string) {
	if result.Fields[table] == nil {
		result.Fields[table] = make(map[string]string)
	}
	result.Fields[table][name] = id
}

func createTableFromDef(ctx context.Context, client *Client, baseID, tableKey string, def TableDef, result *SetupResult) error {
	// convert FieldDefs to CreateFieldRequests (non-link fields, no notNull)
	var fields []CreateFieldRequest
	for _, f := range def.Fields {
		fields = append(fields, CreateFieldRequest{
			Name:    f.Name,
			Type:    f.Type,
			Options: f.Options,
		})
	}

	resp, err := client.CreateTable(ctx, baseID, CreateTableRequest{
		Name:   def.Name,
		Fields: fields,
	})
	if err != nil {
		return err
	}

	result.Tables[tableKey] = resp.ID
	for _, f := range resp.Fields {
		saveField(result, tableKey, f.Name, f.ID)
	}

	// delete default empty rows that Teable creates with every new table
	defaultRecords, _ := client.ListRecords(ctx, resp.ID, "", "", 0)
	for _, r := range defaultRecords {
		client.DeleteRecord(ctx, resp.ID, r.ID)
	}

	// create link fields where the target table already exists
	for _, link := range def.LinkFields {
		foreignID, ok := result.Tables[link.ForeignTable]
		if !ok || foreignID == "" {
			continue // target table not created yet — handled in phase 4
		}

		f, err := client.CreateField(ctx, resp.ID, StandaloneFieldRequest{
			Name: link.Name,
			Type: "link",
			Options: LinkFieldOptions{
				Relationship:   link.Relationship,
				ForeignTableID: foreignID,
				IsOneWay:       true,
			},
		})
		if err != nil {
			return fmt.Errorf("link field %s: %w", link.Name, err)
		}
		saveField(result, tableKey, link.Name, f.ID)
	}

	return nil
}

func createMissingLinks(ctx context.Context, client *Client, tableKey string, def TableDef, result *SetupResult) error {
	tableID := result.Tables[tableKey]
	for _, link := range def.LinkFields {
		if _, exists := result.Fields[tableKey][link.Name]; exists {
			continue
		}

		foreignID, ok := result.Tables[link.ForeignTable]
		if !ok || foreignID == "" {
			return fmt.Errorf("foreign table %q not found", link.ForeignTable)
		}

		f, err := client.CreateField(ctx, tableID, StandaloneFieldRequest{
			Name: link.Name,
			Type: "link",
			Options: LinkFieldOptions{
				Relationship:   link.Relationship,
				ForeignTableID: foreignID,
				IsOneWay:       true,
			},
		})
		if err != nil {
			return fmt.Errorf("link field %s: %w", link.Name, err)
		}
		saveField(result, tableKey, link.Name, f.ID)
	}
	return nil
}

// setRequiredFields uses PUT /convert to set notNull: true on required fields.
func setRequiredFields(ctx context.Context, client *Client, tableKey string, def TableDef, result *SetupResult) error {
	tableID := result.Tables[tableKey]
	notNull := true

	for _, f := range def.Fields {
		if !f.NotNull {
			continue
		}
		fieldID, ok := result.Fields[tableKey][f.Name]
		if !ok {
			continue
		}
		if err := client.ConvertField(ctx, tableID, fieldID, ConvertFieldRequest{
			Type:    f.Type,
			NotNull: &notNull,
			Options: f.Options,
		}); err != nil {
			return fmt.Errorf("setting notNull on %s: %w", f.Name, err)
		}
	}
	return nil
}
