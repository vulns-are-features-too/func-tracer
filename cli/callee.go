package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/vulns-are-features-too/func-tracer/logging"
	"github.com/vulns-are-features-too/func-tracer/model"
	"github.com/vulns-are-features-too/func-tracer/tracer"
	"github.com/vulns-are-features-too/func-tracer/tracer/graph"
)

const calleeExample = `
// specify project root, target file, and function name
callee -r . -f . main.go -n myfunc
// specify target file, and function line/column
callee -f main.go -l 3 -c 5
`

func calleeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "callee",
		Short:         "Trace function callees",
		SilenceErrors: false,
		SilenceUsage:  false,
		Args:          cobra.NoArgs,
		Example:       calleeExample,
		PreRunE:       tracePreRunE,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCallee(cmd)
		},
	}

	parseArgs(cmd)

	return cmd
}

func runCallee(cmd *cobra.Command) error {
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
		return t.TraceCallees
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

	reportCallees(cmd, result, &target, args.Root)

	return nil
}
