default: build

build:
	fmtcheck
	go install

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/fmtcheck.sh'"

importscheck:
	@sh -c "'$(CURDIR)/scripts/importscheck.sh'"

lintcheck:
	@sh -c "'$(CURDIR)/scripts/lintcheck.sh'"

vetcheck:
	@sh -c "'$(CURDIR)/scripts/vetcheck.sh'"

.PHONY: build fmtcheck importscheck lintcheck vetcheck
