#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running tests for Pizzeria application${NC}"

# Function to run tests with color output
run_test() {
    echo -e "${YELLOW}Running tests in $1${NC}"
    if go test -v $1; then
        echo -e "${GREEN}✓ Tests in $1 passed${NC}"
        return 0
    else
        echo -e "${RED}✗ Tests in $1 failed${NC}"
        return 1
    fi
}

# Change to project root directory
cd "$(dirname "$0")/.."

# Check for required packages
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: go is not installed${NC}"
    exit 1
fi

# Ensure all dependencies are installed
echo -e "${YELLOW}Ensuring dependencies are installed...${NC}"
go mod tidy

# Run all tests
echo -e "${YELLOW}Running all tests...${NC}"

# Create an array to collect failed packages
FAILED=()

# Run unit tests for each package
for pkg in $(find . -name "*_test.go" -not -path "*/node_modules/*" | xargs dirname | sort -u); do
    if ! run_test "${pkg#./}"; then
        FAILED+=("${pkg#./}")
    fi
done

# Check if any tests failed
if [ ${#FAILED[@]} -ne 0 ]; then
    echo -e "${RED}The following packages had failing tests:${NC}"
    for pkg in "${FAILED[@]}"; do
        echo -e "  - ${RED}$pkg${NC}"
    done
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
fi

# Run tests with race detection (optional)
echo -e "${YELLOW}Would you like to run tests with race detection? This may take longer. (y/n)${NC}"
read -r run_race

if [[ $run_race == "y" ]]; then
    echo -e "${YELLOW}Running tests with race detection...${NC}"
    if go test -race ./...; then
        echo -e "${GREEN}✓ No race conditions detected${NC}"
    else
        echo -e "${RED}✗ Race conditions detected${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}Testing complete${NC}"