// Package caller provides cmd to trace callers.
package caller

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/vulns-are-features-too/func-tracer/logging"
)

const example = `
// specify project root, target file, and function name
caller -r . -f . main.go -n myfunc
// specify target file, and function line/column
caller -f main.go -l 3 -c 5
`

// Cmd "caller".
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "caller",
		Short:         "Trace function callers",
		SilenceErrors: false,
		SilenceUsage:  false,
		Args:          cobra.NoArgs,
		Example:       example,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			err := validateArgs()
			if err != nil {
				return err
			}

			err = cmd.MarkFlagRequired("file")
			if err != nil {
				panic(err)
			}

			// either name or line+column must be used
			cmd.MarkFlagsOneRequired("name", "line")
			cmd.MarkFlagsMutuallyExclusive("name", "line")
			cmd.MarkFlagsOneRequired("name", "column")
			cmd.MarkFlagsMutuallyExclusive("name", "column")

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger := logging.New(logging.LevelFromVerbosity(args.Verbosity))
			if err := run(context.Background(), cmd, logger); err != nil {
				logger.Errorf("caller cmd failed: %s", err.Error())
				os.Exit(1)
			}

			return nil
		},
	}

	parseArgs(cmd)

	return cmd
}
