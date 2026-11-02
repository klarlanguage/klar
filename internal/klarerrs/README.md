# Klar Compiler Errors (`klarerrs`)

The `klarerrs` package implements most of Klar's compiler errors (and warnings).

In this package are:

- Error structures to provide detailed information about compiler errors
- All error codes used by the compiler, and their messages
- Presets for errors
- Utilities for formatting error messages ([format.go](./format.go))
- Tests to check for unused error codes and codes without a message

## Conventions for writing error messages
