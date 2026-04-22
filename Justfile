#!/usr/bin/env just --justfile

default:
    @just --list

# ----------------------------------------------------------------
# Development
# ----------------------------------------------------------------

[group('Development')]
treeclip dir="":
    treeclip run {{ dir }} -f -t -c -v --stats

[group('Development')]
build:
    go build -o ./bin/doppel ./...

[group('Development')]
test:
    go test -v ./...

[group('Development')]
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# ----------------------------------------------------------------
# Code Quality
# ----------------------------------------------------------------

[group('Code Quality')]
lint:
    golangci-lint run

[group('Code Quality')]
fmt:
    go fmt ./...

[group('Code Quality')]
vet:
    go vet ./...

[group('Code Quality')]
staticcheck:
    staticcheck ./...

[group('Code Quality')]
all-checks: fmt vet lint staticcheck

# ----------------------------------------------------------------
# Dependency
# ----------------------------------------------------------------

[group('Dependency')]
mod-download:
    go mod download

[group('Dependency')]
mod-tidy:
    go mod tidy

[group('Dependency')]
mod-vendor:
    go mod vendor

[group('Dependency')]
mod-clean:
    rm -rf ./vendor
    go clean -modcache

[group('Dependency')]
mod-update:
    go get -u ./...
    go mod tidy

# ----------------------------------------------------------------
# Git & Version Control
# ----------------------------------------------------------------

[group('Git')]
amend:
    git commit -a --amend

[group('Git')]
empty:
    git commit --allow-empty

[group('Git')]
rebase n="3":
    git rebase -i HEAD~{{ n }}

[group('Git')]
diff-cp:
    git diff | xclip -selection clipboard

[group('Git')]
today:
    git log --since="today 00:00:00" --until="today 23:59:59" --oneline