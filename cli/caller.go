package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/vulns-are-features-too/func-tracer/logging"
	"github.com/vulns-are-features-too/func-tracer/model"
	"github.com/vulns-are-features-too/func-tracer/tracer"
	"github.com/vulns-are-features-too/func-tracer/tracer/graph"
)

const callerExample = `
// specify project root, target file, and function name
caller -r . -f . main.go -n myfunc
// specify target file, and function line/column
caller -f main.go -l 3 -c 5
`

func callerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "caller",
		Short:         "Trace function callers",
		SilenceErrors: false,
		SilenceUsage:  false,
		Args:          cobra.NoArgs,
		Example:       callerExample,
		PreRunE:       tracePreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCaller(cmd)
		},
	}

	parseArgs(cmd)

	return cmd
}

func runCaller(cmd *cobra.Command) error {
	logger := logging.New(logging.LevelFromVerbosity(args.Verbosity))

	runner, err := initRunner(logger)
	if err != nil {
		return err
	}

	fnTrace := func(t *tracer.Tracer) func(
		ctx context.Context,
		target *model.Symbol,
		maxDepth int,
	) (*graph.Graph, error) {
		return t.TraceCallers
	}

	target, err := runner.findTarget()
	if err != nil {
		return err
	}

	result, err := runner.trace(context.Background(), &target, fnTrace)
	if err != nil {
		return err
	}

	runner.Close()

	reportCallers(cmd, result, &target, args.Root)

	return nil
}
