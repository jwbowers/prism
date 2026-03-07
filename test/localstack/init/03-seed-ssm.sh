#!/bin/bash
#
# LocalStack SSM Initialization Script
# Seeds LocalStack with SSM parameters for AMI discovery
#

set -e

echo "=== Seeding LocalStack SSM parameters ==="

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
if [ -f /var/lib/localstack/prism-localstack-config.json ]; then
    echo "  Loading AMI configuration..."
else
    echo "ERROR: AMI configuration not found. Run 01-seed-ec2.sh first."
    exit 1
fi

echo ""
echo "Creating SSM parameters for AMI discovery..."

# Ubuntu 22.04 AMI parameters (commonly used by Prism)
UBUNTU_X86_AMI=$(jq -r '.ami_ids["ubuntu-22.04-x86_64"]' /var/lib/localstack/prism-localstack-config.json)
UBUNTU_ARM_AMI=$(jq -r '.ami_ids["ubuntu-22.04-arm64"]' /var/lib/localstack/prism-localstack-config.json)

aws_local ssm put-parameter \
    --name "/aws/service/canonical/ubuntu/server/22.04/stable/current/amd64/hvm/ebs-gp2/ami-id" \
    --value "$UBUNTU_X86_AMI" \
    --type String \
    --description "Ubuntu 22.04 LTS x86_64 AMI" \
    --overwrite >/dev/null

aws_local ssm put-parameter \
    --name "/aws/service/canonical/ubuntu/server/22.04/stable/current/arm64/hvm/ebs-gp2/ami-id" \
    --value "$UBUNTU_ARM_AMI" \
    --type String \
    --description "Ubuntu 22.04 LTS ARM64 AMI" \
    --overwrite >/dev/null

echo "  Ubuntu 22.04 parameters created"

# Rocky Linux 9 AMI parameters
ROCKY_X86_AMI=$(jq -r '.ami_ids["rockylinux-9-x86_64"]' /var/lib/localstack/prism-localstack-config.json)
ROCKY_ARM_AMI=$(jq -r '.ami_ids["rockylinux-9-arm64"]' /var/lib/localstack/prism-localstack-config.json)

aws_local ssm put-parameter \
    --name "/aws/service/rockylinux/9/x86_64/ami-id" \
    --value "$ROCKY_X86_AMI" \
    --type String \
    --description "Rocky Linux 9 x86_64 AMI" \
    --overwrite >/dev/null

aws_local ssm put-parameter \
    --name "/aws/service/rockylinux/9/arm64/ami-id" \
    --value "$ROCKY_ARM_AMI" \
    --type String \
    --description "Rocky Linux 9 ARM64 AMI" \
    --overwrite >/dev/null

echo "  Rocky Linux 9 parameters created"

# Amazon Linux 2023 AMI parameters
AL2023_X86_AMI=$(jq -r '.ami_ids["amazonlinux-2023-x86_64"]' /var/lib/localstack/prism-localstack-config.json)
AL2023_ARM_AMI=$(jq -r '.ami_ids["amazonlinux-2023-arm64"]' /var/lib/localstack/prism-localstack-config.json)

aws_local ssm put-parameter \
    --name "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64" \
    --value "$AL2023_X86_AMI" \
    --type String \
    --description "Amazon Linux 2023 x86_64 AMI" \
    --overwrite >/dev/null

aws_local ssm put-parameter \
    --name "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64" \
    --value "$AL2023_ARM_AMI" \
    --type String \
    --description "Amazon Linux 2023 ARM64 AMI" \
    --overwrite >/dev/null

echo "  Amazon Linux 2023 parameters created"

# Debian 12 AMI parameters
DEBIAN_X86_AMI=$(jq -r '.ami_ids["debian-12-x86_64"]' /var/lib/localstack/prism-localstack-config.json)
DEBIAN_ARM_AMI=$(jq -r '.ami_ids["debian-12-arm64"]' /var/lib/localstack/prism-localstack-config.json)

aws_local ssm put-parameter \
    --name "/aws/service/debian/release/12/latest/amd64" \
    --value "$DEBIAN_X86_AMI" \
    --type String \
    --description "Debian 12 x86_64 AMI" \
    --overwrite >/dev/null

aws_local ssm put-parameter \
    --name "/aws/service/debian/release/12/latest/arm64" \
    --value "$DEBIAN_ARM_AMI" \
    --type String \
    --description "Debian 12 ARM64 AMI" \
    --overwrite >/dev/null

echo "  Debian 12 parameters created"

echo ""
echo "=== LocalStack SSM seeding complete ==="
echo ""
echo "SSM Parameters created:"
echo "  - Ubuntu 22.04 (x86_64 & ARM64)"
echo "  - Rocky Linux 9 (x86_64 & ARM64)"
echo "  - Amazon Linux 2023 (x86_64 & ARM64)"
echo "  - Debian 12 (x86_64 & ARM64)"
echo ""

# Verify parameters
PARAM_COUNT=$(aws_local ssm describe-parameters --query 'Parameters | length(@)' --output text)
echo "Total SSM parameters: $PARAM_COUNT"
