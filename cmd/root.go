package cmd

import (
	"fmt"
	"os"

	"github.com/oxio/selective-mysql-dump/internal"
	"github.com/spf13/cobra"
)

var (
	configFile string
	outputFile string
)

var rootCmd = &cobra.Command{
	Use:     "smdump",
	Short:   "Selectively dump MySQL/MariaDB database",
	Long:    "A tool for selectively dumping MySQL/MariaDB database tables based on a YAML configuration file.",
	Version: "1.0.0",
	RunE:    runDump,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "Path to config file (default: ./.smdump.yaml)")
	rootCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Output file path (default: stdout)")
}

// runDump is the main execution function
func runDump(cmd *cobra.Command, args []string) error {
	// Load config
	config, err := internal.LoadConfig(configFile)
	if err != nil {
		// Check if it's a config not found error
		if _, ok := err.(*internal.ConfigNotFoundError); ok {
			fmt.Fprintln(os.Stderr, err.(*internal.ConfigNotFoundError).ShowHelpMessage())
			return fmt.Errorf("config file not found")
		}
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate tables
	if err := internal.ValidateTables(config.Tables.WithData, config.Tables.StructureOnly); err != nil {
		return fmt.Errorf("invalid table configuration: %w", err)
	}

	// Check if password is in DSN
	password, err := internal.GetPasswordFromDSN(config.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Prompt for password if not set
	var dsn string
	if password == "" {
		promptPassword, err := internal.PromptForPassword()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}

		// Rebuild DSN with password
		parsed, err := internal.ParseDSN(config.DSN)
		if err != nil {
			return fmt.Errorf("failed to parse DSN: %w", err)
		}
		dsn = internal.BuildDSN(parsed, promptPassword)
	} else {
		dsn = config.DSN
	}

	// Create executor
	executor := internal.NewExecutor(dsn, outputFile)

	// Execute dump
	if err := executor.DumpAll(config.Tables.WithData, config.Tables.StructureOnly); err != nil {
		return err
	}

	return nil
}
