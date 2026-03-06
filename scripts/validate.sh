#!/bin/bash

# Prism validation pipeline
# Runs comprehensive validation of the entire project

set -e

echo "🔧 Prism Validation Pipeline"
echo "======================================="

# Color functions
red() { echo -e "\033[31m$*\033[0m"; }
green() { echo -e "\033[32m$*\033[0m"; }
yellow() { echo -e "\033[33m$*\033[0m"; }
blue() { echo -e "\033[34m$*\033[0m"; }

# Step counter
step=1
total_steps=7

print_step() {
    blue "[$step/$total_steps] $1"
    ((step++))
}

# Step 1: Version synchronization check
print_step "Validating version synchronization..."
if ./scripts/validate-versions.sh; then
    green "    ✓ Version numbers synchronized"
else
    red "    ✗ Version mismatch detected"
    exit 1
fi

# Step 2: Go environment validation
print_step "Validating Go environment..."
go version
go env GOPATH
go env GOROOT

# Step 3: Dependency validation
print_step "Validating dependencies..."
go mod verify
go mod tidy -diff

# Step 4: Code quality checks
print_step "Running code quality checks..."
echo "  - Formatting..."
if ! go fmt ./... | grep -q "^"; then
    green "    ✓ Code is properly formatted"
else
    red "    ✗ Code needs formatting. Run 'make fmt'"
    exit 1
fi

echo "  - Vetting..."
if go vet ./...; then
    green "    ✓ Go vet passed"
else
    red "    ✗ Go vet failed"
    exit 1
fi

# Step 5: Build validation
print_step "Validating build..."
if make clean && make build; then
    green "    ✓ Build successful"
else
    red "    ✗ Build failed"
    exit 1
fi

# Step 6: Test validation
print_step "Running test suite..."
if make test-unit; then
    green "    ✓ Unit tests passed"
else
    red "    ✗ Unit tests failed"
    exit 1
fi

# Step 7: Binary validation
print_step "Validating binaries..."
if [ -f "bin/prism" ] && [ -f "bin/prismd" ]; then
    green "    ✓ Core binaries created"
    echo "    - $(file bin/prism)"
    echo "    - $(file bin/prismd)"
    
    if [ -f "bin/prism-gui" ]; then
        green "    ✓ GUI binary created"
        echo "    - $(file bin/prism-gui)"
    else
        yellow "    ! GUI binary not found (acceptable for headless builds)"
    fi
else
    red "    ✗ Missing core binaries"
    exit 1
fi

echo ""
green "🎉 Prism validation completed successfully!"
echo ""
echo "Summary:"
echo "  ✅ Version sync: Consistent"
echo "  ✅ Go environment: Valid"
echo "  ✅ Dependencies: Verified"
echo "  ✅ Code quality: Passed"
echo "  ✅ Build: Successful"
echo "  ✅ Tests: Passed"
echo "  ✅ Binaries: Created"
echo ""
echo "Ready for development or deployment!"