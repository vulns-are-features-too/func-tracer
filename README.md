# `func-tracer` - Statically trace function calls to show reachability

## Features

### Trace commands

- [x] `caller`: recursively find callers of the `target` function
- [ ] `callee`: recursively find functions called by the `target` function
- [ ] `link`: given a `caller` and a `callee` function, find a path that links them to prove/disprove that `caller` may call `callee`
- [ ] `var`: trace a `target` variable to see which functions use it or if it's ever passed to a specific function

Potential ideas:
- Calls by function name strings e.g. `exec('myfunc()')`, `eval(className + "." + funcName)`
- Indirection via types
    * [Mediators](https://refactoring.guru/design-patterns/mediator)
    * [Observers](https://refactoring.guru/design-patterns/observer)
    * maybe for common libraries/frameworks like [Mediator](https://github.com/martinothamar/Mediator)

### Languages

Supported languages and their LSP servers:

- Go
    * [gopls](https://go.dev/gopls/)
- Rust
    * [rust-analyzer](https://rust-analyzer.github.io/)

The `ls` subcommand will also list this.

### Output formats

- [x] Tree
- [ ] UML diagrams

## Examples

`just build` should give you a `func-tracer` executable.
Run this to trace callers of the `request` function in `client.go`:

```sh
./func-tracer caller --root . --file ./lang/lsp/client.go --name request
```

Examples of results can be found in [`./tests/snapshot/_snapshots/`](./tests/snapshot/_snapshots/go/) which are results of files in [`./tests/test_files/*/`](./tests/test_files/go/)

## Running the code

All dev commands can be found in the [justfile](./justfile).
