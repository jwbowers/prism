# Prism GUI Architecture

> **Updated in v0.31.0**: The GUI was rebuilt from shadcn/ui to shadcn/ui with Tailwind CSS. This document reflects the current architecture as of v0.35.3.

## Overview

The Prism GUI is a desktop application built with Go + [Wails v3](https://v3.wails.io/) that provides a clean interface for managing cloud research environments. The frontend is a React single-page application; Wails embeds it in a native window with a Go backend bridge.

## Technology Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Desktop shell | Wails v3 | v3.0.0-alpha.74 |
| Frontend framework | React | 19.1 |
| Language | TypeScript | 5.9 |
| Component library | shadcn/ui (Radix UI primitives) | code-owned |
| Styling | Tailwind CSS v4 | 4.2 |
| Animations | Framer Motion | 12.x |
| Toast notifications | Sonner | latest |
| Icons | Lucide React | latest |
| Terminal emulator | xterm.js | 5.5 |
| Font | Atkinson Hyperlegible Next | variable |

### Component pattern: shadcn/ui

Prism uses the [shadcn/ui](https://ui.shadcn.com/) approach: components are copied into the codebase under `src/components/ui/` and owned by the project. They use [Radix UI](https://radix-ui.com/) primitives for accessibility (focus management, keyboard navigation, ARIA) and Tailwind for styling.

This means components are fully customizable — no version lock-in to an external component library's design decisions.

### Cloudscape compatibility shim

`src/lib/cloudscape-shim.tsx` provides a compatibility layer that maps Cloudscape component APIs to shadcn/ui implementations. This was used during the v0.31.0 migration and remains as a bridge while views are incrementally migrated. New views use shadcn/ui directly.

### SafePrismAPI pattern

`src/lib/api.ts` exposes a `SafePrismAPI` class with named methods (never generic `get()`/`post()`). Each method calls the private `safeRequest()` helper which handles auth headers, error wrapping, and base URL. This keeps the API surface explicit and type-safe.

## Design Philosophy

### Single-Page Application (SPA)
- **No popup windows** — all interactions happen within the main window
- **Toast notifications** — feedback via Sonner (bottom-right)
- **Content switching** — navigation changes the main content area
- **Consistent layout** — shadcn/ui `Sidebar` + `SidebarInset` pattern

### Modern Visual Design
- **Card-based layouts** for organized information presentation
- **Grid systems** for consistent spacing and alignment
- **Typography hierarchy** with proper bold headers and text styling
- **Icon integration** using web-based iconography and CSS styling
- **Visual status indicators** with color-coded state icons

## Architecture Components

### Main Application Layout

```
┌─────────────┬──────────────────────────────────┐
│   Sidebar   │           Main Content           │
│  (20% width)│          (80% width)             │
├─────────────┼──────────────────────────────────┤
│ Navigation  │   ┌─ Notifications (inline) ─┐   │
│ (v0.5.9+)   │   │ ✅ Success/Error alerts   │   │
│             │   │ ℹ️  Info messages         │   │
│ 🏠 Dashboard│   │ ❌ Error notifications    │   │
│ 📋 Templates│   └──────────────────────────┘   │
│ 💻 Workspaces│                                  │
│ 🖥️ Terminal │     📊 Dynamic Content Area      │
│ 🌐 Web Svc  │     ┌─────────────────────────┐   │
│ ─────────── │     │  Dashboard / Templates  │   │
│ 💾 Storage  │     │  Workspaces / Projects  │   │
│ 📊 Projects │     │  Storage / Settings     │   │
│ ─────────── │     └─────────────────────────┘   │
│ ⚙️ Settings │                                  │
│   + Advanced│                                  │
└─────────────┴──────────────────────────────────┘
```

### Settings Internal Navigation (v0.5.9+)

When Settings is selected, the main content area includes a side navigation:

```
┌─────────────┬─────────────┬────────────────────┐
│Main Nav     │Settings Nav │  Settings Content  │
├─────────────┼─────────────┼────────────────────┤
│ ...         │ General     │ ┌────────────────┐ │
│ ⚙️ Settings │ Profiles    │ │ System Status  │ │
│   (active)  │ Users       │ │ Configuration  │ │
│ ...         │ ─────────── │ │ AWS Settings   │ │
│             │ ▶ Advanced  │ │ Feature Mgmt   │ │
│             │   • AMI     │ │ Debug Tools    │ │
│             │   • Sizing  │ └────────────────┘ │
│             │   • Policy  │                    │
│             │   • Market  │                    │
│             │   • Idle    │                    │
│             │   • Logs    │                    │
└─────────────┴─────────────┴────────────────────┘
```

## Navigation Sections

### Main Navigation (9 Items - v0.5.9+)

### 🏠 Dashboard (Primary)
**Purpose**: Overview and quick actions
**Features**:
- Overview cards (active workspaces, daily cost, totals)
- Quick launch form with template/name/size selection
- Recent workspaces list with management shortcuts
- Real-time cost and status updates

### 📋 Templates
**Purpose**: Research environment template discovery and launching
**Features**:
- Visual template gallery with descriptions and badges
- Pre-configured environment details
- AMI-optimized and script-based templates
- One-click template launching
- Template filtering and search

### 💻 My Workspaces
**Purpose**: Complete workspace management
**Features**:
- Detailed workspace cards with full information
- State-aware action buttons (Connect/Start/Stop/Hibernate/Delete)
- Visual status indicators with color coding
- Launch new workspace shortcut
- Connection information and SSH access

### 🖥️ Terminal
**Purpose**: Direct terminal access to workspaces
**Features**:
- Embedded terminal interface
- Quick SSH connection
- Multi-tab terminal support (future)

### 🌐 Web Services
**Purpose**: Web-based service access
**Features**:
- Jupyter Notebook access
- RStudio Server connections
- Custom web services
- Embedded browser interface

### 💾 Storage
**Purpose**: Volume and storage management
**Features**:
- Unified EFS and EBS volume management
- Tabbed interface (Shared/Private)
- Volume creation and deletion
- Attachment/detachment workflows
- Storage cost tracking

### 📊 Projects
**Purpose**: Multi-user collaboration and budgeting
**Features**:
- Project creation and management
- Budget tracking and alerts
- Member management and roles
- Cost analysis and reporting
- Project-specific resource views

### ⚙️ Settings
**Purpose**: Application configuration and advanced features
**Features**: *(See Settings Internal Navigation below)*
- General settings (system status, configuration)
- Profile management (AWS profiles and regions)
- User management (research users)
- **Advanced features** (expandable section with 6 power features)

## Settings Internal Navigation (v0.5.9+)

Settings uses a side navigation to organize configuration and advanced features:

### General (Default)
- System status and health monitoring
- Daemon connection configuration
- Auto-refresh interval settings
- Default workspace sizes
- AWS profile and region information
- Feature toggles and management
- Debug tools and troubleshooting links

### Profiles
- AWS profile management and switching
- Region configuration and selection
- Credential validation
- Profile-specific settings

### Users
- Research user management
- SSH key generation and management
- User provisioning and creation
- Multi-user collaboration setup
- UID/GID mapping configuration

### Advanced (Expandable Section)

**🔧 AMI Management**
- Custom AMI creation from workspaces
- AMI optimization and sharing
- Cross-region AMI distribution
- Community AMI discovery

**📏 Rightsizing**
- Instance sizing recommendations
- Cost optimization suggestions
- Resource utilization analysis
- Scaling predictions

**🔐 Policy Framework**
- Institutional governance controls
- Access control policies
- Template restrictions
- Compliance and audit settings

**🏪 Template Marketplace**
- Community template discovery
- Template rating and reviews
- Template installation
- Repository management

**⏰ Idle Detection**
- Automated hibernation policies
- Cost optimization through idle detection
- Policy configuration (GPU, batch, balanced)
- Hibernation history and savings tracking

**📋 Logs Viewer**
- System logs and diagnostics
- Error tracking and debugging
- API call history
- Performance monitoring

## Backend Integration

### API Client Architecture
```go
type PrismService struct {
    apiClient api.PrismAPI  // Interface to daemon
    // ... service methods exposed to frontend
}

// Daemon connection
apiClient: api.NewClient("http://localhost:8947")
```

### Real-time Data Flow
```
User Action → GUI Handler → API Client → HTTP Request
     ↓
Daemon REST API → AWS SDK → Cloud Operation
     ↓
Response → GUI Update → Notification → Refresh
```

### Supported Operations
- ✅ **Instance Lifecycle**: Launch, start, stop, delete
- ✅ **Template Management**: List, select, quick launch
- ✅ **Connection Info**: SSH details and access
- ✅ **Status Monitoring**: Real-time state and cost updates
- ✅ **Health Checks**: Daemon connectivity and error handling

## User Experience Design

### Notification System
```go
// Web-based notifications through Wails frontend
func (s *PrismService) ShowNotification(notificationType, title, message string)
- Success: Green with checkmark icon
- Error: Red with error icon  
- Info: Blue with info icon
- Auto-dismiss after 5 seconds
- Manual dismiss with × button
```

### Loading States
```go
// Non-blocking operations with visual feedback via web UI
func (s *PrismService) LaunchInstance(req LaunchRequest) {
    // Emit loading state to frontend
    s.emitEvent("launch:loading", true)
    
    // Background API calls
    go func() {
        response, err := s.apiClient.LaunchInstance(req)
        // Update frontend via events
        s.emitEvent("launch:complete", response)
    }()
}
```

### Form Validation
- Inline validation without disrupting workflow
- Clear error messages in notification area
- Required field highlighting
- Smart defaults for improved UX

## State Management

### Data Synchronization
```go
type PrismGUI struct {
    // Data state
    instances     []types.Instance
    templates     map[string]types.Template
    totalCost     float64
    lastUpdate    time.Time
    
    // Background refresh every 30 seconds
    refreshTicker *time.Ticker
}
```

### Form State
```go
// Persistent form state across navigation
launchForm struct {
    templateSelect *widget.Select
    nameEntry     *widget.Entry
    sizeSelect    *widget.Select
    launchBtn     *widget.Button
}
```

## Visual Design System

### Color Coding
- 🟢 **Running**: Green circle - instance is active
- 🟡 **Stopped**: Yellow circle - instance is stopped
- 🟠 **Pending**: Orange circle - transitional states
- 🔴 **Terminated**: Red circle - instance destroyed
- ⚫ **Unknown**: Black circle - unknown state

### Typography
- **Bold headers** for section titles and primary information
- **Regular text** for descriptions and secondary information
- **Italic text** for placeholder and helper text
- **Monospace** for technical details (IDs, commands)

### Layout Principles
- **Card containers** for grouped information
- **Grid layouts** for consistent spacing
- **Separators** for visual hierarchy
- **Spacers** for flexible positioning

## Performance Considerations

### Efficient Updates
- **Selective rendering** - Only update changed content areas
- **Background operations** - Non-blocking API calls
- **Smart refresh** - Avoid unnecessary re-renders
- **Lazy loading** - Load content on demand

### Memory Management
- **Resource cleanup** - Proper disposal of timers and resources
- **Event handling** - Efficient callback management
- **State optimization** - Minimal data retention

## Future Enhancements

### Visual Improvements
- **Dark mode support** - Theme switching capability
- **Custom icons** - Prism branded iconography  
- **Enhanced animations** - Smooth transitions and loading states
- **Responsive design** - Better window resizing behavior

### Functionality Expansion
- **Advanced filtering** - Search and filter instances/templates
- **Bulk operations** - Multi-select for batch actions
- **Activity timeline** - History of operations and changes
- **Usage analytics** - Charts and graphs for usage patterns

### Integration Features
- **Keyboard shortcuts** - Power user productivity features
- **Export capabilities** - Data export and reporting
- **Collaboration tools** - Share workstations and templates
- **Integration hooks** - External tool connections

## Development Guidelines

### Code Organization
```
cmd/prism-gui/main.go
├── Application setup and initialization
├── Navigation and layout management
├── View creation functions (Dashboard, Instances, etc.)
├── Event handlers for user interactions
├── API integration and data management
└── Utility functions and helpers
```

### Best Practices
- **Single responsibility** - Each function has a clear purpose
- **Consistent naming** - Follow Go and React/TypeScript conventions
- **Error handling** - Graceful degradation with user feedback
- **Documentation** - Clear comments for complex logic

### Testing Strategy
- **Manual testing** - User workflow verification
- **Integration testing** - API connectivity validation
- **Visual testing** - Layout and design verification
- **Performance testing** - Responsiveness under load

This GUI architecture provides a solid foundation for a modern, user-friendly cloud workstation management interface that scales with user needs and maintains excellent usability throughout the application lifecycle.