#!/bin/bash

# Check Go files with the linter.
files=$(go list ./... | grep -v vendor/)
golint -set_exit_status $files
