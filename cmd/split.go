package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/app"
	"github.com/sennadevos/kosa/internal/domain"
	"github.com/sennadevos/kosa/internal/output"
)

var splitCmd = &cobra.Command{
	Use:   "split <amount> <description>",
	Short: "split a cost with friends",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		total, err := domain.NewAmount(args[0])
		if err != nil {
			return fmt.Errorf("amount: %w", err)
		}

		cat, _ := cmd.Flags().GetString("cat")
		account, _ := cmd.Flags().GetString("account")
		withStr, _ := cmd.Flags().GetString("with")
		mineStr, _ := cmd.Flags().GetString("mine")

		if withStr == "" {
			return fmt.Errorf("--with is required")
		}

		// parse --with: "Bas,Jan,Lisa" or "Bas:150,Jan:120"
		friendNames, friendShares, err := parseWith(withStr)
		if err != nil {
			return err
		}

		var myShare *domain.Amount
		if mineStr != "" {
			m, err := domain.NewAmount(mineStr)
			if err != nil {
				return fmt.Errorf("--mine: %w", err)
			}
			myShare = &m
		}

		in := app.SplitInput{
			TotalAmount: total,
			Description: args[1],
			Category:    cat,
			Account:     account,
			MyShare:     myShare,
		}

		if friendShares != nil {
			in.Friends = friendShares
		} else {
			in.FriendNames = friendNames
		}

		result, err := application.Split(cmd.Context(), in)
		if err != nil {
			return err
		}

		output.PrintSplitResult(os.Stdout, result.PersonalExpense, result.Loans)
		return nil
	},
}

func parseWith(s string) ([]string, map[string]domain.Amount, error) {
	parts := strings.Split(s, ",")
	hasAmounts := false
	for _, p := range parts {
		if strings.Contains(p, ":") {
			hasAmounts = true
			break
		}
	}

	if !hasAmounts {
		names := make([]string, len(parts))
		for i, p := range parts {
			names[i] = strings.TrimSpace(p)
		}
		return names, nil, nil
	}

	shares := make(map[string]domain.Amount, len(parts))
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), ":", 2)
		if len(kv) != 2 {
			return nil, nil, fmt.Errorf("invalid --with entry: %q (expected name:amount)", p)
		}
		amount, err := domain.NewAmount(strings.TrimSpace(kv[1]))
		if err != nil {
			return nil, nil, fmt.Errorf("invalid amount for %s: %w", kv[0], err)
		}
		shares[strings.TrimSpace(kv[0])] = amount
	}
	return nil, shares, nil
}

func init() {
	splitCmd.Flags().String("cat", "", "category name")
	splitCmd.Flags().String("account", "", "account name")
	splitCmd.Flags().String("with", "", "friends: 'Bas,Jan,Lisa' (equal) or 'Bas:150,Jan:120' (unequal)")
	splitCmd.Flags().String("mine", "", "your explicit share (default: total minus friends)")
	rootCmd.AddCommand(splitCmd)
}
