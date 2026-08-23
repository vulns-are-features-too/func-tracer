module test_files

go 1.27.0

require (
	github.com/stretchr/testify v1.12.1
	github.com/vulns-are-features-too/func-tracer/tests/test_files v0.0.1
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect

replace github.com/vulns-are-features-too/func-tracer/tests/test_files => .
