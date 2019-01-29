#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
RESET='\033[0m'

echo "Running go tests..."
go test -v -timeout 60s `go list ./... | grep -v vendor` | sed ''/PASS/s//$(printf "${GREEN}PASS${RESET}")/'' | sed ''/FAIL/s//$(printf "${RED}FAIL${RESET}")/''
GO_TEST_EXIT_CODE=${PIPESTATUS[0]}
if [[ $GO_TEST_EXIT_CODE -ne 0 ]]; then
    exit $GO_TEST_EXIT_CODE
fi
