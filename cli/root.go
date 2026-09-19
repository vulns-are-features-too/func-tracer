// Package cli provides CLI commands.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute is the CLI entry point.
func Execute() {
	cmd := RootCmd()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// RootCmd to .Execute().
func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "func-trace",
		Short: "Trace function callers",
	}

	cmd.AddCommand(calleeCmd())
	cmd.AddCommand(callerCmd())
	cmd.AddCommand(lsCmd())

	return cmd
}
