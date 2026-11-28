# Klar Linter

Goals:

- **Opinionated defaults** to encourage idiomatic code
- **Extensible.** External linters can be used, including your own. Projects are encouraged to write project-specific rules (they'll even show up in the editor).

## External Linter API

Linters don't have to be written in Klar (there is no compiler API for Klar yet). Type information about the project will be sent to linters from the Klar compiler as JSON. We will release Klar and JavaScript bindings that read the JSON.

The protocol hasn't been decided yet. This includes whether they communicate over stdio (like the Language Server Protocol), the filesystem, HTTP, etc. Discussions on GitHub are [ongoing](<>).

## Future analysis APIs for linting

Go's [`inspector`](https://pkg.go.dev/golang.org/x/tools/go/ast/inspector), including `Preorder`

## References

- `go vet`'s passes, such as [`composites`](https://github.com/golang/tools/tree/master/go/analysis/passes/composite/composite.go)
- Go's [`unitchecker`](https://github.com/golang/tools/tree/master/go/analysis/unitchecker/unitchecker.go), which runs the analyzers
- [`Analyzer`](https://github.com/golang/tools/tree/master/go/analysis/analysis/analysis.go) from `go vet`, which is similar to `Runner`
