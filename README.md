# LLVM C API Searcher

## Introduction

Linux-only. Provides a `search.sh` script that allows for fuzzy searching of the LLVM C API
as the online Doxygen documentation does not expose this functionality.

Before running `search.sh`, use `go run .` to generate a cache of the LLVM C API at
`~/.local/share/llvm-c-search-cache.txt`.

Consider copying `search.sh` to `~/.local/bin/llvm-c-search` so that it can be used from anywhere.

## Dependencies

`go`, `fzf`

