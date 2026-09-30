TEST?=$$(go list ./...)
GOFMT_FILES?=$$(find . -name '*.go')
GOLANGCI_VERSION?=v2.12.2
PKG_NAME=selectel
TFPLUGINDOCS_VERSION?=v0.25.0
TFPLUGINDOCS=go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@$(TFPLUGINDOCS_VERSION)
MISSPELL_VERSION?=v0.8.0
MISSPELL=go run github.com/golangci/misspell/cmd/misspell@$(MISSPELL_VERSION)
DOCS_EXAMPLES=$(wildcard examples/resources/selectel_* examples/data-sources/selectel_*)

default: build

golangci-lint:
	docker run --rm -v ${PWD}:/app:ro -w /app golangci/golangci-lint:$(GOLANGCI_VERSION) golangci-lint run

go-fix:
	go fix ./...

build:
	go build

test:
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) $(TESTARGS) -timeout 360m

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -w $(GOFMT_FILES)

import:
	goimports -w $(GOFMT_FILES)

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

all: fmt import golangci-lint test testacc semgrep pin-sha-tags test-compile

docs:
	$(TFPLUGINDOCS) generate

docs-check:
	@snapshot=$$(mktemp -d); trap 'rm -rf "$$snapshot"' EXIT; cp -R docs "$$snapshot/docs" || exit 1; \
	for cmd in generate validate; do \
		if ! out=$$($(TFPLUGINDOCS) $$cmd 2>&1); then \
			err=$$(echo "$$out" | sed -n '/^Error executing command/,$$p'); echo "$${err:-$$out}"; exit 1; \
		fi; \
	done; \
	if ! diff -rq "$$snapshot/docs" docs; then \
		echo "docs/ was out of date and has been regenerated, review and commit the changes"; \
		exit 1; \
	fi

docs-misspell:
	$(MISSPELL) -error -source text templates/ docs/

examples-fmt:
	@for dir in $(DOCS_EXAMPLES); do terraform fmt $$dir; done

examples-check:
	@for dir in $(DOCS_EXAMPLES); do terraform fmt -check -diff $$dir || exit 1; done

# CLI reference:
# https://github.com/aquasecurity/trivy/blob/main/docs/docs/references/configuration/cli/trivy_filesystem.md
trivy:
	docker run --rm -v ${PWD}:/app:ro -w /app aquasec/trivy fs --exit-code 1 --no-progress --ignorefile .trivyignore.yml .

# CLI reference:
# https://semgrep.dev/docs/cli-reference
semgrep:
	docker run --rm -v ${PWD}:/app:ro -w /app semgrep/semgrep semgrep scan --error --metrics=off \
		--config=p/command-injection \
		--config=p/comment \
		--config=p/cwe-top-25 \
		--config=p/default \
		--config=p/gitleaks \
		--config=p/golang \
		--config=p/gosec \
		--config=p/insecure-transport \
		--config=p/owasp-top-ten \
		--config=p/r2c-best-practices \
		--config=p/r2c-bug-scan \
		--config=p/r2c-security-audit \
		--config=p/secrets \
		--config=p/security-audit \
		--config=p/sql-injection \
		--config=p/xss \
		.

pin-sha-tags:
	docker run --rm \
		-v ${PWD}/.github/workflows:/workflows \
		mheap/pin-github-action@sha256:1a336147444c5be62b5bade8a7b82d971d138aac3c799f9ad72d0598331210aa .

.PHONY: golangci-lint go-fix build test testacc fmt test-compile docs docs-check examples-fmt examples-check docs-misspell trivy semgrep pin-sha-tags
