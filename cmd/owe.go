package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/domain"
	"github.com/sennadevos/kosa/internal/output"
)

var oweCmd = &cobra.Command{
	Use:   "owe <amount> <description>",
	Short: "quick: you owe someone money",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, err := domain.NewAmount(args[0])
		if err != nil {
			return fmt.Errorf("amount: %w", err)
		}
		to, _ := cmd.Flags().GetString("to")
		if to == "" {
			return fmt.Errorf("--to is required")
		}

		loan, err := application.Owe(cmd.Context(), amount, args[1], to)
		if err != nil {
			return err
		}
		output.PrintLoanConfirmation(os.Stdout, loan)
		return nil
	},
}

func init() {
	oweCmd.Flags().String("to", "", "who you owe")
	rootCmd.AddCommand(oweCmd)
}
