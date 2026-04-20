# Nice DCV Architecture for Desktop Applications

**Status**: In Development for v0.5.14 (December 2025)
**Last Updated**: November 12, 2025
**Reference Project**: [Lens](https://github.com/scttfrdmn/lens) - Working DCV implementation
**GitHub Issues**: [#253](https://github.com/scttfrdmn/prism/issues/253), [#254](https://github.com/scttfrdmn/prism/issues/254), [#255](https://github.com/scttfrdmn/prism/issues/255), [#256](https://github.com/scttfrdmn/prism/issues/256)
**Milestone**: [v0.5.14](https://github.com/scttfrdmn/prism/milestone/33)

## Overview

This document outlines Prism's architecture for supporting desktop GUI applications (MATLAB, QGIS, Mathematica, etc.) using AWS Nice DCV (Desktop Cloud Visualization). The design is informed by learnings from the Lens project, which has a complete working DCV implementation.

## What is Nice DCV?

**NICE DCV** (Desktop Cloud Visualization) is AWS's high-performance remote desktop and application streaming protocol designed for cloud-based desktop environments.

### Key Features
- **Browser-based access**: No client installation required
- **GPU acceleration**: Supports OpenGL, DirectX, CUDA
- **High performance**: Up to 60 fps for smooth interaction
- **Low latency**: Responsive desktop experience
- **Secure**: WebSocket-based connections
- **Multi-monitor support**: Professional desktop experience
- **USB device redirection**: Peripheral support
- **Audio streaming**: Complete desktop audio

### Comparison to Alternatives

| Protocol | Performance | GPU Support | Cloud-Native | Browser Access |
|----------|-------------|-------------|--------------|----------------|
| **Nice DCV** | ⭐⭐⭐⭐⭐ | ✅ Full | ✅ Yes | ✅ Yes |
| VNC | ⭐⭐ | ❌ No | ❌ No | ⚠️ Limited |
| RDP | ⭐⭐⭐ | ⚠️ Limited | ❌ No | ❌ No |
| X11 Forwarding | ⭐ | ❌ No | ❌ No | ❌ No |

**Verdict**: Nice DCV is purpose-built for cloud computing and provides the best performance for research desktop applications.

---

## Architecture Comparison

### Current Web-Based Apps (Jupyter, RStudio, VSCode)

```
┌─────────────────────────────────────────────────────┐
│  User Browser                                       │
│  └─> http://localhost:8888                         │
└─────────────────┬───────────────────────────────────┘
                  │ HTTP/WebSocket
                  │ SSM Port Forward
┌─────────────────▼───────────────────────────────────┐
│  EC2 Instance (Ubuntu Server - No GUI)              │
│  ├─> Web Server (Jupyter/RStudio/code-server)       │
│  ├─> Port 8888/8787/8080                            │
│  └─> Process runs headless                          │
└─────────────────────────────────────────────────────┘
```

**Characteristics**:
- Applications are web-native (serve HTTP)
- No desktop environment required
- Minimal resources (2-4 vCPU, 8-16GB RAM)
- Quick launch (~2-5 minutes)
- Standard Ubuntu Server AMI

### Desktop Apps with Nice DCV (MATLAB, QGIS, etc.)

```
┌─────────────────────────────────────────────────────┐
│  User Browser                                       │
│  └─> https://localhost:8443 (DCV Web Client)       │
└─────────────────┬───────────────────────────────────┘
                  │ HTTPS/WebSocket (DCV Protocol)
                  │ SSM Port Forward
┌─────────────────▼───────────────────────────────────┐
│  EC2 Instance (Ubuntu Desktop)                      │
│  ├─> Nice DCV Server (Port 8443)                    │
│  ├─> Desktop Environment (MATE/XFCE)                │
│  ├─> X11 Display Server                             │
│  └─> Desktop Apps (QGIS, MATLAB, etc.)              │
│      └─> Render to virtual display                  │
│      └─> Stream via DCV protocol                    │
└─────────────────────────────────────────────────────┘
```

**Characteristics**:
- Full Linux desktop environment required
- Nice DCV server streams desktop to browser
- Higher resources (4-8+ vCPU, 16-32GB+ RAM)
- GPU optional (for visualization-heavy work)
- Longer launch (~5-10 minutes)
- Ubuntu Desktop or DCV-optimized AMI

---

## Technical Requirements

### AMI Requirements

**Current Apps**: Ubuntu Server 22.04 LTS (minimal)

**Desktop Apps Options**:

#### Option A: Ubuntu Desktop AMI (Recommended for Prism)
```yaml
Base: Ubuntu 22.04 Desktop
Components:
  - MATE or XFCE desktop environment
  - X11 display server
  - Nice DCV server (2023.0+)
  - OpenGL libraries
  - Desktop application (QGIS/MATLAB/etc.)
```

**Pros**:
- Full desktop environment
- Familiar to users
- Easy to customize
- Can install additional tools

**Cons**:
- Larger AMI size (~8-10GB vs 2-3GB)
- Longer launch time
- More memory usage

#### Option B: Custom DCV AMI (Future Optimization)
Pre-built AMIs with desktop + application installed.

**Pros**:
- Faster launch (app pre-installed)
- Optimized configuration
- Consistent environment

**Cons**:
- Maintenance overhead (security updates)
- Multiple AMIs to maintain
- Region distribution complexity

**Prism Strategy**: Start with Option A (Ubuntu Desktop + cloud-init), consider Option B after proven demand.

### Desktop Environment Selection

| Environment | Memory | Performance | Familiarity | Recommendation |
|-------------|--------|-------------|-------------|----------------|
| **MATE** | ~512MB | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ✅ **Recommended** |
| XFCE | ~400MB | ⭐⭐⭐⭐ | ⭐⭐⭐ | ✅ Alternative |
| GNOME | ~1GB | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ❌ Too heavy |
| KDE | ~800MB | ⭐⭐⭐ | ⭐⭐⭐⭐ | ❌ Too heavy |

**Prism Choice**: **MATE** (Ubuntu MATE flavor)
- Lightweight but full-featured
- Good Windows/macOS familiarity
- Excellent performance over DCV
- Well-maintained Ubuntu flavor

### Instance Requirements

#### Minimum Specifications (Basic Desktop)
- **vCPU**: 4 cores
- **Memory**: 16GB RAM
- **Instance**: `t3.xlarge` or `t3a.xlarge`
- **Cost**: ~$0.17/hour
- **Use**: MATLAB, basic QGIS, Mathematica

#### GPU Specifications (Visualization-Heavy)
- **vCPU**: 4 cores
- **Memory**: 16GB RAM
- **GPU**: NVIDIA T4 (16GB VRAM)
- **Instance**: `g4dn.xlarge`
- **Cost**: ~$0.53/hour
- **Use**: QGIS remote sensing, 3D visualization, Simulink

---

## Lens Project Learnings

The [Lens project](https://github.com/scttfrdmn/lens) has a complete working DCV implementation. Key learnings:

### 1. DCV Server Provisioning

**Lens Implementation**: `apps/dcv-desktop/internal/config/userdata.go`

```bash
# Cloud-init installs DCV server
# 1. Download DCV server package
wget https://d1uj6qtbmh3dt5.cloudfront.net/nice-dcv-ubuntu2204-x86_64.tgz

# 2. Extract and install
tar -xvzf nice-dcv-ubuntu2204-x86_64.tgz
cd nice-dcv-*-ubuntu2204-x86_64
sudo apt install -y ./nice-dcv-server_*.deb
sudo apt install -y ./nice-dcv-web-viewer_*.deb

# 3. Configure DCV
sudo systemctl enable dcvserver
sudo systemctl start dcvserver

# 4. Create DCV session
sudo dcv create-session --type=virtual --user ubuntu session-1

# 5. Set password
echo "ubuntu:password123" | sudo chpasswd
```

**Key Learnings**:
- DCV installation takes ~2-3 minutes
- Virtual sessions work without physical display
- Session creation must happen after user creation
- Password auth simpler than key-based for desktop

**Prism Adaptation**:
- Integrate with research user system (create sessions for research users)
- Generate secure random passwords
- Store credentials in encrypted state
- Display credentials on connection

### 2. DCV Connection Management

**Lens Implementation**: `apps/dcv-desktop/internal/cli/connect.go`

```go
// 1. Start SSM port forwarding
cmd := exec.Command("aws", "ssm", "start-session",
    "--target", instanceID,
    "--document-name", "AWS-StartPortForwardingSession",
    "--parameters", fmt.Sprintf("portNumber=8443,localPortNumber=%d", localPort))

// 2. Wait for port forwarding to be ready
time.Sleep(2 * time.Second)

// 3. Open browser to DCV web client
browser.Open(fmt.Sprintf("https://localhost:%d", localPort))

// 4. Display credentials
fmt.Printf("Username: ubuntu\n")
fmt.Printf("Password: %s\n", password)
```

**Key Learnings**:
- SSM port forwarding works perfectly with DCV
- No security group rules needed (SSM handles it)
- 2-second wait before browser open is sufficient
- Users need to see credentials before connecting

**Prism Adaptation**:
- Reuse existing SSM port forwarding infrastructure
- Add DCV-specific port (8443) handling
- Integrate with `prism workspace connect` command
- Store credentials in state for reconnection

### 3. Desktop Environment Setup

**Lens Approach**: XFCE with minimal customization

```bash
# Install desktop environment
sudo apt install -y xfce4 xfce4-goodies

# Install essential tools
sudo apt install -y firefox-esr file-roller xarchiver

# Configure auto-login (for DCV virtual sessions)
sudo mkdir -p /etc/lightdm/lightdm.conf.d
echo "[Seat:*]
autologin-user=ubuntu
autologin-user-timeout=0" | sudo tee /etc/lightdm/lightdm.conf.d/50-ubuntu.conf
```

**Key Learnings**:
- XFCE provides good balance of performance and usability
- Auto-login simplifies DCV session creation
- Basic tools (file manager, archive manager) essential
- Firefox for web access within desktop

**Prism Adaptation**:
- Use MATE instead of XFCE (slightly better familiarity)
- Keep minimal tool installation
- Auto-login for research user (not ubuntu)
- Document additional tool installation

### 4. Application-Specific Customization

**Lens QGIS Example**: `apps/qgis/`

```yaml
# Three environments with different configurations
environments:
  - name: basic-gis
    instance: t3.xlarge
    packages: [qgis, qgis-plugin-grass]

  - name: advanced-gis
    instance: t3.xlarge
    packages: [qgis, grass, saga, postgis]

  - name: remote-sensing
    instance: g4dn.xlarge  # GPU for performance
    packages: [qgis, orfeo-toolbox, snap-esa]
```

**Key Learnings**:
- Multiple environments per application valuable
- GPU matters for visualization workloads
- Package selection affects launch time significantly
- Users appreciate pre-configured environments

**Prism Adaptation**:
- Support multiple templates per desktop app
- Clear GPU vs non-GPU distinction
- Document performance characteristics
- Provide cost estimates per environment

---

## Prism Implementation Plan

### Phase 1: Template System Extension (#216)

**Goal**: Extend template system to recognize `connection_type: "desktop"`

#### Template Schema Extension

```yaml
# Existing web-based template
name: "Jupyter Notebook"
connection_type: "web"
ports:
  - 8888

# New desktop template
name: "MATLAB Workstation"
connection_type: "desktop"
desktop:
  environment: "mate"  # or "xfce", "gnome"
  dcv_port: 8443
  gpu_required: false
  gpu_drivers: []  # ["nvidia", "amd"] if GPU needed
license:
  type: "byol"  # or "cloud-based", "marketplace"
  license_server: "{{ user_config.matlab_license_server }}"  # for BYOL
  # For cloud-based: online activation (MATLAB supports this)
  # For marketplace: use pre-configured AMI
ports:
  - 8443  # DCV port
```

#### Implementation Steps

1. **Update template types** (`pkg/templates/types.go`):
```go
type ConnectionType string

const (
    ConnectionTypeWeb     ConnectionType = "web"
    ConnectionTypeDesktop ConnectionType = "desktop"
    ConnectionTypeSSH     ConnectionType = "ssh"
)

type DesktopConfig struct {
    Environment  string   `yaml:"environment"`  // "mate", "xfce", "gnome"
    DCVPort      int      `yaml:"dcv_port"`
    GPURequired  bool     `yaml:"gpu_required"`
    GPUDrivers   []string `yaml:"gpu_drivers"`
}

type Template struct {
    // ... existing fields ...
    ConnectionType ConnectionType `yaml:"connection_type"`
    Desktop        *DesktopConfig `yaml:"desktop,omitempty"`
}
```

2. **Add desktop validation** (`pkg/templates/validator.go`):
```go
func (v *Validator) validateDesktopConfig(t *Template) error {
    if t.ConnectionType != ConnectionTypeDesktop {
        return nil // Not a desktop template
    }

    if t.Desktop == nil {
        return fmt.Errorf("desktop config required for connection_type: desktop")
    }

    validEnvs := []string{"mate", "xfce", "gnome"}
    if !contains(validEnvs, t.Desktop.Environment) {
        return fmt.Errorf("invalid desktop environment: %s", t.Desktop.Environment)
    }

    if t.Desktop.DCVPort == 0 {
        t.Desktop.DCVPort = 8443 // Default
    }

    return nil
}
```

3. **AMI selection logic** (`pkg/aws/ami.go`):
```go
func (m *Manager) SelectAMI(template *Template) (string, error) {
    if template.ConnectionType == ConnectionTypeDesktop {
        // Use Ubuntu Desktop AMI instead of Server
        return m.findUbuntuDesktopAMI(template.Base)
    }
    // Existing server AMI logic
    return m.findUbuntuServerAMI(template.Base)
}

func (m *Manager) findUbuntuDesktopAMI(version string) (string, error) {
    // Search for Ubuntu MATE Desktop AMI
    input := &ec2.DescribeImagesInput{
        Filters: []types.Filter{
            {
                Name:   aws.String("name"),
                Values: []string{aws.String(fmt.Sprintf("ubuntu-mate-%s-*", version))},
            },
            {
                Name:   aws.String("architecture"),
                Values: []string{aws.String("x86_64")},
            },
        },
        Owners: []string{"099720109477"}, // Canonical
    }
    // ... find most recent AMI ...
}
```

### Phase 2: DCV Server Provisioning (#217)

**Goal**: Automated DCV installation via cloud-init

#### Cloud-Init Script Template

```bash
#!/bin/bash
set -e

# Install desktop environment
apt-get update
apt-get install -y ubuntu-mate-desktop

# Download and install DCV
cd /tmp
wget https://d1uj6qtbmh3dt5.cloudfront.net/nice-dcv-ubuntu2204-x86_64.tgz
tar -xvzf nice-dcv-ubuntu2204-x86_64.tgz
cd nice-dcv-*-ubuntu2204-x86_64
apt-get install -y ./nice-dcv-server_*.deb
apt-get install -y ./nice-dcv-web-viewer_*.deb

# Optional: Install GPU drivers
{{if .GPURequired}}
apt-get install -y nvidia-driver-535 nvidia-utils-535
{{end}}

# Configure DCV
systemctl enable dcvserver
systemctl start dcvserver

# Configure auto-login for research user
{{if .ResearchUser}}
mkdir -p /etc/lightdm/lightdm.conf.d
cat > /etc/lightdm/lightdm.conf.d/50-autologin.conf <<EOF
[Seat:*]
autologin-user={{.ResearchUser}}
autologin-user-timeout=0
EOF
{{end}}

# Create DCV session
{{if .ResearchUser}}
dcv create-session --type=virtual --user {{.ResearchUser}} session-1
{{else}}
dcv create-session --type=virtual --user ubuntu session-1
{{end}}

# Generate and store secure password
PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-20)
echo "{{.PrimaryUser}}:$PASSWORD" | chpasswd

# Store password for retrieval (encrypted)
echo "$PASSWORD" > /root/.dcv_password
chmod 600 /root/.dcv_password

# Install essential desktop tools
apt-get install -y firefox file-roller

# Application-specific packages
{{range .Packages}}
apt-get install -y {{.}}
{{end}}
```

#### Implementation

File: `pkg/dcv/provisioner.go`

```go
package dcv

import (
    "bytes"
    "text/template"
)

type ProvisionConfig struct {
    GPURequired   bool
    ResearchUser  string
    PrimaryUser   string
    Packages      []string
}

func GenerateUserData(cfg *ProvisionConfig) (string, error) {
    tmpl, err := template.New("dcv-userdata").Parse(userDataTemplate)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, cfg); err != nil {
        return "", err
    }

    return buf.String(), nil
}
```

### Phase 3: DCV Connection Management (#218)

**Goal**: Browser-based access via SSM port forwarding

#### Connection Flow

```
prism workspace connect my-desktop
    ↓
Check instance state (running?)
    ↓
Start SSM port forwarding (8443 → 8443)
    ↓
Wait for connection ready (2-3 seconds)
    ↓
Retrieve DCV credentials
    ↓
Open browser to https://localhost:8443
    ↓
Display credentials to user
```

#### Implementation

File: `pkg/connection/dcv.go`

```go
package connection

import (
    "fmt"
    "os/exec"
    "time"
)

type DCVConnection struct {
    InstanceID   string
    LocalPort    int
    RemotePort   int
    Username     string
    Password     string
}

func (c *DCVConnection) Connect() error {
    // 1. Start SSM port forwarding
    cmd := exec.Command("aws", "ssm", "start-session",
        "--target", c.InstanceID,
        "--document-name", "AWS-StartPortForwardingSession",
        "--parameters", fmt.Sprintf(`{"portNumber":["%d"],"localPortNumber":["%d"]}`,
            c.RemotePort, c.LocalPort))

    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start port forwarding: %w", err)
    }

    // 2. Wait for connection to be ready
    time.Sleep(2 * time.Second)

    // 3. Open browser
    url := fmt.Sprintf("https://localhost:%d", c.LocalPort)
    if err := browser.Open(url); err != nil {
        return fmt.Errorf("failed to open browser: %w", err)
    }

    // 4. Display credentials
    fmt.Printf("\n🖥️  DCV Desktop Connection\n\n")
    fmt.Printf("Browser opened to: %s\n\n", url)
    fmt.Printf("Login Credentials:\n")
    fmt.Printf("  Username: %s\n", c.Username)
    fmt.Printf("  Password: %s\n\n", c.Password)
    fmt.Printf("Note: Accept the self-signed certificate warning.\n")

    return nil
}
```

#### CLI Integration

File: `internal/cli/connect.go`

```go
func connectDesktop(instanceName string) error {
    // Get instance details
    instance, err := state.GetInstance(instanceName)
    if err != nil {
        return err
    }

    // Check if it's a desktop instance
    if instance.ConnectionType != types.ConnectionTypeDesktop {
        return fmt.Errorf("instance %s is not a desktop instance", instanceName)
    }

    // Retrieve DCV credentials
    creds, err := retrieveDCVCredentials(instance.ID)
    if err != nil {
        return fmt.Errorf("failed to retrieve credentials: %w", err)
    }

    // Create DCV connection
    conn := &connection.DCVConnection{
        InstanceID: instance.ID,
        LocalPort:  8443,
        RemotePort: 8443,
        Username:   creds.Username,
        Password:   creds.Password,
    }

    return conn.Connect()
}
```

### Phase 4: Generic Desktop Template (#219)

**Goal**: Base template for testing and generic desktop use

#### Template Definition

File: `templates/generic-desktop.yml`

```yaml
name: "Generic Ubuntu Desktop"
slug: "generic-desktop"
description: "Ubuntu MATE desktop with Nice DCV for generic desktop use"
long_description: "Full Ubuntu MATE desktop environment accessible via browser. Use for applications requiring GUI access."

base: "ubuntu-22.04"
connection_type: "desktop"

# Desktop configuration
desktop:
  environment: "mate"
  dcv_port: 8443
  gpu_required: false

# Complexity and categorization
complexity: "simple"
category: "Desktop"
domain: "general"

# Visual presentation
icon: "🖥️"
color: "#87CEEB"

# Instance sizing
instance_types:
  default: "t3.xlarge"  # 4 vCPU, 16GB RAM

# Basic desktop packages
package_manager: "apt"
packages:
  system:
    - ubuntu-mate-desktop
    - firefox
    - vim
    - git
    - file-roller
    - gnome-calculator

# Ports
ports:
  - port: 8443
    name: "DCV"
    description: "Nice DCV remote desktop"

# Estimated cost
cost_estimate:
  hourly: 0.17
  daily: 4.08
  monthly: 122.40

# Documentation
documentation: |
  # Generic Ubuntu Desktop

  Full Ubuntu MATE desktop environment accessible via your browser.

  ## Access
  ```bash
  prism workspace launch generic-desktop my-desktop
  prism workspace connect my-desktop
  ```

  Your browser will open to the DCV web client. Accept the self-signed
  certificate warning and log in with the displayed credentials.

  ## Use Cases
  - Testing desktop applications
  - General GUI work
  - Base for custom desktop configurations

  ## Performance
  - Instance: t3.xlarge (4 vCPU, 16GB RAM)
  - Cost: ~$0.17/hour
  - Desktop: MATE (lightweight and responsive)
```

---

## Security Considerations

### 1. Port Security
- **DCV port (8443) never exposed to internet**
- **All access via SSM port forwarding**
- **No security group ingress rules needed**
- **End-to-end encryption via DCV protocol**

### 2. Authentication
- **Strong random passwords (20+ characters)**
- **Passwords stored encrypted in state**
- **Per-session credentials**
- **Option for key-based auth (future)**

### 3. Desktop Isolation
- **Each user gets isolated DCV session**
- **No session sharing**
- **Automatic session cleanup on termination**
- **Audit logging of desktop access**

---

## Performance Characteristics

### Launch Times
- **Web apps (Jupyter, RStudio)**: 2-5 minutes
- **Desktop apps (basic)**: 5-8 minutes
- **Desktop apps (with GPU)**: 8-12 minutes

### Resource Usage
| Type | vCPU | RAM | Storage | Cost/Hour |
|------|------|-----|---------|-----------|
| Web App | 2-4 | 8-16GB | 30GB | $0.08-0.17 |
| Desktop (basic) | 4 | 16GB | 50GB | $0.17 |
| Desktop (GPU) | 4 | 16GB | 50GB | $0.53 |

### Network Bandwidth
- **Desktop idle**: 0.1-0.5 Mbps
- **Normal use**: 1-3 Mbps
- **Heavy graphics**: 5-10 Mbps
- **4K resolution**: 15-25 Mbps

---

## Future Enhancements

### Post-v0.6.2 Improvements

1. **Custom AMIs** (v0.7+)
   - Pre-built AMIs with applications installed
   - Faster launch times (2-3 min vs 5-10 min)
   - Maintenance via automated rebuilds

2. **Persistent Desktop State** (v0.7+)
   - EBS snapshots of desktop configuration
   - Resume sessions with all applications/windows
   - Cross-instance desktop profiles

3. **Multi-Monitor Support** (v0.8+)
   - DCV supports multiple displays
   - Research workflows with multiple screens
   - Professional desktop experience

4. **Collaborative Desktops** (v0.8+)
   - Multiple users sharing single desktop
   - Screen sharing for teaching
   - Collaborative debugging

5. **Windows Support** (v0.9+)
   - Windows desktop applications (if demand)
   - Commercial GIS tools (ArcGIS)
   - Microsoft ecosystem tools

---

## References

### Lens Project
- **Repository**: https://github.com/scttfrdmn/lens
- **DCV Implementation**: `apps/dcv-desktop/`
- **QGIS Example**: `apps/qgis/`
- **Documentation**: `DESKTOP_APPS.md`

### AWS Nice DCV
- **Documentation**: https://docs.aws.amazon.com/dcv/
- **Installation Guide**: https://docs.aws.amazon.com/dcv/latest/adminguide/setting-up.html
- **Web Client**: https://docs.aws.amazon.com/dcv/latest/userguide/client-web.html

### Related Issues
- #216 - DCV Template System Extension
- #217 - DCV Server Provisioning
- #218 - DCV Connection Management
- #219 - Generic Desktop Template
- #220 - MATLAB Template
- #221 - QGIS Templates
- #222 - Mathematica Template
- #223 - Stata Template

---

**Last Updated**: November 3, 2025
**Next Review**: Before v0.6.1 implementation (August 2026)
