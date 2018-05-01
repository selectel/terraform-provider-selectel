#!/bin/bash

# Find all Go files and check if they have wrong format of imports.
files="$(find ./selvpc -type f -name '*.go')"
goimports_diff="$(goimports -l -d $files)"
if [[ -n "${goimports_diff}" ]]; then
  echo "${goimports_diff}"
  exit 1
fi
