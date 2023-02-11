#!/usr/bin/env bash

echo "==> Running golangci-lint..."
golangci-lint run ./...
if [[ $? -ne 0 ]]; then
    echo ""
    echo "Golangci-lint found suspicious constructs."
    exit 1
fi

exit 0
