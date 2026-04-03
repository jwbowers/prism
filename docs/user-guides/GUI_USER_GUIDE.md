# GUI Guide

The Prism desktop application provides a visual interface for managing workspaces, storage, and settings without typing commands.

## Starting the GUI

```bash
prism gui
```

Or launch Prism from your Applications folder (macOS) or Start menu (Windows).

---

## Main views

### Dashboard

The Dashboard shows all your workspaces at a glance:

- Status indicators (running, stopped, hibernated)
- Real-time cost per instance
- Quick-action buttons: Launch, Stop, Hibernate, Connect
- Profile and region in the sidebar

### Template gallery

Browse available research environments:

- Visual cards with descriptions and included software
- One-click launch
- Size selection (XS to XL)

### Instance manager

Manage all workspaces from one screen:

- Start and stop with a click
- Connect buttons that open SSH or browser connections
- Color-coded status indicators
- Cost and usage details

### Storage manager

Create and manage EFS and EBS volumes:

- Create volumes with a size slider
- Attach/detach volumes to running workspaces
- Usage indicators

---

## System tray / menu bar

The GUI runs in your system tray (Windows) or menu bar (macOS) so you can:

- Monitor workspace status at a glance
- Receive notifications about running workspaces
- Access Prism without opening a terminal

---

## Profile management

Manage multiple AWS accounts from one place:

1. Open Settings → Profile Management
2. **Add a personal profile** — for your own AWS account
3. **Add an invitation profile** — when a collaborator invites you to their account
4. **Switch profiles** — click the profile name in the sidebar

When you switch profiles, the GUI refreshes to show workspaces in that AWS account.

---

## Launching a workspace

1. Click **Launch New Instance**
2. Select a template from the gallery
3. Enter a name
4. Choose a size (XS to XL)
5. Click **Launch**

The workspace appears in the instance list and becomes ready in about 2 minutes.

---

## Connecting to a workspace

1. Find the workspace in the Instances list
2. Click **Connect**
3. Copy the SSH command, or click a web service link (Jupyter, RStudio)

---

## Dark and light themes

The GUI follows your system appearance setting automatically, or you can switch manually in Settings.

---

## Getting help

Report issues: [github.com/scttfrdmn/prism/issues](https://github.com/scttfrdmn/prism/issues)

Community discussions: [github.com/scttfrdmn/prism/discussions](https://github.com/scttfrdmn/prism/discussions)
