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

build:
	go generate ./...
	go build -o bin/ -ldflags="-s -w -X github.com/nikoksr/dbench/internal/build.Version=$(shell git describe --tags --always --dirty)"

clean:
	@rm -rf bin/ \
		completions/ \
		dist/ \
		tmp/ \
		dbench/ \
		dbench* \
		coverage.txt

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

local: build
	rm -rf completions/dev
	mkdir -p completions/dev
	./bin/dbench completion fish > completions/dev/dbench.fish
	sudo cp ./bin/dbench /usr/local/bin/dbench
	sudo cp completions/dev/dbench.fish /usr/share/fish/vendor_completions.d/dbench.fish

release:
	NEXT=$$(svu n)
	git tag $${NEXT}
	echo $${NEXT}
	git push origin --tags

.PHONY: all dev setup build docker-build docker-run clean test cover fmt lint ci release
