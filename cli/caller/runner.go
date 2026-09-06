package caller

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vulns-are-features-too/func-tracer/lang"
	"github.com/vulns-are-features-too/func-tracer/lang/index"
	"github.com/vulns-are-features-too/func-tracer/lang/lsp"
	"github.com/vulns-are-features-too/func-tracer/lang/parser"
	"github.com/vulns-are-features-too/func-tracer/lang/registry"
	"github.com/vulns-are-features-too/func-tracer/logging"
	"github.com/vulns-are-features-too/func-tracer/model"
	"github.com/vulns-are-features-too/func-tracer/tracer"
	"github.com/vulns-are-features-too/func-tracer/tracer/graph"
)

var (
	errLanguageDetection = errors.New("language detection failed")
	errParserInit        = errors.New("parser init failed")
	errLspInit           = errors.New("LSP init failed")
	errIndexing          = errors.New("indexing failed")
	errFindingTarget     = errors.New("couldn't find target")
	errTracing           = errors.New("tracing failed")
)

type runner struct {
	logger logging.Logger
	lang   lang.Language
	parser parser.Adapter
	lsp    lsp.Adapter
	idx    index.Index
}

func (r *runner) detectLanguage() error {
	lang, err := registry.Detect(args.File)
	if err != nil {
		return fmt.Errorf("%w: %w", errLanguageDetection, err)
	}

	r.logger.Infof("Language detected: %s", lang)
	r.lang = lang

	return nil
}

func (r *runner) initParser() error {
	parser, err := registry.Parser(r.lang)
	if err != nil {
		return fmt.Errorf("%w: %w", errParserInit, err)
	}

	r.parser = parser

	return nil
}

func (r *runner) initLsp() error {
	lsp, err := registry.LSP(r.lang)
	if err != nil {
		return fmt.Errorf("%w: %w", errLspInit, err)
	}

	r.lsp = lsp

	return nil
}

func (r *runner) index() error {
	r.logger.Infof("Indexing project from root: %s", args.Root)

	files, err := sourceFiles(args.Root, r.lang)
	if err != nil {
		return fmt.Errorf("%w: %w", errIndexing, err)
	}

	r.logger.Infof("Source files found: %s", len(files))

	r.idx = index.New(r.logger, r.parser)
	if err := r.idx.Build(files); err != nil {
		return fmt.Errorf("%w: %w", errIndexing, err)
	}

	r.logger.Infof("Indexed")

	return nil
}

func (r *runner) findTarget() (model.Symbol, error) {
	fileURI, err := fileURI(args.File)
	if err != nil {
		return model.Symbol{}, fmt.Errorf("%w: %w", errFindingTarget, err)
	}

	if args.Name != "" {
		return r.findTargetByName(fileURI, args.Name)
	}

	return r.findTargetAt(fileURI, model.Position{
		Line:      args.Line,
		Character: args.Column,
	})
}

func (r *runner) findTargetAt(fileURI string, pos model.Position) (model.Symbol, error) {
	target, ok := r.idx.FindFunction(model.Location{
		URI:   fileURI,
		Range: model.Range{Start: pos, End: pos},
	})

	if !ok {
		return target, fmt.Errorf(
			"%w: no function found at %s:%d:%d",
			errFindingTarget,
			args.File,
			args.Line,
			args.Column,
		)
	}

	return target, nil
}

func (r *runner) findTargetByName(fileURI string, name string) (model.Symbol, error) {
	// TODO: support class methods (class.function) and resolve duplicate names
	target, ok := r.idx.FindFunctionByName(fileURI, name)
	if !ok {
		return target, fmt.Errorf(
			"%w: no function found in %s named %s",
			errFindingTarget,
			args.File,
			args.Name,
		)
	}

	return target, nil
}

func (r *runner) trace(ctx context.Context, target *model.Symbol) (*graph.Graph, error) {
	r.logger.Infof("Starting LSP server: %s", r.lsp.Command())

	session, err := lsp.Start(ctx, r.logger, r.lsp, args.Root)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errTracing, err)
	}
	defer session.Close(ctx)

	r.logger.Infof("Tracing with %d workers", args.Workers)
	tracer := tracer.New(
		r.logger,
		session,
		r.idx,
		args.Workers,
	)

	g, err := tracer.TraceCallers(
		ctx,
		target,
		args.Depth,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errTracing, err)
	}

	return g, nil
}

func run(ctx context.Context, cmd *cobra.Command, logger logging.Logger) error {
	r := runner{logger: logger}
	r.logger.Infof("Starting trace")

	if err := r.detectLanguage(); err != nil {
		return err
	}

	if err := r.initParser(); err != nil {
		return err
	}

	if err := r.initLsp(); err != nil {
		return err
	}

	if err := r.index(); err != nil {
		return err
	}
	defer r.idx.Close()

	target, err := r.findTarget()
	if err != nil {
		return err
	}

	result, err := r.trace(ctx, &target)
	if err != nil {
		return err
	}

	report(cmd, result, &target, args.Root)

	return nil
}

//nolint:wrapcheck
func fileURI(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	uri := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(absolute),
	}).String()

	return uri, nil
}
