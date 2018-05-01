#!/bin/bash

# Check Go files with the vet tool.
files=$(go list ./... | grep -v vendor/)
go vet $files
