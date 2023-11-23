#!/bin/bash
FILES=$(git diff --cached --name-only --diff-filter=ACMR)

gofumpt -l -w .
gci write -s standard -s default -s "prefix(github.com/nikoksr/dbench)" .
golangci-lint run --new --fix

git add $FILES
