# Workshops Guide

Prism Workshops are designed for one-time events: conferences, bootcamps, hackathons, and training sessions. Unlike Courses (semester-long), workshops are time-bounded and use enrollment tokens for quick participant onboarding.

---

## Overview

A **Workshop** provides:
- **Enrollment tokens**: distribute a link or code — participants redeem it to get instant access
- **Lifecycle management**: start and end the workshop; all participant workspaces are managed together
- **Configuration templates**: consistent workspace configuration across all participants
- **Dashboard view**: see all participant workspaces and their status at a glance

---

## Creating a Workshop

### Via GUI

1. Navigate to **Workshops** in the sidebar
2. Click **Create Workshop**
3. Fill in: name, description, date/time, max participants
4. Select a workspace template
5. Click **Create**

### Via CLI

```bash
prism workshop create \
  --name "ML Bootcamp 2026" \
  --description "2-day hands-on machine learning workshop" \
  --template python-ml \
  --max-participants 30
```

---

## Onboarding Participants

### Create an enrollment token

In the GUI: Workshop → **Create Enrollment Token**

Or via CLI:
```bash
prism workshop tokens create <workshop-id> \
  --name "Participant Token" \
  --max-uses 30
```

Share the token with participants. They redeem it with:
```bash
prism invitation redeem <token>
```

### Manual enrollment

```bash
prism workshop members enroll <workshop-id> \
  --email participant@university.edu
```

---

## Managing the Workshop Lifecycle

### Starting a workshop

When you're ready, activate the workshop so participants can launch workspaces:

```bash
prism workshop start <workshop-id>
```

In the GUI, use the **Start Workshop** button on the workshop detail page.

### Dashboard view

The workshop dashboard shows all participant workspaces in real time:
- Workspace state (running / stopped / pending)
- Launch progress
- Any errors

### Ending a workshop

```bash
prism workshop end <workshop-id>
```

This broadcasts a warning to all participant workspaces, then stops them all. Participant data is preserved (stopped, not terminated) unless `--terminate` is specified.

---

## Configuration Templates

Save common workshop configurations for reuse:

```bash
# Save current workshop config as a reusable template
prism workshop config save <workshop-id> "ml-bootcamp-config"

# Apply a saved config to a new workshop
prism workshop config apply <workshop-id> "ml-bootcamp-config"
```

---

## CLI Reference

```bash
prism workshop                                    # List your workshops
prism workshop create                             # Create a workshop (interactive)
prism workshop list                               # List workshops you manage
prism workshop start <id>                         # Activate for participants
prism workshop end <id>                           # End and stop workspaces
prism workshop members list <id>                  # List participants
prism workshop members enroll <id>                # Enroll a participant
prism workshop tokens create <id>                 # Create enrollment token
prism workshop config save <id> <name>            # Save config template
prism workshop config apply <id> <name>           # Apply saved config
```
