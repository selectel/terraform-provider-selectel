#!/bin/bash

# Check if Go files have bad code format.
gofmt_diff="$(gofmt -d -s ./selvpc)"
if [[ -n "${gofmt_diff}" ]]; then
  echo "${gofmt_diff}"
  exit 1
fi
