#!/bin/bash
#
# LocalStack SSM Initialization Script
# Seeds LocalStack with SSM parameters for AMI discovery
#
# Note: do NOT use set -e — LocalStack Community may restrict writes to the
# /aws/service/ path prefix. Individual failures are logged and skipped.

echo "=== Seeding LocalStack SSM parameters ==="

# Install jq if not available (LocalStack container may not have it)
if ! command -v jq &>/dev/null; then
    echo "Installing jq..."
    apt-get install -y -q jq 2>/dev/null || true
fi

# Require jq
if ! command -v jq &>/dev/null; then
    echo "ERROR: jq not available, cannot seed SSM parameters"
    exit 0  # soft exit — don't block other init
fi

# Configure AWS CLI to use LocalStack
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-west-2
export AWS_ENDPOINT_URL=http://localhost:4566

# Helper function for awslocal commands
aws_local() {
    aws --endpoint-url=http://localhost:4566 "$@"
}

# Load AMI IDs from EC2 seed script
if [ ! -f /var/lib/localstack/prism-localstack-config.json ]; then
    echo "ERROR: AMI configuration not found. Run 01-seed-ec2.sh first."
    exit 0  # soft exit
fi

echo "  Loading AMI configuration..."
echo ""
echo "Creating SSM parameters for AMI discovery..."

# Helper to put a parameter, logging any failure without aborting
put_param() {
    local name="$1" value="$2" desc="$3"
    if aws_local ssm put-parameter \
        --name "$name" \
        --value "$value" \
        --type String \
        --description "$desc" \
        --overwrite 2>&1; then
        return 0
    else
        echo "  WARNING: failed to create parameter $name (LocalStack may restrict /aws/service/ writes)"
        return 0  # non-fatal
    fi
}

# Ubuntu 22.04 AMI parameters
UBUNTU_X86_AMI=$(jq -r '.ami_ids["ubuntu-22.04-x86_64"]' /var/lib/localstack/prism-localstack-config.json)
UBUNTU_ARM_AMI=$(jq -r '.ami_ids["ubuntu-22.04-arm64"]' /var/lib/localstack/prism-localstack-config.json)

put_param "/aws/service/canonical/ubuntu/server/22.04/stable/current/amd64/hvm/ebs-gp2/ami-id" \
    "$UBUNTU_X86_AMI" "Ubuntu 22.04 LTS x86_64 AMI"
put_param "/aws/service/canonical/ubuntu/server/22.04/stable/current/arm64/hvm/ebs-gp2/ami-id" \
    "$UBUNTU_ARM_AMI" "Ubuntu 22.04 LTS ARM64 AMI"
echo "  Ubuntu 22.04 parameters attempted"

# Rocky Linux 9 AMI parameters
ROCKY_X86_AMI=$(jq -r '.ami_ids["rockylinux-9-x86_64"]' /var/lib/localstack/prism-localstack-config.json)
ROCKY_ARM_AMI=$(jq -r '.ami_ids["rockylinux-9-arm64"]' /var/lib/localstack/prism-localstack-config.json)

put_param "/aws/service/rockylinux/9/x86_64/ami-id" "$ROCKY_X86_AMI" "Rocky Linux 9 x86_64 AMI"
put_param "/aws/service/rockylinux/9/arm64/ami-id" "$ROCKY_ARM_AMI" "Rocky Linux 9 ARM64 AMI"
echo "  Rocky Linux 9 parameters attempted"

# Amazon Linux 2023 AMI parameters
AL2023_X86_AMI=$(jq -r '.ami_ids["amazonlinux-2023-x86_64"]' /var/lib/localstack/prism-localstack-config.json)
AL2023_ARM_AMI=$(jq -r '.ami_ids["amazonlinux-2023-arm64"]' /var/lib/localstack/prism-localstack-config.json)

put_param "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64" \
    "$AL2023_X86_AMI" "Amazon Linux 2023 x86_64 AMI"
put_param "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64" \
    "$AL2023_ARM_AMI" "Amazon Linux 2023 ARM64 AMI"
echo "  Amazon Linux 2023 parameters attempted"

# Debian 12 AMI parameters
DEBIAN_X86_AMI=$(jq -r '.ami_ids["debian-12-x86_64"]' /var/lib/localstack/prism-localstack-config.json)
DEBIAN_ARM_AMI=$(jq -r '.ami_ids["debian-12-arm64"]' /var/lib/localstack/prism-localstack-config.json)

put_param "/aws/service/debian/release/12/latest/amd64" "$DEBIAN_X86_AMI" "Debian 12 x86_64 AMI"
put_param "/aws/service/debian/release/12/latest/arm64" "$DEBIAN_ARM_AMI" "Debian 12 ARM64 AMI"
echo "  Debian 12 parameters attempted"

echo ""
echo "=== LocalStack SSM seeding complete ==="

# Verify parameters (informational only)
PARAM_COUNT=$(aws_local ssm describe-parameters --query 'Parameters | length(@)' --output text 2>/dev/null || echo "unknown")
echo "Total SSM parameters: $PARAM_COUNT"
