# GIT_VERSION := $(shell git describe --abbrev=0 --tags)
GIT_VERSION := "develop"
GIT_REVISION := $(shell git rev-list -1 HEAD)
DATE := $(shell date +%Y-%m-%dT%H:%M%Sz)

help: ## Show help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## build binaries
	CGO_ENABLED=0 go build -ldflags "-s -w -X min.Version=${GIT_VERSION} -X main.revision=${GIT_REVISION} -X main.buildDate=${DATE}" -trimpath -o bin/tfcvars

test: ## Run lint and test
	go vet ./...
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

clean: ## Remove unnecessary files
	go clean
	rm -f pkg/* bin/*

fmt: ## Run go fmt
	@./scripts/gofmt.sh
