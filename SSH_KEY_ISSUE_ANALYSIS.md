# SSH Key Management Issue - Root Cause Analysis

## Problem Statement

SSH access fails after automatic key detection and configuration. The system selected key "cws-aws-default-key" but the local private key doesn't match the public key stored in AWS.

## Symptoms

```bash
# Instance launched with key name
aws ec2 describe-instances --filters "Name=tag:Name,Values=test-r-stack-2" \
  --query 'Reservations[0].Instances[0].KeyName'
# Output: cws-aws-default-key

# SSH fails with permission denied
ssh -i ~/.ssh/cws-aws-default-key ubuntu@52.12.46.245
# Error: ubuntu@52.12.46.245: Permission denied (publickey).

# Fingerprint mismatch
# AWS Key Fingerprint: dc:4d:f4:01:50:0a:e0:33:b5:ba:b2:bc:6e:72:1a:8a
# Local Key Fingerprint: 47:0f:ca:d4:12:b9:3e:b2:98:7d:e8:25:06:c7:2b:80
```

## Root Cause

The SSH key management system has a critical flaw:

1. `setupSSHKeyForLaunch()` (pkg/daemon/instance_handlers.go:935) gets SSH key from profile
2. `GetSSHKeyForProfile()` returns a key name like "cws-aws-default-key"
3. `ensureSSHKeyInAWS()` attempts to upload the public key
4. **BUG**: If a key with that name already exists in AWS, `EnsureKeyPairExists()` doesn't overwrite it
5. Instance launches with existing AWS key that doesn't match local private key
6. SSH fails because we don't have the matching private key

## Code Flow

```
pkg/daemon/instance_handlers.go:234-239
├── setupSSHKeyForLaunch(req)
│   ├── profileManager.GetCurrentProfile()
│   ├── sshKeyManager.GetSSHKeyForProfile(profile)
│   │   └── Returns: (keyPath, keyName, nil)
│   └── req.SSHKeyName = keyName  // e.g., "cws-aws-default-key"
│
├── ensureSSHKeyInAWS(awsManager, req) [Line 265]
│   ├── sshKeyManager.GetPublicKeyContent(publicKeyPath)
│   └── awsManager.EnsureKeyPairExists(keyName, publicKeyContent)
│       └── PROBLEM: Doesn't overwrite if key exists!
│
└── awsManager.LaunchInstance(req)
    └── Uses existing (mismatched) AWS key
```

## Impact

- **User Experience**: "SSH access should just work" - it doesn't
- **Security**: Users can't access instances they've created
- **Debugging**: Difficult to diagnose (requires fingerprint comparison)
- **Workaround**: Manual key management required

## Solutions

### Option 1: Force Key Update (Recommended)
Modify `EnsureKeyPairExists()` to always update the key content:
```go
// pkg/aws/key_pairs.go
func (m *Manager) EnsureKeyPairExists(keyName, publicKeyContent string) error {
    // Check if key exists
    exists, err := m.keyPairExists(keyName)
    if err != nil {
        return err
    }

    if exists {
        // Delete old key
        if err := m.DeleteKeyPair(keyName); err != nil {
            return fmt.Errorf("failed to delete existing key: %w", err)
        }
    }

    // Import new key
    return m.ImportKeyPair(keyName, publicKeyContent)
}
```

### Option 2: Use Unique Key Names
Generate unique key names per profile/timestamp:
```go
// pkg/profile/ssh_keys.go
func (m *SSHKeyManager) generateKeyName(profile *Profile) string {
    timestamp := time.Now().Unix()
    return fmt.Sprintf("cws-%s-%d-key", safeName, timestamp)
}
```
**Problem**: Leaves orphaned keys in AWS

### Option 3: Fingerprint Verification
Verify fingerprint matches before using existing key:
```go
func (m *Manager) ensureSSHKeyInAWS(...) error {
    exists, awsFingerprint := m.getKeyPairFingerprint(keyName)
    localFingerprint := m.calculateFingerprint(publicKeyContent)

    if exists && awsFingerprint != localFingerprint {
        // Delete and recreate
        m.DeleteKeyPair(keyName)
    }

    return m.EnsureKeyPairExists(keyName, publicKeyContent)
}
```

## Testing Needed

1. **Fresh Launch**: Launch with no existing AWS keys
2. **Existing Key**: Launch when key name already exists in AWS
3. **Multiple Profiles**: Verify each profile gets correct key
4. **Key Rotation**: Update local key, verify AWS syncs
5. **Fingerprint Match**: Verify local and AWS keys always match

## Workaround for Users

Until fixed, users must manually configure SSH keys:

```bash
# Option 1: Use explicit key flag
prism workspace launch template name --ssh-key my-existing-key

# Option 2: Delete conflicting AWS key first
aws ec2 delete-key-pair --key-name cws-aws-default-key
# Then launch (will create fresh key)
prism workspace launch template name

# Option 3: Use EC2 Instance Connect (temporary)
aws ec2-instance-connect send-ssh-public-key \
  --instance-id i-xxx \
  --instance-os-user ubuntu \
  --ssh-public-key file://~/.ssh/id_rsa.pub
```

## Next Steps

1. **Immediate**: Document issue in GitHub (Issue #XXX)
2. **Short-term**: Implement Option 1 (force key update)
3. **Medium-term**: Add fingerprint verification
4. **Long-term**: Improve profile SSH key management UX

## Commits Related

- 611108ff9: Added --ssh-key flag (exposed the issue)
- Previous: SSH key management infrastructure (has the bug)

## Testing This Template

The r-research-full-stack template is fully tested and documented in:
- docs/templates/R_RESEARCH_FULL_STACK.md (technical guide)
- docs/templates/R_FULL_STACK_QUICK_START.md (user guide)

Once SSH key management is fixed, all components will be testable.
