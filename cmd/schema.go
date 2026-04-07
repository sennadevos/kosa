package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/backend/teable"
	"github.com/sennadevos/kosa/internal/config"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "manage database schema",
}

var schemaSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync schema: create missing fields in Teable",
	Long: `Compares the expected kosa schema against the actual fields in Teable
and creates any missing fields. This is additive only — it never deletes,
renames, or changes the type of existing fields.

Use this after updating kosa to a version that adds new fields. The command
reads your existing config.toml to find the table IDs, then checks each
table for missing fields and creates them.

After sync, new field IDs are appended to your config.toml automatically.

Limitations:
  - Only creates missing fields (additive)
  - Does not delete removed fields
  - Does not rename or change field types
  - Does not add new select choices to existing fields
  - For destructive changes, use the Teable UI and update config.toml manually`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		cfgPath := flagConfig
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}

		if cfg.Backend.Type != "teable" {
			return fmt.Errorf("schema sync only works with teable backend (got %q)", cfg.Backend.Type)
		}

		client := teable.NewClient(cfg.Backend.Teable.URL, cfg.Backend.Teable.Token)
		ctx := cmd.Context()

		if dryRun {
			fmt.Println("dry run — checking for missing fields...")
			fmt.Println()
		} else {
			fmt.Println("syncing schema...")
			fmt.Println()
		}

		if dryRun {
			// just list what would change
			printDryRun(ctx, client, cfg)
			return nil
		}

		result, err := teable.SchemaSync(ctx, client, cfg.Backend.Teable.Tables)
		if err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		created := 0
		updated := 0
		skipped := 0
		for _, a := range result.Actions {
			switch a.Action {
			case "create_field", "create_link":
				created++
				fmt.Printf("  + %s.%s — %s (id: %s)\n", a.Table, a.Field, a.Details, a.FieldID)
			case "update_constraints":
				updated++
				fmt.Printf("  ~ %s.%s — %s\n", a.Table, a.Field, a.Details)
			case "skip_ok":
				skipped++
			}
		}

		fmt.Printf("\n%d created, %d updated, %d up to date\n", created, updated, skipped)

		if created > 0 || updated > 0 {
			// update config with new field IDs
			for _, a := range result.Actions {
				if a.FieldID == "" {
					continue
				}
				if cfg.Backend.Teable.Fields[a.Table] == nil {
					cfg.Backend.Teable.Fields[a.Table] = make(map[string]string)
				}
				cfg.Backend.Teable.Fields[a.Table][a.Field] = a.FieldID
			}

			// resolve config path
			cfgFile := resolveConfigPath(cfgPath)
			f, err := os.Create(cfgFile)
			if err != nil {
				return fmt.Errorf("writing config: %w", err)
			}
			defer f.Close()

			fmt.Fprintln(f, "# kosa config — updated by kosa schema sync")
			fmt.Fprintln(f)
			if err := toml.NewEncoder(f).Encode(cfg); err != nil {
				return fmt.Errorf("encoding config: %w", err)
			}
			fmt.Printf("config updated: %s\n", cfgFile)
		}

		return nil
	},
}

func printDryRun(ctx context.Context, client *teable.Client, cfg *config.Config) {
	schema := teable.ExpectedSchema()

	for tableKey, tableDef := range schema {
		tableID, ok := cfg.Backend.Teable.Tables[tableKey]
		if !ok || tableID == "" {
			fmt.Printf("  %s: table not in config (skip)\n", tableKey)
			continue
		}

		existing, err := client.ListFields(ctx, tableID)
		if err != nil {
			fmt.Printf("  %s: error listing fields: %v\n", tableKey, err)
			continue
		}

		existingByName := make(map[string]teable.FieldResult, len(existing))
		for _, f := range existing {
			existingByName[f.Name] = f
		}

		changes := 0
		for _, f := range tableDef.Fields {
			if ef, exists := existingByName[f.Name]; exists {
				// check constraint mismatches
				if f.NotNull && !ef.NotNull {
					fmt.Printf("  ~ %s.%s: notNull false→true\n", tableKey, f.Name)
					changes++
				}
				if f.Unique && !ef.Unique {
					fmt.Printf("  ~ %s.%s: unique false→true\n", tableKey, f.Name)
					changes++
				}
			} else {
				fmt.Printf("  + %s.%s (type=%s)\n", tableKey, f.Name, f.Type)
				changes++
			}
		}
		for _, l := range tableDef.LinkFields {
			if _, exists := existingByName[l.Name]; !exists {
				fmt.Printf("  + %s.%s (link -> %s, %s)\n", tableKey, l.Name, l.ForeignTable, l.Relationship)
				changes++
			}
		}

		if changes == 0 {
			fmt.Printf("  %s: up to date\n", tableKey)
		}
	}
}

func resolveConfigPath(path string) string {
	if path != "" {
		return path
	}
	if p := os.Getenv("KOSA_CONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return home + "/.config/kosa/config.toml"
}

var schemaStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "show schema status: what fields exist vs expected",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(flagConfig)
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}

		if cfg.Backend.Type != "teable" {
			return fmt.Errorf("schema status only works with teable backend")
		}

		client := teable.NewClient(cfg.Backend.Teable.URL, cfg.Backend.Teable.Token)
		schema := teable.ExpectedSchema()
		ctx := cmd.Context()

		for tableKey, tableDef := range schema {
			tableID, ok := cfg.Backend.Teable.Tables[tableKey]
			if !ok || tableID == "" {
				fmt.Printf("%s: not configured\n", tableKey)
				continue
			}

			existing, err := client.ListFields(ctx, tableID)
			if err != nil {
				fmt.Printf("%s: error — %v\n", tableKey, err)
				continue
			}

			existingByName := make(map[string]teable.FieldResult, len(existing))
			for _, f := range existing {
				existingByName[f.Name] = f
			}

			totalExpected := len(tableDef.Fields) + len(tableDef.LinkFields)
			present := 0
			var issues []string

			for _, f := range tableDef.Fields {
				if ef, exists := existingByName[f.Name]; exists {
					present++
					if f.NotNull && !ef.NotNull {
						issues = append(issues, f.Name+" (notNull missing)")
					}
					if f.Unique && !ef.Unique {
						issues = append(issues, f.Name+" (unique missing)")
					}
				} else {
					issues = append(issues, f.Name+" (missing)")
				}
			}
			for _, l := range tableDef.LinkFields {
				if _, exists := existingByName[l.Name]; exists {
					present++
				} else {
					issues = append(issues, l.Name+" (missing)")
				}
			}

			if len(issues) == 0 {
				fmt.Printf("%s: ok (%d/%d fields)\n", tableKey, present, totalExpected)
			} else {
				fmt.Printf("%s: %d/%d fields, issues: %s\n",
					tableKey, present, totalExpected, strings.Join(issues, ", "))
			}
		}

		return nil
	},
}

func init() {
	schemaSyncCmd.Flags().Bool("dry-run", false, "show what would change without applying")
	schemaCmd.AddCommand(schemaSyncCmd, schemaStatusCmd)
	rootCmd.AddCommand(schemaCmd)
}
