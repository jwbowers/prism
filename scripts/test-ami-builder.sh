#!/bin/bash
# Test script for the AMI builder integration tests
# Uses Substrate for in-process AWS emulation (no Docker needed for unit tests)

set -e

# Run AMI builder integration tests
echo "Running AMI builder integration tests..."
INTEGRATION_TESTS=1 go test -v -tags=integration ./pkg/ami

# Run AWS package integration tests (Substrate-based, no Docker needed)
echo "Running AWS package Substrate integration tests..."
go test -v -tags=substrate ./pkg/aws

# Collect and display test coverage
echo "Collecting test coverage..."
INTEGRATION_TESTS=1 go test -v -tags=integration -coverprofile=ami_coverage.out ./pkg/ami
go tool cover -func=ami_coverage.out | grep "total"

# Generate HTML coverage report
go tool cover -html=ami_coverage.out -o ami_coverage.html
echo "Coverage report generated: ami_coverage.html"

echo "AMI builder tests completed successfully!"
