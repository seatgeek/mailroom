# renovate: datasource=go depName=github.com/golangci/golangci-lint
GOLANGCI_LINT_VERSION=v2.1.6
# renovate: datasource=go depName=gotest.tools/gotestsum
GOTESTSUM_VERSION=v1.12.3
# renovate: datasource=go depName=github.com/boumenot/gocover-cobertura
GOCOVER_COBERTURA_VERSION=v1.3.0
# renovate: datasource=go depName=github.com/vektra/mockery/v2
MOCKERY_VERSION=v3.5.1
# renovate: datasource=github-releases depName=palantir/go-license
GO_LICENSE_VERSION=v1.41.0
# renovate: datasource=github-tags depName=igorshubovych/markdownlint-cli
MARKDOWNLINT_VERSION=v0.45.0
# renovate: datasource=docker depName=pipelinecomponents/yamllint
YAMLLINT_VERSION=0.35.0

REPORTS_DIR=build/reports
CONFIG_DIR=build/config

.PHONY: all
check: clean generate lint test

.PHONY: clean
clean:
	@rm -rf $(REPORTS_DIR)
	@find . -type f -name "mock_*.go" -delete

.PHONY: lint
lint: lint-go lint-license lint-markdown lint-yaml

.PHONY: lint-go
lint-go: .install-linter generate
	@golangci-lint run --config $(CONFIG_DIR)/.golangci.yml ./...

.PHONY: lint-fix
lint-fix: .install-linter .install-go-license
	@golangci-lint run --fix --config $(CONFIG_DIR)/.golangci.yml ./...
	@find . -type f -name '*.go' ! -name 'mock_*.go' | xargs go-license --config $(CONFIG_DIR)/.go-license.yaml

.PHONY: lint-license
lint-license: .install-go-license
	@find . -type f -name '*.go' ! -name 'mock_*.go' | xargs go-license --config $(CONFIG_DIR)/.go-license.yaml --verify

.PHONY: lint-markdown
lint-markdown:
	@docker run -it --rm -v `pwd`:/workdir:ro ghcr.io/igorshubovych/markdownlint-cli:$(MARKDOWNLINT_VERSION) --config $(CONFIG_DIR)/.markdownlint.yaml .

.PHONY: lint-yaml
lint-yaml:
	@docker run -it --rm -v `pwd`:/code:ro pipelinecomponents/yamllint:$(YAMLLINT_VERSION) yamllint --config-file $(CONFIG_DIR)/.yamllint.yaml .

.PHONY: test
test: .install-gotestsum .ensure-reports-dir generate-mocks
	@gotestsum --junitfile $(REPORTS_DIR)/unit-tests.xml -- -race -v -coverprofile=$(REPORTS_DIR)/coverage.out -covermode=atomic -v -cover ./...

.PHONY: report-coverage
report-coverage: .install-cover-cobertura .ensure-reports-dir
	@sed -i.bak '/\/mock_.*\.go/d' $(REPORTS_DIR)/coverage.out
	@sed -i.bak '/internal\/example\.go/d' $(REPORTS_DIR)/coverage.out
	@go tool cover -func=$(REPORTS_DIR)/coverage.out
	@gocover-cobertura < $(REPORTS_DIR)/coverage.out > $(REPORTS_DIR)/coverage.xml

.PHONY: test-with-coverage
test-with-coverage: test report-coverage

.PHONY: generate
generate: clean generate-mocks

.PHONY: generate-mocks
generate-mocks: .install-mockery
	@mockery --config $(CONFIG_DIR)/.mockery.yaml

.install-linter:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.install-go-license:
	@go install github.com/palantir/go-license@$(GO_LICENSE_VERSION)

.install-gotestsum:
	@go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)

.install-cover-cobertura:
	@go install github.com/boumenot/gocover-cobertura@$(GOCOVER_COBERTURA_VERSION)

.install-mockery:
	@go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)

.ensure-reports-dir:
	@mkdir -p $(REPORTS_DIR)
