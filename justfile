# ttanic task runner -- dev-time only, never needed by end users.

export CGO_ENABLED := "0"

# list available recipes
default:
    @just --list

# compile all packages
build:
    go build ./...

# run all tests
test:
    go test ./...

# run tests with a coverage profile and print the per-function summary
cover:
    go test -coverprofile=cover.out ./...
    go tool cover -func=cover.out

# run golangci-lint
lint:
    golangci-lint run

# rewrite all Go sources into canonical gofmt style
fmt:
    gofmt -w .

# build and run the ttanic binary
run *args:
    go run ./cmd/ttanic {{ args }}

# build a local snapshot release (requires goreleaser)
snapshot:
    goreleaser release --snapshot --clean

# everything CI runs: fmt-check + lint + test
ci: fmt-check lint test

# fail if any file is not gofmt-formatted (changes nothing)
fmt-check:
    @unformatted="$(gofmt -l .)"; if [ -n "$unformatted" ]; then echo "gofmt needed on:" >&2; echo "$unformatted" >&2; exit 1; fi
