.PHONY: all build clean image check vendor dependencies

NAME				:= azure-devops-exporter
GIT_TAG				:= $(shell git describe --dirty --tags --always)
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
LDFLAGS             := -X "main.gitTag=$(GIT_TAG)" -X "main.gitCommit=$(GIT_COMMIT)" -extldflags "-static"

PKGS				:= $(shell go list ./... | grep -v -E '/vendor/|/test')
FIRST_GOPATH		:= $(firstword $(subst :, ,$(shell go env GOPATH)))
GOLANGCI_LINT_BIN	:= $(FIRST_GOPATH)/bin/golangci-lint


all: build

clean:
	git clean -Xfd .

build:
	CGO_ENABLED=0 go build -a -ldflags '$(LDFLAGS)' -o $(NAME) .

vendor:
	go mod tidy
	go mod vendor
	go mod verify

image: build
	docker build -t $(NAME):$(TAG) .

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	# megacheck fails to respect build flags, causing compilation failure during linting.
	# instead, use the unused, gosimple, and staticcheck linters directly
	$(GOLANGCI_LINT_BIN) run -D megacheck -E unused,gosimple,staticcheck --timeout=10m

dependencies: $(GOLANGCI_LINT_BIN)

$(GOLANGCI_LINT_BIN):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(FIRST_GOPATH)/bin v1.30.0

