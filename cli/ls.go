package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vulns-are-features-too/func-tracer/lang/registry"
)

func lsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ls",
		Short:         "List supported languages and their LSP servers",
		SilenceErrors: false,
		SilenceUsage:  false,
		Args:          cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			runLs(cmd)
		},
	}

	return cmd
}

func runLs(cmd *cobra.Command) {
	lsps := registry.ListLSP()

	stdout := cmd.OutOrStdout()
	for _, lang := range registry.ListLanguages() {
		_, _ = fmt.Fprintln(stdout, lang)

		for _, lsp := range lsps[lang] {
			_, _ = fmt.Fprintf(stdout, "- %s\n", lsp.Command())
		}
	}
}
