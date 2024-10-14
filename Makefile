# renovate: datasource=go depName=github.com/golangci/golangci-lint
GOLANGCI_LINT_VERSION=v1.60.3
# renovate: datasource=go depName=gotest.tools/gotestsum
GOTESTSUM_VERSION=v1.12.0
# renovate: datasource=go depName=github.com/boumenot/gocover-cobertura
GOCOVER_COBERTURA_VERSION=v1.2.0
# renovate: datasource=github-releases depName=mockery/mockery
MOCKERY_VERSION=v2.42.0
# renovate: datasource=github-releases depName=palantir/go-license
GO_LICENSE_VERSION=v1.39.0
# renovate: datasource=github-tags depName=igorshubovych/markdownlint-cli
MARKDOWNLINT_VERSION=v0.42.0
# renovate: datasource=docker depName=pipelinecomponents/yamllint
YAMLLINT_VERSION=0.32.1

.PHONY: all
check: clean generate lint test

.PHONY: clean
clean:
	@rm -rf build
	@find . -type f -name "mock_*.go" -delete

.PHONY: lint
lint: lint-go lint-license lint-markdown lint-yaml

.PHONY: lint-go
lint-go: .install-linter generate
	@golangci-lint run ./...

.PHONY: lint-fix
lint-fix: .install-linter .install-go-license
	@golangci-lint run --fix ./...
	@find . -type f -name '*.go' ! -name 'mock_*.go' | xargs go-license --config .go-license.yaml

.PHONY: lint-license
lint-license: .install-go-license
	@find . -type f -name '*.go' ! -name 'mock_*.go' | xargs go-license --config .go-license.yaml --verify

.PHONY: lint-markdown
lint-markdown:
	@docker run -it --rm -v `pwd`:/workdir:ro ghcr.io/igorshubovych/markdownlint-cli:$(MARKDOWNLINT_VERSION) .

.PHONY: lint-yaml
lint-yaml:
	@docker run -it --rm -v `pwd`:/code:ro pipelinecomponents/yamllint:$(YAMLLINT_VERSION) yamllint .

.PHONY: test
test: .install-gotestsum .ensure-build-dir generate-mocks
	@gotestsum --junitfile build/unit-tests.xml -- -race -v -coverprofile=build/coverage.out -covermode=atomic -v -cover ./...

.PHONY: report-coverage
report-coverage: .install-cover-cobertura .ensure-build-dir
	@sed -i.bak '/\/mock_.*\.go/d' build/coverage.out
	@go tool cover -func=build/coverage.out
	@gocover-cobertura < build/coverage.out > build/coverage.xml

.PHONY: test-with-coverage
test-with-coverage: test report-coverage

.PHONY: generate
generate: clean generate-mocks

.PHONY: generate-mocks
generate-mocks: .install-mockery
	@mockery

.install-linter:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.install-go-license:
	@go install github.com/palantir/go-license@$(GO_LICENSE_VERSION)

.install-gotestsum:
	@go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)

.install-cover-cobertura:
	@go install github.com/boumenot/gocover-cobertura@$(GOCOVER_COBERTURA_VERSION)

.install-mockery:
	@go install github.com/vektra/mockery/v2@$(MOCKERY_VERSION)

.ensure-build-dir:
	@mkdir -p build
