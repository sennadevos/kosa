package cmd

import (
	"github.com/spf13/cobra"
)

var (
	flagJSON   bool
	flagToon   bool
	flagConfig string
)

var rootCmd = &cobra.Command{
	Use:   "kosa",
	Short: "personal finance from the terminal",
	Long:  "kosa — track transactions, loans, balances, and recurring rules.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output as json")
	rootCmd.PersistentFlags().BoolVar(&flagToon, "toon", false, "output in token-optimized notation")
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "config file path")
}

func Execute() error {
	return rootCmd.Execute()
}
