TEST?=$$(go list ./...)
GOFMT_FILES?=$$(find . -name '*.go')
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=selectel

default: build

golangci-lint:
	@sh -c "'$(CURDIR)/scripts/golangci_lint_check.sh'"

build:
	go build

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) $(TESTARGS) -timeout 360m

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -w $(GOFMT_FILES)

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)


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

.PHONY: golangci-lint build test testacc fmt test-compile semgrep website website-test
