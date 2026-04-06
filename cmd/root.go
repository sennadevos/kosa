package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sennadevos/kosa/internal/app"
	"github.com/sennadevos/kosa/internal/backend"
	"github.com/sennadevos/kosa/internal/config"
	"github.com/sennadevos/kosa/internal/output"
)

var (
	flagJSON   bool
	flagToon   bool
	flagConfig string

	application *app.App
)

func outputFormat() output.Format {
	if flagJSON {
		return output.FormatJSON
	}
	if flagToon {
		return output.FormatToon
	}
	return output.FormatTable
}

var rootCmd = &cobra.Command{
	Use:   "kosa",
	Short: "personal finance from the terminal",
	Long:  "kosa — track transactions, loans, balances, and recurring rules.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(flagConfig)
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}

		b, err := backend.New(cfg)
		if err != nil {
			return fmt.Errorf("backend: %w", err)
		}

		application = app.New(b, cfg)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output as json")
	rootCmd.PersistentFlags().BoolVar(&flagToon, "toon", false, "output in token-optimized notation")
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "config file path")
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
