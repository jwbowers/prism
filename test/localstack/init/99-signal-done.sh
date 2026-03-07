#!/bin/bash
#
# LocalStack Init Completion Signal
# Writes a marker file after all other init scripts have finished.
# The CI workflow waits for this file before running tests.
#

echo "=== LocalStack initialization complete ==="
echo ""
echo "All init scripts have finished:"
echo "  - 01-seed-ec2.sh: VPC, subnets, AMIs, key pair"
echo "  - 02-seed-efs.sh: EFS (Pro only, may have skipped)"
echo "  - 03-seed-ssm.sh: SSM AMI discovery parameters"
echo ""

# Write completion marker (visible on host via volume mount)
touch /var/lib/localstack/prism-init-complete
echo "Completion marker written: /var/lib/localstack/prism-init-complete"
