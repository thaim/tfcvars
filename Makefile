
GIT_VER := $(shell git describe --tags | sed -e 's/-/+/')

help: ## Show help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## build binaries
	go build -ldflags "-s -w -X min.Version=${GIT_VER}" -o bin/terravars

test: ## Run lint and test
	go vet ./...
	golint ./...
	go test ./...

clean: ## Remove unnecessary files
	go clean
	rm -f pkg/* bin/*
