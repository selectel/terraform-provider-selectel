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

import:
	goimports -w $(GOFMT_FILES)

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

all: fmt import golangci-lint test testacc semgrep test-compile

build-dev:
	go build -gcflags="all=-N -l" -o terraform-provider-selectel

debug-tf:
	@echo "Cleaning old provider processes..."
	@rm -f .provider.log

	@echo "Starting provider..."
	@TF_LOG=TRACE \
	TF_DEBUG=1 dlv exec ./terraform-provider-selectel \
		--listen=127.0.0.1:40000 \
		--headless=true \
		--api-version=2 \
		--accept-multiclient \
		--continue 2>&1 | tee .provider.log &

	@echo "Waiting for provider to start..."
	@while [ ! -f .provider.log ] || ! grep -q "TF_REATTACH_PROVIDERS=" .provider.log; do sleep 0.2; done

	@REATTACH_JSON=$$(grep "TF_REATTACH_PROVIDERS=" .provider.log \
		| sed "s/.*TF_REATTACH_PROVIDERS='\(.*\)'/\1/" \
		| tr -d '\r\n' \
		| sed 's/"provider"/"registry.terraform.io\/selectel\/selectel"/'); \
	echo "TF_REATTACH_PROVIDERS='"$$REATTACH_JSON"'" terraform apply

debug-kill:
	pkill -f "dlv exec ./terraform-provider-selectel"

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
