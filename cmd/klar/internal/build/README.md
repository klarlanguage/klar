# klar build

This package implements the `klar build` CLI command.

## The compiler pipeline

The compiler pipeline runs here. It includes:

- **Input resolution:** Creating `build.Input` objects from each command line input
- **CLI flag overrides:** Allowing flags like `--target` and `--bundle` to override klar.build settings
- **Output reporting:** The "build succeeded/failed" message

This is also exported so other commands like `klar run`, `klar test`, `klar lint`, and `klar format` can use it.

### How other commands use it
