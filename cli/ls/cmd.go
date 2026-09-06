// Package ls provides cmd to list supported languages & LSP servers.
package ls

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulns-are-features-too/func-tracer/lang/registry"
)

// Cmd "ls".
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ls",
		Short:         "List supported languages and their LSP servers",
		SilenceErrors: false,
		SilenceUsage:  false,
		Args:          cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			run(cmd)
		},
	}

	return cmd
}

func run(cmd *cobra.Command) {
	lsps := registry.ListLSP()

	stdout := cmd.OutOrStdout()
	for _, lang := range registry.ListLanguages() {
		_, _ = fmt.Fprintln(stdout, lang)

		for _, lsp := range lsps[lang] {
			_, _ = fmt.Fprintf(stdout, "- %s\n", lsp.Command())
		}
	}
}
