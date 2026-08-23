package caller

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vulns-are-features-too/func-tracer/src/model"
	"github.com/vulns-are-features-too/func-tracer/src/tracer/graph"
)

const indent string = "  "

func report(cmd *cobra.Command, graph *graph.Graph, target *model.Symbol, projectRoot string) {
	rootPath, err := filepath.Abs(projectRoot)
	if err == nil {
		projectRoot = rootPath
	}

	stdout := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(stdout, "%s\n", formatSymbol(projectRoot, target))
	graph.WalkCallers(target, func(curr *model.Symbol, level int) {
		_, _ = fmt.Fprintf(
			stdout,
			"%s- %s\n",
			strings.Repeat(indent, level),
			formatSymbol(projectRoot, curr),
		)
	})
}

func formatSymbol(projectRoot string, s *model.Symbol) string {
	loc := s.Location
	pos := loc.Range.Start
	file := normalizePath(projectRoot, loc.URI)

	return fmt.Sprintf("%s @ %s %d:%d", s.Name, file, pos.Line, pos.Character)
}

func normalizePath(projectRoot, fileURL string) string {
	const schemaPrefix = "file://"
	if !strings.HasPrefix(fileURL, schemaPrefix) {
		return fileURL
	}

	fileURL = strings.TrimPrefix(fileURL, schemaPrefix)

	if rel, err := filepath.Rel(projectRoot, fileURL); err == nil {
		fileURL = rel
	}

	return fileURL
}
