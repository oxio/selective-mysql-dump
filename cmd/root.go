package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "smdump",
	Short:   "Selectively dump MySQL/MariaDB database",
	Version: "1.0.0",
}

func Execute() {
	var rootCmd = &cobra.Command{
		Use:   "dump",
		Short: "Creates a MySQL/MariaDB dump file",
		RunE: func(cmd *cobra.Command, args []string) error {
			// todo: implement
			return nil
		},
	}

	_ = rootCmd.Execute()
}
