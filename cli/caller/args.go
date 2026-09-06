package caller

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

var args = struct {
	Root      string
	File      string
	Name      string
	Line      uint
	Column    uint
	Depth     int
	Workers   int
	Verbosity int
}{}

var (
	errNegativeDepth        = errors.New("depth cannot be negative")
	errAtLeast1WorkerNeeded = errors.New("workers must be greater than zero")
	errFileNotFound         = errors.New("file not found")
)

func parseArgs(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(&args.Root, "root", "r", ".", "workspace root")
	flags.StringVarP(&args.File, "file", "f", "", "target file")

	flags.StringVarP(&args.Name, "name", "n", "", "function name")
	flags.UintVarP(&args.Line, "line", "l", 0, "target line, zero based")
	flags.UintVarP(&args.Column, "column", "c", 0, "target column, zero based")

	flags.IntVarP(&args.Depth, "depth", "d", 0, "maximum caller depth (0=unlimited)")

	flags.IntVarP(&args.Workers, "workers", "w", 1, "number of parallel workers")

	flags.CountVarP(&args.Verbosity, "verbose", "v", "verbosity level")
}

func validateArgs() error {
	if _, err := os.Stat(args.File); errors.Is(err, os.ErrNotExist) {
		return errFileNotFound
	}

	if args.Depth < 0 {
		return errNegativeDepth
	}

	if args.Workers < 1 {
		return errAtLeast1WorkerNeeded
	}

	return nil
}
