#!/bin/bash

# Stop script execution if any error occurs
set -e

echo "Setting up test environment..."

# Check if Ginkgo is installed
if ! command -v ginkgo &> /dev/null; then
    echo "Ginkgo not found. Installing..."
    go install github.com/onsi/ginkgo/v2/ginkgo@latest
fi

# Check if mockgen is installed
if ! command -v mockgen &> /dev/null; then
    echo "mockgen not found. Installing..."
    go install go.uber.org/mock/mockgen@latest
fi

# Delete existing suite_test.go files
find . -name "*_suite_test.go" -type f -delete
echo "Removed existing suite_test.go files"

# Find test directories (directories containing Go files)
TEST_DIRS=$(find . -name "*.go" -not -path "*/\.*" -not -path "*/vendor/*" -not -path "*/bin/*" -not -path "*/tmp/*" | xargs dirname | sort | uniq)

# Create Ginkgo test suite for each directory
for dir in $TEST_DIRS; do
    # Check if the directory contains _test.go files
    if ls $dir/*_test.go 1> /dev/null 2>&1; then
        echo "Bootstrapping Ginkgo test suite in $dir"
        (cd $dir && ginkgo bootstrap --nodot)
    fi
done

echo "Generating mocks..."
# Continue script execution even if go generate command fails
go generate ./... || echo "Warning: Some mock generation commands failed, but continuing..."

echo "Test environment setup complete!"
