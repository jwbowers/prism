# Approvals Guide

The Prism approvals system lets institutions require PI (Principal Investigator) or administrator sign-off before certain resources are launched. This prevents runaway costs and enforces institutional policies on high-cost workloads.

---

## Overview

When an approval policy is in place, certain launch requests — instead of launching immediately — create an **approval request** that must be reviewed by an authorized approver. The researcher is notified when the request is approved or denied.

Common use cases:
- Large instance types (GPU instances, memory-optimized)
- Launches over a cost threshold
- Emergency budget requests
- Cross-project resource sharing

---

## For Researchers: Requesting Approval

### Requesting approval at launch

Instead of launching immediately, create an approval request:

```bash
prism workspace launch deep-learning gpu-project --request-approval
```

You'll receive a confirmation with the request ID. The request is visible in the GUI under **Approvals**.

### Launching with a pre-approved request

Once an approver approves your request, launch using the approval ID:

```bash
prism workspace launch deep-learning gpu-project --approval-id <id>
```

This bypasses the approval check since pre-approval was granted.

### Checking request status

```bash
prism approvals list --status pending     # Your pending requests
prism approvals list --status approved    # Approved requests
```

In the GUI, **Approvals** in the sidebar shows all your requests with status.

---

## For Approvers: Reviewing Requests

### Via GUI

Navigate to **Approvals** in the sidebar. The dashboard shows:
- Pending requests (badge count on sidebar item)
- Request details: requester, template, instance type, estimated cost, justification
- Approve / Deny buttons with optional notes

### Via CLI

```bash
# List pending requests
prism approvals list --status pending

# Approve a request
prism approvals approve <request-id> --note "Approved for paper deadline"

# Deny a request
prism approvals deny <request-id> --note "Please use a smaller instance type first"
```

---

## Approval Request Details

Each approval request includes:

| Field | Description |
|-------|-------------|
| Requester | User who made the request |
| Template | Workspace template requested |
| Instance type | Requested instance size/type |
| Estimated cost | Projected hourly and daily cost |
| Project | Associated project (if any) |
| Justification | Optional note from requester |
| Created | Timestamp |
| Expiry | When the approval expires if unused |

---

## Filtering the Approvals View

In the GUI, filter by status using the dropdown:

| Filter | Shows |
|--------|-------|
| Pending | Awaiting review |
| Approved | Approved (may be launched) |
| Denied | Denied requests |
| All | Complete history |

Via CLI:
```bash
prism approvals list --status all
prism approvals list --project <project-id>
```

---

## Admin: Configuring Approval Policies

Approval thresholds are configured by administrators via the **Governance** panel in the GUI (Settings → Policy Framework) or the admin CLI.

Common policy configurations:
- Require approval for instances larger than `L` size
- Require approval for estimated cost > $X/day
- Require approval for GPU instances
- Require approval for instances outside the default region

Contact your Prism administrator to configure approval policies for your institution.

---

## CLI Reference

```bash
prism approvals list                              # List your requests
prism approvals list --status pending             # Filter by status
prism approvals approve <id>                      # Approve a request (admin)
prism approvals approve <id> --note "..."         # Approve with note
prism approvals deny <id> --note "..."            # Deny with note

# At launch:
prism workspace launch <template> <name> --request-approval
prism workspace launch <template> <name> --approval-id <id>
```
