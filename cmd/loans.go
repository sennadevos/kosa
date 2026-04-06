package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/domain"
	"github.com/sennadevos/kosa/internal/output"
)

var loansCmd = &cobra.Command{
	Use:   "loans",
	Short: "list loans",
	RunE: func(cmd *cobra.Command, args []string) error {
		payable, _ := cmd.Flags().GetBool("payable")
		receivable, _ := cmd.Flags().GetBool("receivable")

		filter := domain.LoanFilter{}
		unsettled := false
		filter.Settled = &unsettled

		if payable {
			t := domain.LoanPayable
			filter.Type = &t
		}
		if receivable {
			t := domain.LoanReceivable
			filter.Type = &t
		}

		loans, err := application.Backend.ListLoans(cmd.Context(), filter)
		if err != nil {
			return err
		}

		output.PrintLoans(os.Stdout, loans, outputFormat())
		return nil
	},
}

func init() {
	loansCmd.Flags().Bool("payable", false, "show only payable loans")
	loansCmd.Flags().Bool("receivable", false, "show only receivable loans")
	rootCmd.AddCommand(loansCmd)
}
