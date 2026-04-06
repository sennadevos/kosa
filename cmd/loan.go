package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/app"
	"github.com/sennadevos/kosa/internal/domain"
	"github.com/sennadevos/kosa/internal/output"
)

var loanCmd = &cobra.Command{
	Use:   "loan",
	Short: "manage loans",
}

var loanNewCmd = &cobra.Command{
	Use:   "new <description> <amount>",
	Short: "create a new loan",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, err := domain.NewAmount(args[1])
		if err != nil {
			return fmt.Errorf("amount: %w", err)
		}

		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		dueStr, _ := cmd.Flags().GetString("due")
		interestStr, _ := cmd.Flags().GetString("interest")
		rateStr, _ := cmd.Flags().GetString("rate")
		periodStr, _ := cmd.Flags().GetString("period")
		notes, _ := cmd.Flags().GetString("notes")

		var loanType domain.LoanType
		var counterparty string
		if from != "" {
			loanType = domain.LoanPayable
			counterparty = from
		} else if to != "" {
			loanType = domain.LoanReceivable
			counterparty = to
		} else {
			return fmt.Errorf("specify --from (payable) or --to (receivable)")
		}

		var dueDate *time.Time
		if dueStr != "" {
			d, err := time.Parse("2006-01-02", dueStr)
			if err != nil {
				return fmt.Errorf("due date: %w", err)
			}
			dueDate = &d
		}

		var interestType domain.InterestType
		if interestStr != "" {
			interestType = domain.InterestType(interestStr)
		}

		var interestRate domain.Amount
		if rateStr != "" {
			interestRate, err = domain.NewAmount(rateStr)
			if err != nil {
				return fmt.Errorf("rate: %w", err)
			}
		}

		loan, err := application.LoanNew(cmd.Context(), app.LoanNewInput{
			Type:             loanType,
			CounterpartyName: counterparty,
			Description:      args[0],
			Amount:           amount,
			DueDate:          dueDate,
			InterestType:     interestType,
			InterestRate:     interestRate,
			InterestPeriod:   domain.InterestPeriod(periodStr),
			Notes:            notes,
		})
		if err != nil {
			return err
		}

		output.PrintLoanConfirmation(os.Stdout, loan)
		return nil
	},
}

var loanPayCmd = &cobra.Command{
	Use:   "pay <loan-id> <amount>",
	Short: "record a payment against a loan",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		amount, err := domain.NewAmount(args[1])
		if err != nil {
			return fmt.Errorf("amount: %w", err)
		}
		account, _ := cmd.Flags().GetString("account")
		notes, _ := cmd.Flags().GetString("notes")

		_, txn, err := application.LoanPay(cmd.Context(), app.LoanPayInput{
			LoanID:  args[0],
			Amount:  amount,
			Account: account,
			Notes:   notes,
		})
		if err != nil {
			return err
		}
		if txn != nil {
			output.PrintTransactionConfirmation(os.Stdout, txn)
		}
		return nil
	},
}

var loanShowCmd = &cobra.Command{
	Use:   "show <loan-id>",
	Short: "show loan details with payment timeline",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		loan, status, timeline, err := application.LoanShow(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		f := outputFormat()
		output.PrintLoanStatus(os.Stdout, loan, status, f)
		output.PrintLoanTimeline(os.Stdout, timeline, f)
		return nil
	},
}

func init() {
	loanNewCmd.Flags().String("from", "", "counterparty (payable — you owe them)")
	loanNewCmd.Flags().String("to", "", "counterparty (receivable — they owe you)")
	loanNewCmd.Flags().String("due", "", "due date (YYYY-MM-DD)")
	loanNewCmd.Flags().String("interest", "", "interest type: none/flat/periodic")
	loanNewCmd.Flags().String("rate", "", "interest rate percentage")
	loanNewCmd.Flags().String("period", "", "interest period: weekly/monthly/quarterly/yearly")
	loanNewCmd.Flags().String("notes", "", "notes")

	loanPayCmd.Flags().String("account", "", "account name")
	loanPayCmd.Flags().String("notes", "", "notes")

	loanCmd.AddCommand(loanNewCmd, loanPayCmd, loanShowCmd)
	rootCmd.AddCommand(loanCmd)
}
