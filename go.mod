module github.com/ProCode-Software/klar

go 1.27.0

require (
	github.com/ergochat/readline v0.1.3
	github.com/sanity-io/litter v1.5.8
	golang.org/x/sync v0.24.0
	golang.org/x/term v0.47.0
	golang.org/x/tools v0.52.0
)

require (
	github.com/aclements/go-moremath v0.0.0-20241023150245-c8bbc672ef66 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/perf v0.0.0-20261009192801-be2c69fb417e // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

ignore (
	./docs
	./samples
	./std
	node_modules
)

tool (
	golang.org/x/perf/cmd/benchstat
	golang.org/x/tools/cmd/stringer
)
