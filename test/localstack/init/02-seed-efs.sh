#!/bin/bash
#
# LocalStack EFS Initialization Script
# Seeds LocalStack with test EFS resources
#

set -e

echo "=== Seeding LocalStack EFS resources ==="

# Configure AWS CLI to use LocalStack
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-west-2
export AWS_ENDPOINT_URL=http://localhost:4566

# Helper function for awslocal commands
aws_local() {
    aws --endpoint-url=http://localhost:4566 "$@"
}

# Load VPC configuration from EC2 seed script
if [ -f /tmp/prism-localstack-config.json ]; then
    VPC_ID=$(jq -r '.vpc_id' /tmp/prism-localstack-config.json)
    SUBNET_IDS=($(jq -r '.subnet_ids[]' /tmp/prism-localstack-config.json))
    SG_ID=$(jq -r '.security_group_id' /tmp/prism-localstack-config.json)
    echo "  Using VPC: $VPC_ID"
else
    echo "ERROR: VPC configuration not found. Run 01-seed-ec2.sh first."
    exit 1
fi

echo ""
echo "Creating test EFS file system..."

# Create EFS file system
FS_ID=$(aws_local efs create-file-system \
    --performance-mode generalPurpose \
    --throughput-mode bursting \
    --encrypted \
    --tags Key=Name,Value=prism-test-efs Key=Environment,Value=localstack \
    --query 'FileSystemId' --output text)
echo "  EFS file system created: $FS_ID"

# Wait for file system to be available
echo "  Waiting for file system to be available..."
sleep 3

# Create mount targets in each subnet
echo ""
echo "Creating EFS mount targets..."
declare -a MOUNT_TARGET_IDS=()

for subnet_id in "${SUBNET_IDS[@]}"; do
    MT_ID=$(aws_local efs create-mount-target \
        --file-system-id "$FS_ID" \
        --subnet-id "$subnet_id" \
        --security-groups "$SG_ID" \
        --query 'MountTargetId' --output text)
    MOUNT_TARGET_IDS+=("$MT_ID")
    echo "  Mount target created in subnet $subnet_id: $MT_ID"
done

echo ""
echo "=== LocalStack EFS seeding complete ==="
echo ""
echo "Resources created:"
echo "  - File System: $FS_ID"
echo "  - Mount Targets: ${MOUNT_TARGET_IDS[*]}"
echo ""

# Update configuration file with EFS details
TMP_FILE=$(mktemp)
jq --arg fs_id "$FS_ID" \
   --argjson mount_targets "$(printf '%s\n' "${MOUNT_TARGET_IDS[@]}" | jq -R . | jq -s .)" \
   '. + {efs_file_system_id: $fs_id, efs_mount_target_ids: $mount_targets}' \
   /tmp/prism-localstack-config.json > "$TMP_FILE"
mv "$TMP_FILE" /tmp/prism-localstack-config.json

echo "Updated configuration: /tmp/prism-localstack-config.json"
