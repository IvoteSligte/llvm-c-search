#!/usr/bin/env bash

cachePath="$HOME/.local/share/llvm-c-search-cache.txt"

if ! [ -f "$cachePath" ]; then
    echo "Cache does not exist. Use the go script to generate the cache."
    exit 1
fi

if ! command -v "fzf"; then
    echo "fzf must be installed and in $PATH."
    exit 1
fi

cat "$cachePath" | sed -n 's/.* \(.*\) (.*/\1/p' | fzf
