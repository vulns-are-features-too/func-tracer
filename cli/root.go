// Package cli provides CLI commands.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vulns-are-features-too/func-tracer/cli/caller"
	"github.com/vulns-are-features-too/func-tracer/cli/ls"
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

	cmd.AddCommand(caller.Cmd())
	cmd.AddCommand(ls.Cmd())

	return cmd
}
