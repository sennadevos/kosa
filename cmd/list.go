package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/app"
	"github.com/sennadevos/kosa/internal/output"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cat, _ := cmd.Flags().GetString("cat")
		tag, _ := cmd.Flags().GetString("tag")
		account, _ := cmd.Flags().GetString("account")
		txnType, _ := cmd.Flags().GetString("type")

		if limit == 0 {
			limit = 20
		}

		txns, err := application.ListTransactions(cmd.Context(), app.ListInput{
			Limit:    limit,
			Category: cat,
			Tag:      tag,
			Account:  account,
			Type:     txnType,
		})
		if err != nil {
			return err
		}

		output.PrintTransactions(os.Stdout, txns, outputFormat())
		return nil
	},
}

func init() {
	listCmd.Flags().Int("limit", 20, "number of transactions to show")
	listCmd.Flags().String("cat", "", "filter by category")
	listCmd.Flags().String("tag", "", "filter by tag")
	listCmd.Flags().String("account", "", "filter by account")
	listCmd.Flags().String("type", "", "filter by type (income/expense/transfer/refund)")
	rootCmd.AddCommand(listCmd)
}
