module integration_test

go 1.27.0

require (
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/tree-sitter/go-tree-sitter v0.25.0 // indirect
	github.com/tree-sitter/tree-sitter-go v0.25.0 // indirect
	github.com/tree-sitter/tree-sitter-rust v0.24.2 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)

require (
	github.com/stretchr/testify v1.12.1
	github.com/vulns-are-features-too/func-tracer/src v0.0.1
	github.com/vulns-are-features-too/func-tracer/tests/test_files v0.0.1
)

replace (
	github.com/vulns-are-features-too/func-tracer/src => ../../src
	github.com/vulns-are-features-too/func-tracer/src/cli/caller => ../../src/cli/caller
	github.com/vulns-are-features-too/func-tracer/src/lang/index => ../../src/lang/index
	github.com/vulns-are-features-too/func-tracer/src/lang/parser => ../../src/lang/parser
	github.com/vulns-are-features-too/func-tracer/src/lang/parser/adapters => ../../src/lang/parser/adapters
	github.com/vulns-are-features-too/func-tracer/src/lang/registry => ../../src/lang/registry
	github.com/vulns-are-features-too/func-tracer/src/model => ../../src/model
	github.com/vulns-are-features-too/func-tracer/tests/test_files => ../test_files
)
