#!/bin/bash

# Stop script execution if any error occurs
set -e

echo "Setting up test environment..."

# Function to handle errors
handle_error() {
    echo "ERROR: $1"
    exit 1
}

# Prepare pinned Ginkgo runner to match go.mod
GINKGO_VERSION=$(awk '/github.com\/onsi\/ginkgo\/v2[[:space:]]/{print $2; exit}' go.mod)
if [ -z "$GINKGO_VERSION" ]; then
    echo "WARNING: Could not detect Ginkgo version from go.mod; using latest for bootstrap."
    GINKGO_VERSION="latest"
fi
GINKGO_RUN="go run github.com/onsi/ginkgo/v2/ginkgo@${GINKGO_VERSION}"

# Check if mockgen is installed
if ! command -v mockgen &> /dev/null; then
    echo "mockgen not found. Installing..."
    go install go.uber.org/mock/mockgen@latest || handle_error "Failed to install mockgen"
fi

# Delete existing suite_test.go files (prune VCS and tooling dirs)
find . \
  -path "./.git" -prune -o \
  -path "./vendor" -prune -o \
  -path "./bin" -prune -o \
  -path "./tmp" -prune -o \
  -path "./venv" -prune -o \
  -path "./.gocache" -prune -o \
  -path "./.idea" -prune -o \
  -path "./.vscode" -prune -o \
  -name "*_suite_test.go" -type f -delete
echo "Removed existing suite_test.go files"

# Find test directories (directories containing Go files), prune hidden/VCS/tooling dirs
TEST_DIRS=$(find . \
  -path "./.git" -prune -o \
  -path "./vendor" -prune -o \
  -path "./bin" -prune -o \
  -path "./tmp" -prune -o \
  -path "./venv" -prune -o \
  -path "./.gocache" -prune -o \
  -path "./.idea" -prune -o \
  -path "./.vscode" -prune -o \
  -name "*.go" -print | xargs dirname | sort -u)

# Create Ginkgo test suite for each directory
for dir in $TEST_DIRS; do
    # Skip test/e2e directory as it uses standard Go testing with testify BDD
    if [[ "$dir" == "./test/e2e" ]]; then
        echo "Skipping Ginkgo bootstrap for $dir (uses standard Go testing with testify BDD)"
        continue
    fi

    # Check if the directory contains _test.go files
    if ls $dir/*_test.go 1> /dev/null 2>&1; then
        echo "Bootstrapping Ginkgo test suite in $dir"
        (cd $dir && $GINKGO_RUN bootstrap --nodot) || handle_error "Failed to bootstrap Ginkgo in $dir"
    fi
done

echo "Test environment setup complete!"
