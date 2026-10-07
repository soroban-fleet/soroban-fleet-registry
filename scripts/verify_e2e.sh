#!/usr/bin/env bash
set -euo pipefail

echo "=== 1. Checking formatting ==="
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
  echo "The following files are not formatted:"
  echo "$UNFORMATTED"
  exit 1
fi
echo "Formatting OK"

echo "=== 2. Running go vet ==="
go vet ./...
echo "Vet OK"

echo "=== 3. Running all tests with race detector ==="
go test -v -race -p 1 ./...
echo "Tests OK"

echo "=== 4. Building binary ==="
make build
./bin/sfr --help > /dev/null
echo "Build OK"

echo "=== Verification complete! All checks passed. ==="
