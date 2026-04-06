package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/output"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "search transactions by description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		if limit == 0 {
			limit = 20
		}

		txns, err := application.SearchTransactions(cmd.Context(), args[0], limit)
		if err != nil {
			return err
		}

		output.PrintTransactions(os.Stdout, txns, outputFormat())
		return nil
	},
}

func init() {
	searchCmd.Flags().Int("limit", 20, "max results")
	rootCmd.AddCommand(searchCmd)
}
