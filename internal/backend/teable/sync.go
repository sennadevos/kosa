package teable

import (
	"context"
	"fmt"
)

// SyncAction describes a single change to apply or report.
type SyncAction struct {
	Table   string
	Field   string
	Action  string // "create_field", "create_link", "skip_exists"
	Details string
	FieldID string // populated after applying
}

// SyncResult holds the outcome of a schema sync.
type SyncResult struct {
	Actions []SyncAction
}

// SchemaSync compares the expected schema (from schema_def.go) against the
// actual fields in Teable, and creates any missing fields. It is additive
// only — it never deletes, renames, or changes the type of existing fields.
//
// It uses the table IDs from the existing config to find the tables. If a
// table doesn't exist in config, it is skipped (use `kosa init` for that).
//
// After sync, the caller should update config.toml with any new field IDs
// returned in the SyncResult.
func SchemaSync(ctx context.Context, client *Client, tables map[string]string) (*SyncResult, error) {
	expected := ExpectedSchema()
	result := &SyncResult{}

	for tableKey, tableDef := range expected {
		tableID, ok := tables[tableKey]
		if !ok || tableID == "" {
			result.Actions = append(result.Actions, SyncAction{
				Table:   tableKey,
				Action:  "skip_exists",
				Details: "table not in config — run kosa init first",
			})
			continue
		}

		// get existing fields
		existing, err := client.ListFields(ctx, tableID)
		if err != nil {
			return nil, fmt.Errorf("listing fields for %s: %w", tableKey, err)
		}

		existingByName := make(map[string]FieldResult, len(existing))
		for _, f := range existing {
			existingByName[f.Name] = f
		}

		// check regular fields
		for _, fieldDef := range tableDef.Fields {
			if _, exists := existingByName[fieldDef.Name]; exists {
				result.Actions = append(result.Actions, SyncAction{
					Table:   tableKey,
					Field:   fieldDef.Name,
					Action:  "skip_exists",
					Details: "already exists",
				})
				continue
			}

			// create missing field
			created, err := client.CreateField(ctx, tableID, StandaloneFieldRequest{
				Name:    fieldDef.Name,
				Type:    fieldDef.Type,
				Options: fieldDef.Options,
			})
			if err != nil {
				return nil, fmt.Errorf("creating field %s.%s: %w", tableKey, fieldDef.Name, err)
			}

			// set notNull if required
			if fieldDef.NotNull {
				notNull := true
				if err := client.UpdateField(ctx, tableID, created.ID, UpdateFieldRequest{
					NotNull: &notNull,
				}); err != nil {
					return nil, fmt.Errorf("setting notNull on %s.%s: %w", tableKey, fieldDef.Name, err)
				}
			}

			result.Actions = append(result.Actions, SyncAction{
				Table:   tableKey,
				Field:   fieldDef.Name,
				Action:  "create_field",
				Details: fmt.Sprintf("type=%s required=%v", fieldDef.Type, fieldDef.NotNull),
				FieldID: created.ID,
			})
		}

		// check link fields
		for _, linkDef := range tableDef.LinkFields {
			if _, exists := existingByName[linkDef.Name]; exists {
				result.Actions = append(result.Actions, SyncAction{
					Table:   tableKey,
					Field:   linkDef.Name,
					Action:  "skip_exists",
					Details: "already exists",
				})
				continue
			}

			foreignTableID, ok := tables[linkDef.ForeignTable]
			if !ok || foreignTableID == "" {
				result.Actions = append(result.Actions, SyncAction{
					Table:   tableKey,
					Field:   linkDef.Name,
					Action:  "skip_exists",
					Details: fmt.Sprintf("foreign table %q not in config", linkDef.ForeignTable),
				})
				continue
			}

			created, err := client.CreateField(ctx, tableID, StandaloneFieldRequest{
				Name: linkDef.Name,
				Type: "link",
				Options: LinkFieldOptions{
					Relationship:   linkDef.Relationship,
					ForeignTableID: foreignTableID,
					IsOneWay:       true,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("creating link %s.%s: %w", tableKey, linkDef.Name, err)
			}

			result.Actions = append(result.Actions, SyncAction{
				Table:   tableKey,
				Field:   linkDef.Name,
				Action:  "create_link",
				Details: fmt.Sprintf("-> %s (%s)", linkDef.ForeignTable, linkDef.Relationship),
				FieldID: created.ID,
			})
		}
	}

	return result, nil
}
