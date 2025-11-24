# Prism v0.5.14 Release Plan: Desktop Applications Foundation

**Release Date**: Target December 6, 2025 (3-4 weeks)
**Focus**: Nice DCV integration for desktop GUI applications
**Issues**: #216-#219 (To Be Created)

## 🎯 Release Goals

### Primary Objective
Enable desktop GUI applications (MATLAB, QGIS, Mathematica, Stata) through AWS Nice DCV integration, providing researchers with browser-based remote desktop access to commercial and specialized software.

**Why This Release Is Critical**:
- **Researcher Demand**: Testers have been requesting desktop GUI support
- **Commercial Software**: Unlock MATLAB, QGIS, Mathematica, Stata templates
- **Non-CLI Users**: Provides graphical interface for researchers who prefer GUI tools
- **Competitive Advantage**: Few research platforms offer seamless desktop application support
- **Enterprise Appeal**: Commercial software support critical for institutional adoption

### Success Metrics
- ✅ Generic desktop template launches successfully with MATE desktop
- ✅ Browser-based DCV connection works via SSM port forwarding
- ✅ No exposed ports (SSM-only access maintained)
- ✅ Connection time: <30 seconds from `prism connect`
- ✅ Template system supports `connection_type: "desktop"`
- ✅ Ready for v0.5.15: MATLAB, QGIS, Mathematica, Stata templates

---

## 📦 Features & Implementation

### 1. Nice DCV Architecture Documentation (#216) 📚
**Priority**: P0 (Foundation)
**Effort**: Small (3-4 hours)
**Status**: 📋 Planned

#### Deliverables
**File**: `docs/architecture/NICE_DCV_ARCHITECTURE.md`

#### Content Sections
1. **DCV Overview**:
   - What is Nice DCV (AWS remote desktop protocol)
   - Why DCV vs VNC/RDP/X11 forwarding
   - Performance characteristics (H.264 streaming, GPU support)

2. **Architecture**:
   - DCV Server on EC2 instance (port 8443)
   - SSM Session Manager for port forwarding
   - Browser-based client (no installation required)
   - Connection flow diagram

3. **Security Model**:
   - No exposed ports (DCV port 8443 only accessible via SSM)
   - Secure credential generation and storage
   - Session isolation between users

4. **Lens Project Learnings**:
   - Complete DCV implementation from Lens
   - Cloud-init scripts for automated setup
   - Connection management patterns
   - Multi-environment approach (QGIS basic, advanced, remote-sensing)

5. **Implementation Strategy**:
   - Phase 1: Generic desktop template (MATE + DCV)
   - Phase 2: Template system extension
   - Phase 3: Connection management
   - Phase 4: Application-specific templates (v0.5.15)

#### Reference Implementation
**Lens Project**: https://github.com/scttfrdmn/lens
- `apps/dcv-desktop/` - Complete DCV infrastructure
- `apps/qgis/` - Production QGIS templates (3 environments)
- `DESKTOP_APPS.md` - Architectural documentation

---

### 2. Template System Extension (#217) 🔧
**Priority**: P0 (Core Infrastructure)
**Effort**: Medium (2-3 days)
**Status**: 📋 Planned

#### Overview
Extend the template system to support desktop applications with `connection_type: "desktop"` configuration.

#### Implementation Components

**2.1 Template Schema Extension**
**File**: `pkg/templates/types.go`

Add desktop-specific fields to template schema:
```go
type Template struct {
    // ... existing fields ...

    ConnectionType string          `yaml:"connection_type"` // "web" or "desktop"
    Desktop        *DesktopConfig   `yaml:"desktop,omitempty"`
}

type DesktopConfig struct {
    Environment    string   `yaml:"environment"`     // "mate", "gnome", "xfce"
    DCVPort        int      `yaml:"dcv_port"`        // Default: 8443
    GPURequired    bool     `yaml:"gpu_required"`    // For GPU-accelerated apps
    DCVVersion     string   `yaml:"dcv_version"`     // DCV version to install
}
```

**2.2 Template Validation**
**File**: `pkg/templates/validator.go`

Add validation rules:
- Validate `connection_type` is "web" or "desktop"
- If `connection_type: "desktop"`, require `desktop` section
- Validate desktop environment is supported
- Validate DCV port (default 8443)

**2.3 Launch Integration**
**File**: `pkg/templates/provisioner.go`

Update provisioning to handle desktop templates:
- Include DCV installation in cloud-init
- Configure desktop environment
- Set up DCV server with credentials
- Add DCV port to security group (but not expose it)

#### Testing
- Unit tests for template validation
- Integration test: Launch generic desktop template
- Verify DCV server starts correctly
- Verify desktop environment loads

---

### 3. Generic Desktop Template (#219) 🖥️
**Priority**: P0 (Reference Implementation)
**Effort**: Medium (2-3 days)
**Status**: 📋 Planned

#### Overview
Create a base desktop template with MATE desktop environment and Nice DCV server. This serves as the foundation for all future desktop application templates.

#### Template Configuration
**File**: `templates/generic-desktop.yml`

```yaml
name: "Generic Desktop (MATE + Nice DCV)"
description: "Lightweight MATE desktop environment with Nice DCV remote desktop"
category: "Desktop"
connection_type: "desktop"

desktop:
  environment: "mate"
  dcv_port: 8443
  gpu_required: false
  dcv_version: "latest"

base_os:
  distribution: "ubuntu"
  version: "22.04"
  architecture: ["x86_64", "arm64"]

instance_defaults:
  size: "M"  # 4 vCPU, 16GB RAM
  instance_type: "t3.xlarge"  # Or c6g.xlarge for ARM

packages:
  - name: "ubuntu-mate-desktop"
    manager: "apt"
  - name: "firefox"
    manager: "apt"
  - name: "nice-dcv-server"
    source: "https://d1uj6qtbmh3dt5.cloudfront.net/nice-dcv-ubuntu2204-x86_64.tgz"

services:
  - name: "dcvserver"
    enabled: true
    ports:
      - 8443

users:
  - name: "researcher"
    groups: ["sudo"]
    shell: "/bin/bash"

provisioning:
  cloud_init: |
    #!/bin/bash
    set -euxo pipefail

    # Install MATE desktop
    DEBIAN_FRONTEND=noninteractive apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install -y ubuntu-mate-desktop firefox

    # Download and install Nice DCV
    cd /tmp
    wget https://d1uj6qtbmh3dt5.cloudfront.net/nice-dcv-ubuntu2204-x86_64.tgz
    tar -xvzf nice-dcv-ubuntu2204-x86_64.tgz
    cd nice-dcv-*
    apt-get install -y ./nice-dcv-server_*.deb
    apt-get install -y ./nice-dcv-web-viewer_*.deb

    # Generate DCV session password
    DCV_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-25)
    echo "$DCV_PASSWORD" > /home/researcher/.dcv_password
    chown researcher:researcher /home/researcher/.dcv_password
    chmod 600 /home/researcher/.dcv_password

    # Configure DCV server
    cat > /etc/dcv/dcv.conf <<EOF
    [security]
    authentication="system"

    [session-management]
    create-session=true

    [session-management/automatic-console-session]
    owner="researcher"
    storage-root="/home/researcher"

    [connectivity]
    enable-quic-frontend=true
    EOF

    # Start DCV server
    systemctl enable dcvserver
    systemctl start dcvserver

    # Create desktop session
    dcv create-session --owner researcher --type console researcher-session

    echo "DCV setup complete. Password stored in /home/researcher/.dcv_password"
```

#### Testing Plan
1. **Launch Test**: `prism launch "Generic Desktop (MATE + Nice DCV)" test-desktop`
2. **Verify**: DCV server running on port 8443
3. **Verify**: MATE desktop environment installed
4. **Verify**: researcher user created with password
5. **Connection Test**: Via SSM port forwarding (tested in #218)

---

### 4. DCV Connection Management (#218) 🔌
**Priority**: P0 (User Experience)
**Effort**: Medium-Large (3-4 days)
**Status**: 📋 Planned

#### Overview
Implement seamless browser-based connection to DCV desktop sessions via SSM port forwarding.

#### Implementation Components

**4.1 Connection Type Detection**
**File**: `pkg/connection/manager.go`

Update connection manager to detect desktop templates:
```go
func (m *Manager) Connect(instanceName string) error {
    instance, err := m.getInstanceInfo(instanceName)
    if err != nil {
        return err
    }

    template, err := m.getTemplateInfo(instance.Template)
    if err != nil {
        return err
    }

    switch template.ConnectionType {
    case "web":
        return m.connectWeb(instance, template)
    case "desktop":
        return m.connectDesktop(instance, template)
    default:
        return fmt.Errorf("unknown connection type: %s", template.ConnectionType)
    }
}
```

**4.2 DCV Connection Handler**
**File**: `pkg/connection/dcv.go` (new file)

Implement DCV-specific connection logic:
```go
package connection

import (
    "fmt"
    "os/exec"
    "time"
)

// ConnectDesktop establishes SSM port forwarding to DCV server and launches browser
func (m *Manager) connectDesktop(instance *types.Instance, template *types.Template) error {
    dcvPort := 8443
    if template.Desktop != nil && template.Desktop.DCVPort != 0 {
        dcvPort = template.Desktop.DCVPort
    }

    localPort := findAvailablePort() // Find unused local port

    // Start SSM port forwarding
    ssmCmd := exec.Command("aws", "ssm", "start-session",
        "--target", instance.ID,
        "--document-name", "AWS-StartPortForwardingSession",
        "--parameters", fmt.Sprintf(`{"portNumber":["%d"],"localPortNumber":["%d"]}`, dcvPort, localPort),
    )

    if err := ssmCmd.Start(); err != nil {
        return fmt.Errorf("failed to start SSM session: %w", err)
    }

    // Wait for tunnel to establish
    time.Sleep(2 * time.Second)

    // Get DCV credentials
    credentials, err := m.getDCVCredentials(instance)
    if err != nil {
        return fmt.Errorf("failed to get DCV credentials: %w", err)
    }

    // Launch browser
    dcvURL := fmt.Sprintf("https://localhost:%d", localPort)

    fmt.Printf("🖥️  Desktop Connection Established\n")
    fmt.Printf("   Instance: %s\n", instance.Name)
    fmt.Printf("   Desktop: MATE\n")
    fmt.Printf("   URL: %s\n", dcvURL)
    fmt.Printf("   Username: researcher\n")
    fmt.Printf("   Password: %s\n", credentials.Password)
    fmt.Printf("\n   Opening browser...\n")

    return openBrowser(dcvURL)
}

// getDCVCredentials retrieves DCV session password from instance
func (m *Manager) getDCVCredentials(instance *types.Instance) (*DCVCredentials, error) {
    // Use SSM RunCommand to read /home/researcher/.dcv_password
    cmd := exec.Command("aws", "ssm", "send-command",
        "--instance-ids", instance.ID,
        "--document-name", "AWS-RunShellScript",
        "--parameters", `commands=["cat /home/researcher/.dcv_password"]`,
    )

    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    // Parse command output to get password
    // (Implementation details omitted for brevity)

    return &DCVCredentials{
        Username: "researcher",
        Password: password,
    }, nil
}
```

**4.3 CLI Integration**
**File**: `internal/cli/connect.go`

Update connect command to handle desktop connections:
- Detect connection type from template
- Display appropriate connection information
- Handle SSM port forwarding lifecycle
- Show DCV credentials clearly

**4.4 GUI Integration**
**File**: `cmd/prism-gui/main.go`

Update GUI connect handler:
- Show DCV connection dialog with credentials
- Display SSM tunnel status
- Provide "Copy Password" button
- Launch browser automatically

#### Testing Plan
1. **Launch desktop template**: `prism launch generic-desktop test-desktop`
2. **Connect via CLI**: `prism connect test-desktop`
3. **Verify**: SSM tunnel established
4. **Verify**: Browser opens to DCV web client
5. **Verify**: Credentials displayed and work
6. **Verify**: MATE desktop loads in browser
7. **Test GUI**: Connect via GUI interface

---

## 🧪 Testing Strategy

### Unit Tests
- Template validation for desktop templates
- Connection type detection logic
- DCV credential parsing
- Port forwarding command generation

### Integration Tests
- Launch generic desktop template
- Verify DCV server installation
- Verify desktop environment
- Connect via CLI
- Connect via GUI
- Verify browser launch

### Manual Testing
**Test Scenarios**:
1. **Basic Desktop Launch**:
   ```bash
   prism launch generic-desktop test-desktop
   # Wait for instance to be ready (~5-8 minutes)
   prism connect test-desktop
   # Verify: Browser opens, DCV loads, MATE desktop visible
   ```

2. **Connection Workflow**:
   - Launch desktop template
   - Connect from CLI (verify credentials displayed)
   - Connect from GUI (verify connection dialog)
   - Test multiple concurrent connections
   - Test reconnection after tunnel timeout

3. **Cross-Platform Testing**:
   - Test on macOS (primary development platform)
   - Test on Linux (verify browser detection)
   - Test on Windows (if possible)

---

## 📊 Implementation Timeline

### Week 1 (Nov 13-17, 2025): Foundation
- **Day 1-2**: Create #216 documentation (NICE_DCV_ARCHITECTURE.md)
  - Review Lens implementation in detail
  - Document architecture and design decisions
  - Create connection flow diagrams

- **Day 3-5**: Implement #217 template system extension
  - Add `ConnectionType` and `DesktopConfig` to types
  - Update template validation
  - Add provisioning hooks for DCV installation

### Week 2 (Nov 18-22, 2025): Implementation
- **Day 1-2**: Create #219 generic desktop template
  - Write template YAML
  - Test cloud-init script locally
  - Verify DCV installation works

- **Day 3-5**: Implement #218 connection management
  - DCV connection handler
  - SSM port forwarding logic
  - Credential retrieval

### Week 3 (Nov 25-29, 2025): Integration & Testing
- **Day 1-2**: CLI/GUI integration
  - Update connect commands
  - Add connection dialogs
  - Browser launch logic

- **Day 3-4**: Testing and bug fixes
  - Integration tests
  - Manual testing across platforms
  - Fix connection issues

- **Day 5**: Documentation and polish
  - Update user guides
  - Connection workflow documentation
  - Prepare release notes

### Week 4 (Dec 2-6, 2025): Release
- **Day 1-2**: Final testing and validation
- **Day 3**: Create release notes and documentation
- **Day 4**: Version bump and tag
- **Day 5**: Release v0.5.14

---

## 📚 Documentation Deliverables

### Technical Documentation
1. **NICE_DCV_ARCHITECTURE.md** (#216):
   - Complete DCV architecture guide
   - Lens project learnings
   - Security model and design decisions

2. **Template System Extension** (in existing docs):
   - Update TEMPLATE_SYSTEM.md with desktop template schema
   - Document `connection_type` field
   - Desktop template examples

3. **Connection Management** (in existing docs):
   - Update CONNECTION_GUIDE.md with DCV connection workflow
   - SSM port forwarding documentation
   - Troubleshooting guide

### User-Facing Documentation
1. **Desktop Applications Guide**:
   - How to launch desktop templates
   - Connecting to desktop sessions
   - Using DCV web client
   - Common issues and solutions

2. **Template Development Guide**:
   - Creating desktop application templates
   - DCV configuration options
   - Desktop environment selection
   - GPU-accelerated applications

---

## 🎯 Success Criteria

### Functional Requirements ✅
- [ ] Generic desktop template launches successfully
- [ ] DCV server installed and configured automatically
- [ ] MATE desktop environment loads in browser
- [ ] SSM port forwarding works reliably
- [ ] Credentials displayed and functional
- [ ] Browser launches automatically
- [ ] Connection workflow under 30 seconds

### Quality Requirements ✅
- [ ] Zero compilation errors
- [ ] All unit tests passing
- [ ] Integration tests for desktop launch/connect
- [ ] Smoke tests include desktop template
- [ ] Documentation complete and accurate

### User Experience ✅
- [ ] Clear connection instructions displayed
- [ ] Credentials easy to copy
- [ ] Browser opens to correct URL
- [ ] Desktop responsive in browser
- [ ] Reconnection straightforward

---

## 🚀 Post-Release: v0.5.15 Preview

With desktop foundation in place, v0.5.15 will add application-specific templates:

### v0.5.15 Desktop Applications (December 2025)
**Issues**: #220-#223

1. **#220 - MATLAB Template**:
   - MATLAB R2024a with toolboxes
   - Cloud license integration
   - Parallel computing support

2. **#221 - QGIS Templates** (3 environments):
   - QGIS Basic (core GIS tools)
   - QGIS Advanced (plugins + processing)
   - QGIS Remote Sensing (raster analysis)

3. **#222 - Mathematica Template**:
   - Wolfram Mathematica 14
   - Notebook interface
   - Symbolic computation

4. **#223 - Stata Template**:
   - Stata 18 statistical software
   - Data analysis workflows

Each template builds on the generic desktop foundation from v0.5.14.

---

## 📝 Notes & Considerations

### Platform Support
- **macOS**: Full support (primary development platform)
- **Linux**: Full support
- **Windows**: Should work but not primary focus

### AWS Regions
- Nice DCV available in all commercial regions
- ARM instances (c6g, t4g) supported where available

### GPU Support
- GPU-accelerated DCV for future templates (not v0.5.14)
- g4dn, g5 instance types for GPU workloads
- Required for: 3D visualization, deep learning GUIs

### Lens Project Reference
The Lens project has a complete, production-tested DCV implementation:
- **Repository**: https://github.com/scttfrdmn/lens
- **Key Files**:
  - `apps/dcv-desktop/` - Complete infrastructure
  - `apps/qgis/` - Three production QGIS templates
  - `DESKTOP_APPS.md` - Architecture documentation

### Security Considerations
- No exposed ports (DCV only via SSM)
- Credentials stored securely on instance
- Session isolation between users
- HTTPS for DCV web client

---

## 🙏 Expected User Impact

### Immediate Benefits
- **Desktop GUI Support**: First-class support for graphical applications
- **Commercial Software**: Unlock MATLAB, QGIS, Mathematica, Stata
- **Non-CLI Users**: Researchers who prefer GUI tools can now use Prism
- **Browser-Based**: No client installation required

### Strategic Benefits
- **Enterprise Appeal**: Commercial software support critical for institutions
- **Competitive Advantage**: Few platforms offer seamless desktop apps
- **Template Ecosystem**: Foundation for dozens of desktop application templates
- **Tester Feedback**: Addresses directly requested feature

---

**Timeline**: December 6, 2025 (3-4 weeks)
**Status**: 📋 Planning Phase - Ready to Begin Implementation
