package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/app"
	"github.com/sennadevos/kosa/internal/domain"
	"github.com/sennadevos/kosa/internal/output"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot <balance>",
	Short: "record a balance snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		balance, err := domain.NewAmount(args[0])
		if err != nil {
			return fmt.Errorf("balance: %w", err)
		}

		account, _ := cmd.Flags().GetString("account")
		source, _ := cmd.Flags().GetString("source")
		notes, _ := cmd.Flags().GetString("notes")

		snap, err := application.RecordSnapshot(cmd.Context(), app.SnapshotInput{
			Balance: balance,
			Account: account,
			Source:  domain.SnapshotSource(source),
			Notes:   notes,
		})
		if err != nil {
			return err
		}

		output.PrintSnapshotConfirmation(os.Stdout, snap)
		return nil
	},
}

func init() {
	snapshotCmd.Flags().String("account", "", "account name")
	snapshotCmd.Flags().String("source", "manual", "source: manual/bank_import/reconciliation")
	snapshotCmd.Flags().String("notes", "", "notes")
	rootCmd.AddCommand(snapshotCmd)
}
