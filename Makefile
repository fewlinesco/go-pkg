CREATED_AT := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

GO_BIN := go
GO_STATICCHECK_BIN := $(GO_BIN) run ./vendor/honnef.co/go/tools/cmd/staticcheck
GO_FMT_BIN := gofmt
GO_LINT_BIN := $(GO_BIN) run ./vendor/golang.org/x/lint/golint
GO_VERSION := $(shell awk '/^golang / {print $$2}' .tool-versions)

GH_WORKFLOWS_TPL_DIR := .github/workflows
GH_WORKFLOWS_TPL := $(wildcard $(GH_WORKFLOWS_TPL_DIR)/*.yaml.in)
GH_WORKFLOWS := $(GH_WORKFLOWS_TPL:%.in=%)

PHONY: generate-github-workflows test test-unit test-fmt test-lint test-staticcheck test-github-workflows

generate-github-workflows: $(GH_WORKFLOWS)

$(GH_WORKFLOWS): %.yaml: %.yaml.in
	@echo "+ generate-github-workflow ($@)"
	@echo "# File generated by make; DO NOT EDIT" > $@
	@sed -e "s/<GO_VERSION>/$(GO_VERSION)/" $< >> $@

$(GH_WORKFLOWS): .tool-versions

test: test-unit test-fmt test-lint test-staticcheck test-github-workflows-up-to-date

test-github-workflows-up-to-date: $(GH_WORKFLOWS)
	@for workflow in "$(GH_WORKFLOWS)"; do \
		echo "+ $@ ($${workflow})"; \
		test -z "$$(git diff $${workflow} | tee /dev/stderr)" || \
			( >&2 echo "=> please regenerate github workflows with 'make generate-github-workflows'" && false); \
	done

test-unit:
	@echo "+ $@"
	@$(GO_BIN) test -p=1 ./...

test-fmt:
	@echo "+ $@"
	@test -z "$$($(GO_FMT_BIN) -l -e -s platform | tee /dev/stderr)" || \
	  ( >&2 echo "=> please format Go code with '$(GO_FMT_BIN) -s -w .'" && false)

test-lint:
	@echo "+ $@"
	@test -z "$$($(GO_LINT_BIN) ./platform/... | tee /dev/stderr)"

test-staticcheck:
	@echo "+ $@"
	@$(GO_STATICCHECK_BIN) ./platform/...

test-tidy:
	@echo "+ $@"
	@$(GO_BIN) mod tidy
	@test -z "$$($(GIT_BIN) status --short go.mod go.sum | tee /dev/stderr)" || \
	  ( >&2 echo "=> please tidy the Go modules with '$(GO_BIN) mod tidy'" && false)
