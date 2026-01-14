#!/bin/bash
#
# LocalStack EC2 Initialization Script
# Seeds LocalStack with test EC2 resources for Prism integration tests
#

set -e

echo "=== Seeding LocalStack EC2 resources ==="

# Wait for LocalStack to be fully ready
echo "Waiting for LocalStack services..."
sleep 5

# Configure AWS CLI to use LocalStack
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-west-2
export AWS_ENDPOINT_URL=http://localhost:4566

# Helper function for awslocal commands
aws_local() {
    aws --endpoint-url=http://localhost:4566 "$@"
}

echo "Creating test VPC and networking..."

# Create VPC
VPC_ID=$(aws_local ec2 create-vpc --cidr-block 10.0.0.0/16 --query 'Vpc.VpcId' --output text)
echo "  VPC created: $VPC_ID"

# Create Internet Gateway
IGW_ID=$(aws_local ec2 create-internet-gateway --query 'InternetGateway.InternetGatewayId' --output text)
aws_local ec2 attach-internet-gateway --vpc-id "$VPC_ID" --internet-gateway-id "$IGW_ID"
echo "  Internet Gateway created: $IGW_ID"

# Create subnets in multiple availability zones
declare -a SUBNET_IDS=()
for az in us-west-2a us-west-2b us-west-2c; do
    CIDR_BLOCK="10.0.$((${#SUBNET_IDS[@]} + 1)).0/24"
    SUBNET_ID=$(aws_local ec2 create-subnet \
        --vpc-id "$VPC_ID" \
        --cidr-block "$CIDR_BLOCK" \
        --availability-zone "$az" \
        --query 'Subnet.SubnetId' --output text)
    SUBNET_IDS+=("$SUBNET_ID")
    echo "  Subnet created in $az: $SUBNET_ID"
done

# Create security group
SG_ID=$(aws_local ec2 create-security-group \
    --group-name prism-test-sg \
    --description "Prism LocalStack test security group" \
    --vpc-id "$VPC_ID" \
    --query 'GroupId' --output text)
echo "  Security group created: $SG_ID"

# Add security group rules (SSH, HTTP, HTTPS, Jupyter, RStudio)
for port in 22 80 443 8787 8888; do
    aws_local ec2 authorize-security-group-ingress \
        --group-id "$SG_ID" \
        --protocol tcp \
        --port "$port" \
        --cidr 0.0.0.0/0 >/dev/null 2>&1 || true
done
echo "  Security group rules configured"

echo ""
echo "Creating test AMIs..."

# Define test AMIs for different OS distributions
# These mock the real AMIs that Prism templates use
declare -A AMI_CONFIGS=(
    ["ubuntu-22.04-x86_64"]="Ubuntu 22.04 LTS|x86_64"
    ["ubuntu-22.04-arm64"]="Ubuntu 22.04 LTS ARM|arm64"
    ["rockylinux-9-x86_64"]="Rocky Linux 9|x86_64"
    ["rockylinux-9-arm64"]="Rocky Linux 9 ARM|arm64"
    ["debian-12-x86_64"]="Debian 12|x86_64"
    ["debian-12-arm64"]="Debian 12 ARM|arm64"
    ["amazonlinux-2023-x86_64"]="Amazon Linux 2023|x86_64"
    ["amazonlinux-2023-arm64"]="Amazon Linux 2023 ARM|arm64"
)

declare -A AMI_IDS=()

for ami_key in "${!AMI_CONFIGS[@]}"; do
    IFS='|' read -r name arch <<< "${AMI_CONFIGS[$ami_key]}"

    # Create a minimal AMI (LocalStack will mock the actual image)
    AMI_ID=$(aws_local ec2 register-image \
        --name "$name-localstack" \
        --description "LocalStack test AMI for $name ($arch)" \
        --architecture "$arch" \
        --virtualization-type hvm \
        --root-device-name /dev/sda1 \
        --block-device-mappings "[{\"DeviceName\":\"/dev/sda1\",\"Ebs\":{\"VolumeSize\":8,\"VolumeType\":\"gp3\"}}]" \
        --query 'ImageId' --output text)

    AMI_IDS[$ami_key]=$AMI_ID
    echo "  AMI created: $ami_key -> $AMI_ID ($arch)"
done

echo ""
echo "Creating test key pairs..."

# Create test SSH key pair
aws_local ec2 create-key-pair --key-name prism-test-key --query 'KeyMaterial' --output text > /tmp/prism-test-key.pem
chmod 600 /tmp/prism-test-key.pem
echo "  Key pair created: prism-test-key"

echo ""
echo "=== LocalStack EC2 seeding complete ==="
echo ""
echo "Resources created:"
echo "  - VPC: $VPC_ID"
echo "  - Internet Gateway: $IGW_ID"
echo "  - Subnets: ${SUBNET_IDS[*]}"
echo "  - Security Group: $SG_ID"
echo "  - AMIs: ${#AMI_IDS[@]} images"
echo "  - Key Pair: prism-test-key"
echo ""
echo "Configuration stored in: /tmp/prism-localstack-config.json"

# Store configuration for tests to use
cat > /tmp/prism-localstack-config.json <<EOF
{
  "vpc_id": "$VPC_ID",
  "internet_gateway_id": "$IGW_ID",
  "subnet_ids": [$(printf '"%s",' "${SUBNET_IDS[@]}" | sed 's/,$//')],
  "security_group_id": "$SG_ID",
  "ami_ids": {
$(for key in "${!AMI_IDS[@]}"; do
    echo "    \"$key\": \"${AMI_IDS[$key]}\","
done | sed '$ s/,$//')
  },
  "key_pair": "prism-test-key"
}
EOF

echo "LocalStack is ready for Prism tests!"
