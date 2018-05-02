#!/usr/bin/env bash

# Check linter
echo "==> Checking that code complies with golint requirements..."
golint_files=$(go list ./... | grep -v vendor/)
golint -set_exit_status $golint_files

exit 0
