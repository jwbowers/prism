# Prism Screenshot Catalog

This file catalogs all screenshots referenced in docs and guides.
Run the app and capture each screenshot to the path shown.

## How to Capture Screenshots

### CLI Screenshots
```bash
# Build first
make build

# Start daemon (auto-starts on first CLI command)
./bin/prism workspace list

# Capture using your terminal emulator's screenshot or:
# macOS: Cmd+Shift+4 → select region
# Linux: scrot / gnome-screenshot
```

### TUI Screenshots
```bash
./bin/prism tui
# Navigate to the desired screen, then capture
```

### GUI Screenshots
```bash
./bin/prism-gui
# Navigate to desired screen, capture at 1x or 2x retina
```

---

## Scenario 1: Solo Researcher (`01-solo-researcher/`)

| File | Screen | How to Capture |
|------|--------|----------------|
| `cli-init-wizard.png` | `prism init` first-run wizard | Run `./bin/prism init` on fresh config |
| `gui-quick-start-wizard.png` | GUI Quick Start wizard (Step 1: Templates) | GUI → first launch, Quick Start wizard auto-opens |
| `gui-settings-profiles.png` | GUI Settings tab → Profiles section | GUI → Settings → Profiles |
| `gui-workspaces-list.png` | GUI Workspaces tab with running instance | GUI → Workspaces (with a running workspace) |
| `gui-storage-management.png` | GUI Storage tab showing EFS/EBS sections | GUI → Storage |
| `gui-projects-dashboard.png` | GUI Projects tab | GUI → Projects |

---

## Scenario 2: Lab Environment (`02-lab-environment/`)

| File | Screen | How to Capture |
|------|--------|----------------|
| `gui-quick-start-wizard.png` | GUI Quick Start wizard | GUI → first launch |
| `gui-settings-profiles.png` | Settings → Profiles | GUI → Settings |
| `gui-workspaces-list.png` | Workspaces list | GUI → Workspaces |
| `gui-storage-management.png` | Storage tab | GUI → Storage |
| `gui-projects-dashboard.png` | Projects dashboard with budget bars | GUI → Projects (with projects created) |

---

## Scenario 3: University Class (`03-university-class/`)

| File | Screen | How to Capture |
|------|--------|----------------|
| `gui-quick-start-wizard.png` | Quick Start wizard | GUI → first launch |
| `gui-settings-profiles.png` | Settings → Profiles | GUI → Settings |
| `gui-workspaces-list.png` | Workspaces list | GUI → Workspaces |
| `gui-storage-management.png` | Storage tab | GUI → Storage |
| `gui-projects-dashboard.png` | Projects → budget tracking per student | GUI → Projects |

---

## Scenario 4: Conference Workshop (`04-conference-workshop/`)

| File | Screen | How to Capture |
|------|--------|----------------|
| `gui-quick-start-wizard.png` | Quick Start wizard | GUI → first launch |
| `gui-settings-profiles.png` | Settings → Profiles | GUI → Settings |
| `gui-workspaces-list.png` | Workspaces list | GUI → Workspaces |
| `gui-storage-management.png` | Storage tab | GUI → Storage |
| `gui-projects-dashboard.png` | Projects → invitations stats | GUI → Projects |

---

## Scenario 5: Cross-Institutional (`05-cross-institutional/`)

| File | Screen | How to Capture |
|------|--------|----------------|
| `gui-quick-start-wizard.png` | Quick Start wizard | GUI → first launch |
| `gui-settings-profiles.png` | Settings → Profiles | GUI → Settings |
| `gui-workspaces-list.png` | Workspaces list | GUI → Workspaces |
| `gui-storage-management.png` | Storage tab | GUI → Storage |
| `gui-projects-dashboard.png` | Projects dashboard | GUI → Projects |

---

## Generic CLI Screens (`cli/`)

| File | Command | Notes |
|------|---------|-------|
| `workspace-list.png` | `prism workspace list` | Shows running/stopped workspaces |
| `workspace-launch.png` | `prism workspace launch python-ml my-project` | Launch output |
| `templates-list.png` | `prism templates` | Template catalog |
| `storage-list.png` | `prism storage list` | EFS volumes |
| `budget-status.png` | `prism budget list` | Budget overview |
| `init-wizard.png` | `prism init` | First-run wizard |

## Generic TUI Screens (`tui/`)

| File | Screen | How to Capture |
|------|--------|----------------|
| `dashboard.png` | TUI Dashboard tab (key: 1) | `prism tui` → Dashboard |
| `workspaces.png` | TUI Workspaces tab (key: 2) | `prism tui` → Workspaces |
| `templates.png` | TUI Templates tab (key: 3) | `prism tui` → Templates |
| `storage.png` | TUI Storage tab (key: 4) | `prism tui` → Storage |
| `settings.png` | TUI Settings tab | `prism tui` → Settings |

## Generic GUI Screens (`gui/`)

| File | Screen | How to Capture |
|------|--------|----------------|
| `home-dashboard.png` | GUI home/dashboard | GUI → Home |
| `workspaces-running.png` | GUI workspaces with running instance | GUI → Workspaces |
| `templates-gallery.png` | GUI template gallery with filters | GUI → Templates |
| `storage-tabs.png` | GUI storage with EFS/EBS tabs | GUI → Storage |
| `settings-profiles.png` | GUI settings/profiles | GUI → Settings |
| `launch-wizard.png` | GUI launch wizard | GUI → Launch |

---

## Screenshot Spec

- **Format**: PNG
- **Resolution**: 1440×900 minimum; 2880×1800 preferred (2x retina)
- **Window state**: Clean state — no personal data, demo credentials visible
- **Naming**: kebab-case, descriptive, matches the `src` path in the docs
- **Background**: Use terminal with dark theme for CLI/TUI; system default for GUI

## Adding to Docs

Screenshots are referenced with relative paths from the doc location:
```markdown
![Alt text](images/01-solo-researcher/gui-workspaces-list.png)
*Caption explaining what the screenshot shows.*
```
