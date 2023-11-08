#!/usr/bin/env sh

# Credit to https://github.com/goreleaser/goreleaser

set -e

rm -rf completions
mkdir completions

for sh in bash zsh fish; do
	go run main.go completion "$sh" >"completions/dbench.$sh"
done
