# Phase 5A: Multi-User Foundation Planning Document

**Target Version**: v0.5.0 - Q1 2025
**Planning Date**: September 28, 2025
**Status**: Planning Phase

## Overview

Phase 5A represents the foundation for Prism's evolution into a comprehensive multi-user research platform. This phase builds upon the completed Phase 4 enterprise features while laying the groundwork for advanced AWS-native research ecosystem integration.

## Current Architecture Analysis

### Phase 4 Achievements (COMPLETED)
✅ **Project-Based Organization**: Complete project lifecycle management with role-based access control
✅ **Advanced Budget Management**: Project-specific budgets with real-time tracking and automated controls
✅ **Cost Analytics**: Detailed cost breakdowns, hibernation savings, and resource utilization metrics
✅ **Multi-User Collaboration**: Project member management with granular permissions (Owner/Admin/Member/Viewer)
✅ **Enterprise API**: Full REST API for project management, budget monitoring, and cost analysis
✅ **Budget Automation**: Configurable alerts and automated actions (hibernate/stop instances, prevent launches)

### Current Multi-Modal Architecture
```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ CLI Client  │  │ TUI Client  │  │ GUI Client  │
│ (cmd/cws)   │  │ (prism tui)   │  │ (cmd/prism-gui)│
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │
       └────────────────┼────────────────┘
                        │
                 ┌─────────────┐
                 │ Backend     │
                 │ Daemon      │
                 │ (cwsd:8947) │
                 └─────────────┘
```

## Phase 5A Core Objectives

### 1. Research User Architecture 🎯

**Current State Analysis:**
- Prism currently uses single-user instances with AWS EC2 user accounts
- No consistent UID/GID mapping across instances or templates
- SSH access primarily through ec2-user or template-specific users
- Limited user identity management and persistence

**Target Architecture:**
```
Instance User Model (Dual-User System):
┌─────────────────────────┐
│ System User (ec2-user)  │ <- AWS instance management
│ ├── System services     │
│ ├── Prism    │
│ └── Administrative ops  │
└─────────────────────────┘
┌─────────────────────────┐
│ Research User           │ <- Research work environment
│ ├── Consistent UID/GID │ <- 2000-2999 range
│ ├── EFS home directory │ <- /home/researcher
│ ├── Research data      │ <- Project-specific access
│ └── Cross-template     │
│     compatibility      │
└─────────────────────────┘
```

**Implementation Requirements:**

**Research User Provisioning:**
- Consistent UID/GID mapping (e.g., UID 2001 for primary research user)
- Standardized username (`researcher`) across all templates and instances
- Home directory persistence via EFS integration
- SSH key management with research user identity
- Template compatibility layer ensuring dual-user model works across all templates

**EFS Home Directory Integration:**
```
/mnt/efs-home/
├── researcher/           <- Research user home
│   ├── .ssh/            <- SSH keys and config
│   ├── .config/         <- User configuration
│   ├── projects/        <- Project directories
│   └── shared/          <- Shared resources
└── profiles/            <- Multiple research user profiles
    ├── grad-student-1/
    ├── grad-student-2/
    └── post-doc-1/
```

**Cross-Template Compatibility:**
- Template inheritance system updated to support dual-user model
- Research user creation in all template base layers
- Standardized environment setup for research user
- Permissions and group management for project access

### 2. Optional Globus Auth Integration 🎯

**Authentication Architecture Options:**

**Option A: Enhanced Profile-Based Authentication (Recommended)**
```go
type ResearchProfile struct {
    Username         string            `json:"username"`
    UID             int               `json:"uid"`
    GID             int               `json:"gid"`
    SSHPublicKeys   []string          `json:"ssh_public_keys"`
    EFSHomePath     string            `json:"efs_home_path"`
    ProjectAccess   []string          `json:"project_access"`
    GlobalAuth      *GlobalAuthConfig `json:"globus_auth,omitempty"`
}

type GlobalAuthConfig struct {
    Enabled         bool   `json:"enabled"`
    Username        string `json:"globus_username"`
    InstitutionID   string `json:"institution_id"`
    VerifiedEmail   string `json:"verified_email"`
    AccessToken     string `json:"access_token,omitempty"` // Encrypted
}
```

**Option B: Globus Auth Integration (Advanced)**
- OAuth 2.0/OIDC integration with Globus Auth
- Institutional identity verification
- Research data access permissions
- Federated identity across research institutions

**Implementation Phases:**
1. **Phase 5A.1**: Enhanced profile system with research user management
2. **Phase 5A.2**: Optional Globus Auth integration for institutions
3. **Phase 5A.3**: Advanced federated identity features

### 3. Basic Policy Framework Integration 🎯

**Current Policy State:**
- Budget policies implemented in Phase 4
- Hibernation policies operational
- Project-based access control functional

**Enhanced Policy Architecture:**
```go
type LaunchPolicy struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Description     string                 `json:"description"`
    Scope           string                 `json:"scope"` // "user", "project", "institution"

    // Template restrictions
    AllowedTemplates    []string          `json:"allowed_templates"`
    ForbiddenTemplates  []string          `json:"forbidden_templates"`

    // Resource limitations
    MaxInstanceSize     string            `json:"max_instance_size"`
    MaxConcurrentInst   int               `json:"max_concurrent_instances"`
    MaxStorageGB        int               `json:"max_storage_gb"`

    // Time restrictions
    MaxSessionHours     int               `json:"max_session_hours"`
    AllowedTimeWindows  []TimeWindow      `json:"allowed_time_windows"`

    // Budget controls
    MaxDailyCost        float64           `json:"max_daily_cost"`
    RequireApproval     bool              `json:"require_approval"`

    // Enforcement actions
    ViolationActions    []ViolationAction `json:"violation_actions"`
}

type ViolationAction struct {
    Trigger    string `json:"trigger"`    // "exceed_budget", "forbidden_template"
    Action     string `json:"action"`     // "block", "hibernate", "notify"
    Parameters map[string]interface{} `json:"parameters"`
}
```

**Policy Integration Points:**
- Template launch validation
- Real-time resource monitoring
- Violation detection and response
- Policy inheritance (institution → project → user)
- Audit trail and compliance reporting

## Technical Implementation Plan

### Development Phases

#### Phase 5A.1: Research User Foundation (Weeks 1-3)
- [ ] Research user provisioning system
- [ ] Template dual-user model integration
- [ ] Basic EFS home directory setup
- [ ] SSH key management for research users
- [ ] Cross-template compatibility testing

#### Phase 5A.2: Enhanced Profile Management (Weeks 4-5)
- [ ] Extended profile system with research user support
- [ ] Multi-profile management UI (CLI/TUI/GUI)
- [ ] Profile switching and identity management
- [ ] Research user configuration persistence

#### Phase 5A.3: Basic Policy Engine (Weeks 6-7)
- [ ] Policy definition and storage system
- [ ] Launch-time policy enforcement
- [ ] Template filtering based on policies
- [ ] Basic violation detection and response

#### Phase 5A.4: Integration & Testing (Weeks 8)
- [ ] Multi-modal interface updates
- [ ] Comprehensive integration testing
- [ ] Documentation and user guides
- [ ] Performance optimization

### Architecture Changes

#### New Package Structure
```
pkg/
├── user/                 # Research user management
│   ├── provisioner.go    # User creation and setup
│   ├── profile.go        # Research user profiles
│   └── ssh_manager.go    # SSH key management
├── policy/               # Policy framework
│   ├── engine.go         # Policy evaluation engine
│   ├── templates.go      # Template policy integration
│   └── violations.go     # Violation detection
└── efs/                  # EFS integration
    ├── home_manager.go   # Home directory management
    └── permissions.go    # EFS permissions and access
```

#### Database Schema Extensions
```sql
-- Research users table
CREATE TABLE research_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    uid INTEGER UNIQUE NOT NULL,
    gid INTEGER NOT NULL,
    efs_home_path VARCHAR(255),
    ssh_public_keys TEXT[],
    globus_auth_data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Policy definitions
CREATE TABLE launch_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    scope VARCHAR(20) NOT NULL, -- 'user', 'project', 'institution'
    scope_id VARCHAR(100) NOT NULL, -- user_id, project_id, or institution_id
    policy_data JSONB NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Policy violations log
CREATE TABLE policy_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID REFERENCES launch_policies(id),
    user_id VARCHAR(100) NOT NULL,
    violation_type VARCHAR(50) NOT NULL,
    violation_data JSONB,
    action_taken VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);
```

### API Extensions

#### New REST Endpoints
```
# Research User Management
GET    /api/v1/users/research              # List research users
POST   /api/v1/users/research              # Create research user
GET    /api/v1/users/research/{username}   # Get research user details
PUT    /api/v1/users/research/{username}   # Update research user
DELETE /api/v1/users/research/{username}   # Delete research user

# Policy Management
GET    /api/v1/policies                    # List policies
POST   /api/v1/policies                    # Create policy
GET    /api/v1/policies/{id}              # Get policy details
PUT    /api/v1/policies/{id}              # Update policy
DELETE /api/v1/policies/{id}              # Delete policy

# Policy Validation
POST   /api/v1/policies/validate-launch   # Validate template launch against policies
GET    /api/v1/policies/violations        # Get policy violations
```

## Integration with Existing Systems

### Template System Integration
- Update template inheritance to include research user setup
- Modify base templates to create research user accounts
- Ensure SSH access works for both system and research users
- Template validation includes policy compatibility

### EFS Volume Integration
- Home directory persistence across instances
- Project-specific EFS volumes with research user access
- Shared resources and collaboration features
- Backup and snapshot management for home directories

### Profile System Enhancement
```go
type EnhancedProfile struct {
    // Existing profile fields
    Name       string `json:"name"`
    AWSProfile string `json:"aws_profile"`
    Region     string `json:"region"`

    // New research user fields
    ResearchUser    *ResearchProfile `json:"research_user,omitempty"`
    DefaultPolicies []string         `json:"default_policies"`
    GlobalAuthID    string          `json:"globus_auth_id,omitempty"`
}
```

### Multi-Modal Interface Updates

#### CLI Enhancements
```bash
# Research user management
prism user create researcher --uid 2001 --home-efs fs-abc123
prism user list --research
prism user ssh-key add researcher ~/.ssh/id_rsa.pub

# Policy management
prism policy list --scope project --project my-research
prism policy create academic-limits --template limits.yaml
prism policy validate --template ml-gpu --user researcher

# Profile with research user
prism profile create research-profile --research-user researcher --globus-auth
```

#### TUI Enhancements
- Research user management page (Page 7: Users)
- Policy visualization and management
- Enhanced profile switching with research user context
- Policy violation alerts and notifications

#### GUI Enhancements
- Research user setup wizard
- Policy management interface with visual policy builder
- Enhanced profile management with research user integration
- Compliance dashboard showing policy adherence

## Risk Analysis & Mitigation

### Security Considerations
**Risk**: Research user privilege escalation
**Mitigation**: Strict UID/GID management, sudo restrictions, security group isolation

**Risk**: EFS home directory permissions
**Mitigation**: AWS IAM integration, file-level permissions, audit logging

**Risk**: Policy bypass attempts
**Mitigation**: Server-side validation, immutable policy enforcement, violation logging

### Performance Considerations
**Risk**: EFS home directory latency
**Mitigation**: EFS performance mode selection, local caching strategies

**Risk**: Policy evaluation overhead
**Mitigation**: Policy caching, pre-computed policy results, async evaluation

### Operational Considerations
**Risk**: Template compatibility breaking
**Mitigation**: Comprehensive testing matrix, backward compatibility layers

**Risk**: Complex user onboarding
**Mitigation**: Setup wizards, automated provisioning, clear documentation

## Success Metrics

### Technical Metrics
- Research user provisioning time < 30 seconds
- EFS home directory mount time < 10 seconds
- Policy evaluation latency < 100ms
- Template compatibility rate > 95%
- Zero security violations during testing

### User Experience Metrics
- User onboarding completion rate > 90%
- Research user adoption rate > 60% within 3 months
- Policy violation false positive rate < 5%
- Cross-template user experience consistency score > 8/10

### Business Metrics
- Educational institution pilot program readiness
- Multi-user collaboration workflow enablement
- Foundation for Phase 5B AWS research services integration
- Preparation for institutional deployment capabilities

## Next Steps

1. **Immediate Actions** (This Week):
   - [ ] Finalize technical architecture decisions
   - [ ] Create detailed implementation tickets
   - [ ] Set up development environment for Phase 5A
   - [ ] Begin research user provisioning system development

2. **Phase 5A.1 Sprint Planning** (Next Week):
   - [ ] Research user provisioning implementation
   - [ ] Template dual-user model design
   - [ ] EFS home directory integration prototype
   - [ ] Initial testing framework setup

3. **Stakeholder Communication**:
   - [ ] Share Phase 5A plan with research community
   - [ ] Gather feedback from potential institutional users
   - [ ] Coordinate with AWS education program for testing
   - [ ] Plan Phase 5B AWS services integration

## Conclusion

Phase 5A establishes the critical foundation for Prism's evolution into a comprehensive multi-user research platform. By implementing research user architecture, enhanced policy management, and optional Globus Auth integration, we create the necessary infrastructure for institutional adoption while maintaining the simplicity and power that defines Prism.

This phase bridges the gap between Phase 4's enterprise features and Phase 5B's advanced AWS research services integration, positioning Prism as the leading cloud-native research computing platform for academic institutions and collaborative research environments.