# Environment variables
GO111MODULE=on
GOPROXY=https://proxy.golang.org,direct
DOCKER ?= docker

# Default task
all: ci

# Individual tasks
dev:
	cp -f scripts/pre-commit.sh .git/hooks/pre-commit

setup:
	go mod tidy

build: ./**/*.go
	go build -o bin/ ./...

test:
	go test -failfast -race -timeout=5m ./...

cover:
	go tool cover -html=coverage.txt

fmt:
	gofumpt -w -l .
	gci write -s standard -s default -s "prefix(github.com/nikoksr/dbench)" .

lint:
	golangci-lint run ./...

ci: setup build test

release:
	NEXT=$$(svu n)
	git tag $${NEXT}
	echo $${NEXT}
	git push origin --tags

.PHONY: dev setup build test cover fmt lint ci schema-generate docs-generate docs-releases docs-imgs docs-serve docs-build release
