TEST?=$$(go list ./... |grep -v 'vendor')

default: build

build: fmtcheck
	go install

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/fmtcheck.sh'"

importscheck:
	@sh -c "'$(CURDIR)/scripts/importscheck.sh'"

lintcheck:
	@sh -c "'$(CURDIR)/scripts/lintcheck.sh'"

vetcheck:
	@sh -c "'$(CURDIR)/scripts/vetcheck.sh'"

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 30m

.PHONY: build fmtcheck importscheck lintcheck vetcheck testacc
