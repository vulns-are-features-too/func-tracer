alias b := build
alias r := run
alias l := lint
alias f := fixup
alias t := test
alias tu := unit_test
alias tr := race_test
alias ti := integration_test
alias ts := snapshot_test
alias tc := test_coverage
alias cov := test_coverage

set dotenv-load
set dotenv-path := "./.env"

ROOT := justfile_directory()
GOCOVERDIR := ROOT / ".coverage"
MODULE := "github.com/vulns-are-features-too/func-tracer"

[working-directory: '.']
build:
  go build -o ../func-tracer .

[working-directory: '.']
run:
  go run .

_all *CMD:
  {{CMD}}
  cd tests/integration && {{CMD}}
  cd tests/snapshot && {{CMD}}
  cd tests/test_files && {{CMD}}

lint:
  @just _all golangci-lint run

fixup: tidy fmt fix_lint

tidy:
  @just _all go mod tidy

fmt:
  @just _all golangci-lint fmt

fix_lint:
  @just _all golangci-lint run --fix

test: validate_test_files unit_test race_test integration_test snapshot_test

[working-directory: './tests/test_files']
@validate_test_files *FLAGS:
  echo "Validating test files"
  go test . {{FLAGS}}

[working-directory: '.']
@unit_test *FLAGS:
  echo "Running unit tests"
  go test ./... {{FLAGS}}

[working-directory: '.']
@race_test *FLAGS:
  echo "Running race tests"
  go test -tags race -race ./... {{FLAGS}}

[working-directory: './tests/integration']
@integration_test *FLAGS:
  echo "Running integration tests"
  go test . {{FLAGS}}

[working-directory: './tests/snapshot']
@snapshot_test *FLAGS:
  echo "Running snapshot tests"
  go test . {{FLAGS}}

[working-directory: './tests/snapshot']
@update_test_snapshot *FLAGS:
  echo "Updating test snapshots"
  UPDATE_SNAPSHOT= go test . {{FLAGS}}

@test_coverage:
  echo "Running tests with coverage"
  mkdir -p {{GOCOVERDIR}}
  just unit_test -coverprofile={{GOCOVERDIR}}/unit.out
  just integration_test -coverprofile={{GOCOVERDIR}}/integration.out -coverpkg={{MODULE}}/...
  just snapshot_test -coverprofile={{GOCOVERDIR}}/snapshot.out -coverpkg={{MODULE}}/...
  cd {{GOCOVERDIR}} && cat unit.out <(tail -n +2 integration.out) <(tail -n +2 snapshot.out) > all.out
  go tool cover -html={{GOCOVERDIR}}/all.out
