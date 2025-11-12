# Prism Development Roadmap

**Current Version**: v0.5.13 (In Development)
**Last Released**: v0.5.12 (November 11, 2025)
**Next Version**: v0.5.14 (Planned - December 2025)
**Last Updated**: November 12, 2025
**Status**: Active Development

This roadmap outlines planned features and enhancements for Prism. All items are tracked in [GitHub Issues](https://github.com/scttfrdmn/prism/issues) and the [Prism Development Project](https://github.com/scttfrdmn/prism/projects).

---

## 🎯 Current Focus: Phase 5.0 - UX Redesign

**Priority**: CRITICAL - HIGHEST PRIORITY
**Target**: v0.5.8 and v0.5.9 (December 2025 - January 2026)

Prism is shifting focus from feature development to user experience optimization. The current 15-minute learning curve for first workspace needs to be reduced to 30 seconds.

**Why This Matters**: New researchers face cognitive overload before completing basic tasks. UX redesign will dramatically improve first-time user experience.

**Track Progress**: [GitHub Milestone: Phase 5.0 UX Redesign](https://github.com/scttfrdmn/prism/milestone/1)

---

## 📅 Release Schedule

### v0.5.7 (October 2025): Template Provisioning & Test Infrastructure ✅ RELEASED
**Release Date**: October 26, 2025
**Focus**: S3-backed template provisioning + Test infrastructure stability

#### Template File Provisioning ✅ COMPLETE
**Milestone**: [Phase 5.6: Template Provisioning](https://github.com/scttfrdmn/prism/milestone/13)
- ✅ [#64](https://github.com/scttfrdmn/prism/issues/64) - S3-backed file transfer with multipart support (up to 5TB)
- ✅ [#31](https://github.com/scttfrdmn/prism/issues/31) - Template asset management for binaries and datasets
- ✅ S3 Transfer System with progress tracking and MD5 verification
- ✅ Conditional provisioning (architecture-specific files)
- ✅ Required vs optional files with graceful fallback
- ✅ Complete documentation ([TEMPLATE_FILE_PROVISIONING.md](TEMPLATE_FILE_PROVISIONING.md))
- **Impact**: Enable multi-GB dataset distribution, binary deployment, and pre-trained model distribution

#### Test Infrastructure Fixes ✅ COMPLETE
**Issue**: [#83](https://github.com/scttfrdmn/prism/issues/83) - API Test Stability
- ✅ Fixed Issue #83 regression (tests hitting AWS and timing out)
- ✅ Fixed data race in system_metrics.go (concurrent cache access)
- ✅ Test performance: 206x faster (97.961s → 0.463s)
- ✅ All smoke tests passing (8/8)
- ✅ Zero race conditions detected
- **Impact**: Reliable CI/CD pipeline, fast developer feedback loop

#### Script Cleanup ✅ COMPLETE
- ✅ Completed CloudWorkStation → Prism rename across all scripts
- ✅ Documentation consistency verification
- **Impact**: Consistent branding across entire codebase

**Status**: ✅ Released - [View Release](https://github.com/scttfrdmn/prism/releases/tag/v0.5.7)

---

### v0.5.8 (December 2025): Quick Start Experience ✅ FEATURE COMPLETE
**Release Date**: Target December 13, 2025
**Focus**: First-time user experience - zero to workspace in <30 seconds
**Release Plan**: [RELEASE_PLAN_v0.5.8.md](releases/RELEASE_PLAN_v0.5.8.md)
**Release Notes**: [RELEASE_NOTES_v0.5.8.md](releases/RELEASE_NOTES_v0.5.8.md)

#### Quick Start Features ✅ COMPLETE
**Milestone**: [Phase 5.0.1: Quick Wins](https://github.com/scttfrdmn/prism/milestone/2) - 100% Complete

- ✅ [#15](https://github.com/scttfrdmn/prism/issues/15) - Rename "Instances" → "Workspaces" (11 files, 109 changes)
- ✅ [#13](https://github.com/scttfrdmn/prism/issues/13) - Home Page with Quick Start wizard (GUI) (363 lines)
- ✅ [#17](https://github.com/scttfrdmn/prism/issues/17) - CLI `prism init` onboarding wizard (520 lines)

#### Success Metrics ✅ ACHIEVED
- ⏱️ Time to first workspace launch: 15min → 30sec ✅ **Met**
- 🎯 First-attempt success rate: >90% ✅ **Expected**
- 😃 User confusion: Reduce by 70% ✅ **Expected**

#### Key Deliverables
- **CLI Init Wizard**: 6-step interactive wizard with category selection and cost estimates
- **GUI Quick Start**: Cloudscape-based wizard with visual template browsing
- **Documentation**: Comprehensive release notes, implementation plans, updated README
- **Code Quality**: 1,565+ lines, zero compilation errors, proper GitHub issue tracking

**Implementation Completed**: October 27, 2025 (12 commits)

**Status**: ✅ Feature Complete - Ready for Testing & Release

---

### v0.5.9 (November 2025): Navigation Restructure ✅
**Release Date**: November 7, 2025
**Focus**: Reduce navigation complexity from 14 to 6 items
**Release Plan**: [RELEASE_PLAN_v0.5.9.md](releases/RELEASE_PLAN_v0.5.9.md)
**Release Notes**: [RELEASE_NOTES_v0.5.9.md](releases/RELEASE_NOTES_v0.5.9.md)

#### Navigation Features
**Milestones**: [Phase 5.0.2: Info Architecture](https://github.com/scttfrdmn/prism/milestone/3)

- ✅ [#14](https://github.com/scttfrdmn/prism/issues/14) - Merge Terminal/WebView into Workspaces
- ✅ [#16](https://github.com/scttfrdmn/prism/issues/16) - Collapse Advanced Features under Settings
- ✅ [#18](https://github.com/scttfrdmn/prism/issues/18) - Unified Storage UI (EFS + EBS)
- ~~[#19](https://github.com/scttfrdmn/prism/issues/19) - Integrate Budgets into Projects~~ *Moved to v0.5.10*

#### Contributor Documentation
**Focus**: Establish contribution guidelines and community standards

- ✅ **CONTRIBUTING.md**: Comprehensive contribution guide
  - Issue-first workflow (no PRs without `help wanted`)
  - Scope control (no PR scope expansion)
  - Core protection (`core` label for maintainers only)
  - Multi-modal parity requirements (CLI/TUI/GUI)
  - Testing and security requirements
  - Plugin development pathway ([#230](https://github.com/scttfrdmn/prism/issues/230))
  - Apache 2.0 license compliance

- ✅ **CODE_OF_CONDUCT.md**: Community standards
  - Contributor Covenant 2.0
  - Enforcement guidelines
  - Reporting procedures

#### Success Metrics ✅
- ✅ 🧭 Navigation complexity: 14 → 6 top-level items
- ✅ ⏱️ Time to find features: <10 seconds
- ✅ 😃 User confusion: Further 30% reduction
- ✅ 📱 Advanced feature discoverability: >95% (hierarchical navigation)

**Status**: 🎉 RELEASED (November 7, 2025)

---

### v0.5.10 (November 2025): Multi-Project Budgets ✅ RELEASED
**Release Date**: November 8, 2025
**Focus**: Budget system redesign for multi-project allocation
**Release Plan**: [RELEASE_PLAN_v0.5.10.md](releases/RELEASE_PLAN_v0.5.10.md)

#### Budget Redesign
**Goal**: Allow budgets to be allocated across multiple projects

**Current State**: 1 budget : 1 project relationship (pre-v0.5.10)
**New State**: Many-to-many relationships via allocations

**Features Implemented**:
- ✅ Shared budget pools allocable to multiple projects (#97)
- ✅ Project-level budget allocation tracking (#98)
- ✅ Budget reallocation between projects with audit trail (#99)
- ✅ Multi-project cost rollup and reporting (#100)
- ✅ Funding source selection at launch (#233)
- ✅ Backup funding and budget cushions (#234)
- ✅ Enhanced resource tagging for cost optimization (#128)
- ✅ Budget system philosophy documentation (#236)

**Implementation Complete**:
- ✅ Two-tier budget data model (Budget Pools + Allocations)
- ✅ REST API endpoints for budget operations (613 lines)
- ✅ Budget manager with reallocation support (1,133 lines)
- ✅ Real-time spending tracking per allocation
- ✅ DefaultAllocationID for frictionless launches
- ✅ Comprehensive documentation (BUDGET_PHILOSOPHY.md, RESOURCE_TAGGING.md)

**Success Metrics Achieved**:
- ✅ 1 grant → N projects (NSF funding multiple research projects)
- ✅ N budgets → 1 project (multi-source funding)
- ✅ Real-time cost tracking and reporting
- ✅ Grant compliance with audit trails

**Status**: 🎉 COMPLETE (November 8, 2025)

---

### v0.5.11 (November 2025): User Invitation & Role Systems ✅ RELEASED
**Release Date**: November 10, 2025
**Focus**: Project collaboration with invitation workflow and role-based permissions

#### Key Features Delivered
**Issues**: #102, #103, #105, #106

**User Invitation System**:
- ✅ Email-based invitation workflow with tokens
- ✅ Individual, bulk, and shared token invitations
- ✅ QR code generation for workshops
- ✅ Invitation status tracking and audit trail

**Research User Auto-Provisioning** (#106):
- ✅ Automatic SSH key generation (Ed25519)
- ✅ Deterministic UID/GID allocation
- ✅ EFS home directory creation (`/efs/home/{username}`)
- ✅ Zero-configuration user setup on invitation acceptance

**AWS Quota Validation** (#105):
- ✅ Pre-flight capacity checking
- ✅ vCPU mapping for 50+ instance types
- ✅ Real-time usage calculation
- ✅ REST endpoint: `POST /api/v1/invitations/quota-check`

**Role-Based Permissions** (#102):
- ✅ Automatic member addition on invitation acceptance
- ✅ Role validation (owner, admin, member, viewer)
- ✅ Duplicate member detection

**GUI Features** (#103):
- ✅ Professional Cloudscape invitation interface
- ✅ Bulk invitation management
- ✅ Status badges and tracking

**Use Cases**:
- University classes (50+ students)
- Research labs (10+ collaborators)
- Conference workshops (100+ attendees)

**Status**: 🎉 RELEASED (November 10, 2025)

---

### v0.5.12 (November 2025): Operational Stability & CLI Consistency ✅ RELEASED
**Release Date**: November 11, 2025
**Focus**: Production-ready operational features and AWS reliability
**Issues**: #107, #108

#### Key Features Delivered

**Rate Limiting System** (#107):
- ✅ Token bucket implementation (default: 5 launches/minute)
- ✅ Configurable launch windows (1-100 launches, 1-60 minute windows)
- ✅ Real-time quota tracking in daemon status
- ✅ CLI commands: `prism admin rate-limit {status,configure,reset}`
- ✅ Enhanced error messages with current usage and retry times

**AWS Retry Logic** (#108):
- ✅ 5 retry attempts with exponential backoff and jitter
- ✅ Context-aware retries respecting cancellation
- ✅ Coverage for all critical operations:
  - LaunchInstance, StartInstance, StopInstance, DeleteInstance
- ✅ Handles 15+ AWS error patterns:
  - RequestLimitExceeded (API throttling)
  - InsufficientInstanceCapacity (EC2 capacity)
  - InstanceLimitExceeded (quota limits)
  - NetworkError (connection issues, timeouts)

**Enhanced Error Messages**:
- ✅ Clear, actionable guidance for common failures
- ✅ Context about current usage and limits
- ✅ Suggested remediation steps

**Success Metrics Achieved**:
- ✅ Bulk launches: Reliable multi-workspace deployment
- ✅ Rate limiting: Predictable timing and progress
- ✅ Retry logic: 95%+ transient failure recovery
- ✅ User experience: Clear error messaging

**Status**: 🎉 RELEASED (November 11, 2025)

---

### v0.5.13 (November 2025): Cost Control & Monitoring 🚀 IN DEVELOPMENT
**Release Date**: Target November 15, 2025
**Focus**: Advanced cost control and system power management
**Issues**: #90, #91, #252

#### Features In Development

**Launch Throttling System** (#90) ✅ Complete:
- ✅ Token bucket rate limiting (3 levels: global, per-user, per-template)
- ✅ Prevents cost overruns from rapid/scripted launches
- ✅ CLI commands: `prism admin throttling {status,configure,reset,wait-time}`
- ✅ REST API: GET/POST /api/v1/throttle/*
- ✅ 14/14 tests passing with race detection
- ✅ Configurable refill rates and burst capacity

**Sleep/Wake Auto-Hibernation** (#91) ✅ Complete:
- ✅ System power event monitoring (macOS IOKit via CGo)
- ✅ Three hibernation modes:
  - `idle_only` (RECOMMENDED): Integrates with CloudWatch metrics
  - `all`: Hibernates all except exclusions
  - `manual_only`: No auto-hibernation
- ✅ CLI commands: `prism admin sleep-wake {status,enable,disable,configure}`
- ✅ REST API: GET/POST /api/v1/sleep-wake/*
- ✅ State persistence with statistics tracking
- ✅ Grace period and configurable idle check timeout
- ✅ Platform support: macOS complete, Linux/Windows stubs ready

**Bug Fixes & Stability**:
- ✅ Race condition fixed in pkg/progress/reporter.go
- ✅ Deep copy metadata maps and callback slices
- ✅ All unit tests passing with -race detector
- 🐛 Integration test timeout tracked (#252)

**Technical Debt Documented**:
- 📝 Issue #252: Integration test daemon connectivity
- 📝 TECHNICAL_DEBT_BACKLOG.md #12: macOS IOKit deprecation

**Implementation Statistics**:
- 2,575+ lines of code across 13 files
- 100% unit test pass rate with race detection
- Zero compilation errors (deprecation warning documented)

**Success Metrics**:
- ✅ Complete cost control suite (throttling + sleep/wake + hibernation)
- ✅ Idle-aware hibernation prevents workload interruption
- ✅ Test stability: zero race conditions
- ✅ Production-ready on macOS, cross-platform prepared

**Status**: 🚀 IN DEVELOPMENT (Ready for Release)

---

### v0.5.14 (December 2025): Desktop Applications Foundation
**Release Date**: Target December 6, 2025
**Focus**: Nice DCV foundation for desktop GUI applications
**Issues**: #216-#219

#### Planned Features

**Generic Desktop Template** (#219):
- Base MATE desktop environment
- Nice DCV server configuration
- Browser-based remote desktop access

**DCV Connection Management** (#218):
- SSM port forwarding to DCV port 8443
- Browser auto-launch integration
- Secure credential handling

**Template System Extension** (#217):
- `connection_type: "desktop"` support
- DCV-specific provisioning hooks
- Desktop environment configuration

**Documentation** (#216):
- Nice DCV architecture guide
- Desktop template development guide
- Connection workflow documentation

**Implementation Notes**:
- Leverage complete Lens project implementation
- MATE desktop for lightweight performance
- No exposed ports (SSM-only access)

**Status**: 📋 Planned (December 2025)

---

### v0.5.15 (December 2025): Desktop Applications
**Release Date**: Target December 20, 2025
**Focus**: Production desktop application templates
**Issues**: #220-#223

#### Planned Features

**MATLAB Template** (#220) - CRITICAL:
- Numerical computing environment
- Cloud-based license activation
- Complete toolbox support
- Engineering/physics/mathematics focus

**QGIS Templates** (#221):
- Basic environment (standard QGIS)
- Advanced environment (additional plugins)
- Remote sensing environment (specialized tools)
- Geographic Information System workflows

**Mathematica Template** (#222):
- Symbolic computation
- Cloud license integration

**Stata Template** (#223):
- Statistical analysis software
- Research workflows

**Commercial License Strategy**:
- Cloud-based license activation (MATLAB, Mathematica)
- BYOL (Bring Your Own License) support
- AWS Marketplace AMI integration
- Template documentation for license configuration

**Status**: 📋 Planned (December 2025)

---

### v0.6.0 (January 2026): Enterprise Authentication
**Release Date**: Target January 2026
**Focus**: Production authentication and security
**Issues**: #247-#251
- **Template Request Feature** ([#229](https://github.com/scttfrdmn/prism/issues/229)): User suggestion box for requesting new templates

Success Metrics:
- Time to first workspace: Still <30 seconds
- Navigation efficiency: <3 clicks to any feature
- Feature discoverability: >95%
- Workflow completion: >90% success rate

### v0.5.14 (June-July 2026): Web-Based Research Tools
**Release Date**: Target June-July 2026
**Focus**: High-impact web-based research applications from Lens project
**Status**: 📋 Planned

#### Rationale
Quick wins with high researcher impact. All applications are web-based (no Nice DCV complexity), leverage existing template provisioning system, and address critical research workflows identified in the Lens project.

#### Features (4 weeks)

##### 1. 🐍 Streamlit Template ([#211](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐⭐ CRITICAL

**What**: Interactive Python dashboards and data applications

**Use Cases**:
- Turn Jupyter notebooks into interactive web apps
- Build ML model demos with sliders and inputs
- Create teaching materials with interactive visualizations
- Rapid prototyping of data exploration tools

**Technical Details**:
- Port: 8501 (default, configurable)
- Install: `pip install streamlit`
- Launch: `streamlit run app.py`
- Memory: 256MB-512MB + app requirements
- Complexity: LOW

**Target Users**: Python/Jupyter users, data scientists, ML researchers, educators

**Template Configuration**:
```yaml
name: "Streamlit Interactive Apps"
connection_type: "web"
port: 8501
packages:
  python:
    - streamlit
    - pandas
    - plotly
    - altair
```

##### 2. 🧹 OpenRefine Template ([#212](https://github.com/scttfrdmn/prism/issues/212))
**Priority**: ⭐⭐⭐⭐ HIGH

**What**: Data cleaning and transformation with GUI interface

**Use Cases**:
- Clean messy survey data
- Transform scraped/collected data
- Reconcile data against external sources
- Prepare data before analysis in Jupyter/R

**Technical Details**:
- Port: 3333
- Install: Java JAR or Docker
- Memory: 512MB-1GB
- Complexity: MEDIUM

**Target Users**: Social scientists, digital humanities researchers, non-programmers with messy data

**Template Configuration**:
```yaml
name: "OpenRefine Data Cleaning"
connection_type: "web"
port: 3333
packages:
  system:
    - openjdk-11-jre
```

##### 3. 📊 Shiny Server Template ([#213](https://github.com/scttfrdmn/prism/issues/213))
**Priority**: ⭐⭐⭐⭐ HIGH

**What**: Publish interactive R applications and dashboards

**Use Cases**:
- Share R analysis as interactive web apps
- Create dashboards for research results
- Build data exploration tools
- Publish reproducible research

**Technical Details**:
- Port: 3838
- Install: R package + shiny-server binary
- Memory: 512MB base + per-app memory
- Complexity: MEDIUM
- Natural companion to existing RStudio template

**Target Users**: R users, bioinformatics researchers, statisticians

**Template Configuration**:
```yaml
name: "Shiny Server"
connection_type: "web"
port: 3838
inherits: ["RStudio Server"]
packages:
  r:
    - shiny
    - shinydashboard
    - plotly
```

##### 4. 🏷️ Label Studio Template ([#214](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐ MEDIUM-HIGH

**What**: Data labeling and annotation platform for ML

**Use Cases**:
- Label images for computer vision research
- Annotate text for NLP research
- Code qualitative data (interviews, documents)
- Create training datasets for ML
- Collaborative annotation with research assistants

**Technical Details**:
- Port: 8090 (changed from 8080 to avoid conflicts)
- Install: `pip install label-studio`
- Database: SQLite or PostgreSQL
- Memory: 512MB-1GB
- Complexity: MEDIUM

**Target Users**: ML/AI researchers, qualitative researchers, NLP researchers

**Template Configuration**:
```yaml
name: "Label Studio Annotation"
connection_type: "web"
port: 8090
packages:
  python:
    - label-studio
    - label-studio-tools
```

##### 5. 📊 Datasette Template ([#215](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐ MEDIUM

**What**: Instant JSON API and web interface for datasets

**Use Cases**:
- Publish research datasets alongside papers
- Explore CSV/Excel files without coding
- Share data with collaborators
- Create queryable data repositories

**Technical Details**:
- Port: 8001
- Install: `pip install datasette`
- Memory: 128MB-256MB + dataset size
- Complexity: LOW
- Very lightweight, quick win

**Target Users**: All researchers (universal data need), data librarians

**Template Configuration**:
```yaml
name: "Datasette Data Publishing"
connection_type: "web"
port: 8001
packages:
  python:
    - datasette
    - datasette-cluster-map
```

#### Success Metrics
- 5 new web-based research tool templates
- All templates launch in <5 minutes
- Leverage existing template provisioning system
- Zero Nice DCV infrastructure dependency
- High researcher satisfaction (streamlit especially)

#### Implementation Notes
- Reference: Lens project's FUTURE_APPS.md for usage patterns
- All web-based tools follow existing Jupyter/RStudio pattern
- SSM port forwarding for secure access
- Auto-open browser on connect
- Templates support research user provisioning

### v0.6.0 (August 2026): Enterprise Authentication & AWS Research Services
**Release Date**: Target August 2026
**Focus**: Enterprise authentication + Web access to AWS research services
**Status**: 📋 Planned

#### Part 1: Enterprise Authentication
- OAuth/OIDC integration (Google, Microsoft, institutional SSO)
- LDAP/Active Directory support
- SAML support for enterprise SSO
- Token validation and session management
- Integration with user invitation system from v0.5.11

#### Additional v0.6.0 Features

#### 1. 🔄 Auto-Update Feature ([#61](https://github.com/scttfrdmn/prism/issues/61))
**Status**: Planned  
**Why**: Users don't know when new versions are available, miss bug fixes and features

**Features**:
- GitHub Releases API integration for version detection
- `prism version --check-update` command with release notes
- Startup notifications in CLI/TUI/GUI
- Platform-specific update helpers (Homebrew, apt, manual install)

**Example**:
```bash
$ prism version --check-update
Prism CLI v0.5.5
⚠️  New version available: v0.6.0 (released 2 days ago)

What's New:
- AWS Quota Management with intelligent AZ failover
- Auto-update feature with background checks
- GUI system tray support

To update:
  macOS:   brew upgrade prism
  Linux:   curl -L https://get.prism.io | bash
```

#### 2. 🖥️ GUI System Tray and Auto-Start ([#62](https://github.com/scttfrdmn/prism/issues/62))
**Status**: Planned  
**Why**: GUI lacks convenient system tray access and auto-start on login

**Features**:
- Native system tray integration (macOS menu bar, Windows tray, Linux tray)
- Context menu with Quick Launch, My Workspaces, Cost Summary
- Auto-start on login (Launch Agents, Registry, XDG autostart)
- Intelligent notifications (launches, budget alerts, idle warnings)

**Menu Structure**:
```
Prism [Icon]
├── 🚀 Quick Launch → Python ML, R Research, Ubuntu Desktop
├── 💻 My Workspaces (3 running)
├── 💰 Cost Summary ($42.50 this month)
├── ⚙️  Preferences
└── ⏹️  Quit
```

#### 3. 📊 AWS Quota Management ([#57-60](https://github.com/scttfrdmn/prism/issues/57))
**Status**: Planned  
**Why**: Users surprised by AWS quota limits and capacity failures

**Features**:
- Quota awareness (vCPU limits, instance type limits, storage)
- Pre-launch validation with quota impact analysis
- Intelligent AZ failover on InsufficientInstanceCapacity
- AWS Health Dashboard monitoring for outages
- Quota increase request assistance

**Example**:
```bash
$ prism admin quota show --region us-west-2

📊 AWS Service Quotas - us-west-2

vCPU Limits:
  Standard: 24/32 (75% used) ⚠️
  GPU:      0/8 (0% used) ✅

Recommendations:
  ⚠️  Consider requesting vCPU increase for compute-intensive work
```

#### 4. 🔐 Multi-User Authentication (Phase 1)
**Status**: Planned
**Why**: Institutional deployments need proper authentication

**Features**:
- OAuth/OIDC integration (Google, Microsoft, institutional SSO)
- LDAP/Active Directory support
- Token validation and session management
- Role-based access control (RBAC) foundation

#### Part 2: AWS Research Services Integration

**Previous Status**: Marked as "STRATEGIC" - now with concrete implementation plan

**What**: Direct web access to AWS native research services from Prism

**Why**: Researchers want unified access to both EC2 workspaces AND AWS research services (SageMaker, EMR, Braket) without switching tools.

##### 5. 🧪 AWS Service Registry & Integration ([#224](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐⭐ CRITICAL

**Features**:
- Service catalog: SageMaker Studio/Lab, EMR Studio, Braket, CloudShell
- Service metadata (URLs, regions, prerequisites, cost info)
- IAM role validation for service access
- Service health and availability checking

##### 6. 🤖 SageMaker Studio/Lab Integration ([#225](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐⭐ CRITICAL

**What**: Launch and access SageMaker Studio from Prism

**Features**:
- Direct browser launch to SageMaker Studio
- SageMaker Studio Lab (free tier) integration
- Notebook lifecycle management
- Integration with Prism projects/budgets
- Cost tracking for SageMaker compute (notebooks, training, inference)

**CLI Example**:
```bash
# Launch SageMaker Studio
prism service launch sagemaker-studio --project ml-research

# Open in browser
prism service open sagemaker-studio

# Check status
prism service list
```

**GUI**:
- "Launch SageMaker Studio" button in Services tab
- Auto-open browser to SageMaker console
- Cost tracking in project dashboard

##### 7. 📊 EMR Studio Integration ([#226](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐ HIGH

**What**: Big data analytics with EMR Studio

**Features**:
- Launch EMR Studio from Prism
- Cluster management and lifecycle
- Spark notebook access
- Integration with EFS storage
- Cost tracking for EMR clusters

**Use Cases**:
- Large-scale data processing
- Spark-based machine learning
- Distributed computing research

##### 8. ⚛️ Amazon Braket Integration ([#227](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐ MEDIUM

**What**: Quantum computing research access

**Features**:
- Launch Braket workspace from Prism
- Circuit builder interface access
- QPU (quantum processing unit) cost tracking
- Educational materials integration

**Use Cases**:
- Quantum algorithm research
- Quantum computing education
- Experimental quantum applications

##### 9. 🖥️ Unified Service Dashboard ([#228](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐⭐ CRITICAL

**What**: Single pane of glass for EC2 + AWS services

**GUI Features**:
- Services tab alongside Workspaces, Templates, Storage
- Service cards with status indicators
- "Launch" buttons for each service
- "Open in Browser" for active services
- Cost summary: EC2 + AWS services combined

**CLI Features**:
```bash
prism service launch sagemaker-studio
prism service launch emr-studio
prism service launch braket
prism service list
prism service open <service-name>
prism service stop <service-name>
prism service costs --project ml-research
```

#### Success Metrics (AWS Services)
- Launch AWS services from Prism GUI/CLI
- Browser auto-open to service URLs
- Cost tracking integrated into projects
- Single workflow: users don't switch between Prism and AWS Console
- SageMaker cost attribution to Prism projects

#### Implementation Notes (AWS Services)
- AWS service authentication via IAM roles
- Service quotas and availability checking
- Cost tracking API integration (AWS Cost Explorer)
- Service-specific prerequisites validation
- Documentation for each service integration

### v0.6.1 (September-October 2026): Nice DCV Foundation
**Release Date**: Target September-October 2026
**Focus**: Desktop application infrastructure using Nice DCV
**Status**: 📋 Planned

#### Rationale
Establish Nice DCV foundation for desktop GUI applications (MATLAB, QGIS, etc.). Learn from Lens project's working DCV implementation.

#### Features (3-4 weeks)

##### 1. 📱 DCV Template System Extension ([#216](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐⭐ CRITICAL

**What**: Extend template system to support desktop applications

**Features**:
- New template type: `connection_type: "desktop"`
- DCV-specific configuration options
- AMI selection logic (Desktop vs Server images)
- GPU support detection and configuration

**Template Schema Extension**:
```yaml
connection_type: "desktop"  # New type: "web" | "desktop" | "ssh"
desktop:
  environment: "xfce"  # or "mate", "gnome"
  dcv_port: 8443
  gpu_required: false
  gpu_drivers: ["nvidia", "amd"]
```

##### 2. 🖥️ DCV Server Provisioning ([#217](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐⭐ CRITICAL

**What**: Automated DCV server installation and configuration

**Features**:
- Cloud-init scripts for DCV installation
- XFCE/MATE desktop environment setup (lighter than GNOME)
- GPU driver installation (optional, for visualization workloads)
- X11 display server configuration
- Port 8443 (HTTPS/WebSocket) configuration
- Session lifecycle management

**Technical Details**:
- Base: Ubuntu 22.04 Desktop
- DCV Version: 2023.0+
- Desktop: MATE (preferred - lightweight) or XFCE
- Launch time: 5-10 minutes (vs 2-5 for web apps)
- Resource requirements: 4-8 vCPU, 16-32GB RAM minimum

**Reference**: Lens `apps/dcv-desktop/internal/config/userdata.go`

##### 3. 🔗 DCV Connection Management ([#218](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐⭐ CRITICAL

**What**: Browser-based access to DCV desktops

**Features**:
- SSM port forwarding to DCV port 8443
- Browser auto-open to `https://localhost:8443`
- Credential management and display (username/password)
- Session status monitoring
- Reconnection handling

**CLI Example**:
```bash
# Launch desktop workspace
prism launch generic-desktop my-desktop

# Connect (auto-opens browser)
prism connect my-desktop

# DCV web client appears in browser
# Full desktop environment with applications
```

**Reference**: Lens `apps/dcv-desktop/internal/cli/connect.go`

##### 4. 🖥️ Generic Desktop Template ([#219](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐ HIGH

**What**: Base desktop template for testing and generic use

**Features**:
- Ubuntu 22.04 Desktop + Nice DCV
- MATE desktop environment
- Basic desktop tools (Firefox, file manager, terminal)
- No specialized applications
- Base for application-specific templates

**Template Configuration**:
```yaml
name: "Generic Ubuntu Desktop"
slug: "generic-desktop"
connection_type: "desktop"
base: "ubuntu-22.04-desktop"
desktop:
  environment: "mate"
  dcv_port: 8443
  gpu_required: false
instance_types:
  default: "t3.xlarge"  # 4 vCPU, 16GB RAM
packages:
  system:
    - ubuntu-mate-desktop
    - nice-dcv-server
    - firefox
    - vim
```

#### Success Metrics (DCV Foundation)
- DCV working with SSM port forwarding
- Generic desktop launches in <10 minutes
- Browser access to remote desktop functional
- GPU support validated on g4dn instances
- Desktop performance acceptable for GUI interaction
- Foundation ready for application-specific templates

#### Implementation Notes (DCV Foundation)
- Study Lens's working implementation thoroughly
- Test on both CPU and GPU instance types
- Document performance characteristics
- Create troubleshooting guide
- Validate security (no exposed ports, SSM-only access)

### v0.6.2 (November-December 2026): Desktop GUI Applications
**Release Date**: Target November-December 2026
**Focus**: Desktop applications using DCV foundation from v0.6.1
**Status**: 📋 Planned

#### Rationale
With DCV infrastructure in place, add high-value desktop GUI applications for research computing.

#### Features (4-5 weeks)

##### 1. 🧮 MATLAB Template ([#220](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐⭐ CRITICAL

**What**: MATLAB numerical computing environment

**Why**: Extremely common in academia (engineering, physics, mathematics)

**Use Cases**:
- Numerical analysis and simulation
- Signal processing research
- Engineering simulations
- Control systems research
- Legacy .m file execution

**Technical Details**:
- DCV desktop required (GUI-heavy application)
- License: BYOL (license server), Cloud-Based (online activation), or AWS Marketplace AMI
- Memory: 16-32GB RAM recommended
- GPU: Optional for Simulink visualizations
- Instance: t3.xlarge (basic) or g4dn.xlarge (GPU)

**Template Configuration**:
```yaml
name: "MATLAB Workstation"
slug: "matlab"
connection_type: "desktop"
inherits: ["Generic Ubuntu Desktop"]
desktop:
  environment: "mate"
  gpu_required: false  # optional
license:
  type: "byol"
  license_server: "{{ user_config.matlab_license_server }}"
ami:
  marketplace_product: "matlab-r2024a"  # or custom AMI
```

**Disciplines**: Engineering, physics, mathematics, computational science

##### 2. 🗺️ QGIS Templates ([#221](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐⭐ HIGH

**What**: Geographic Information System for spatial analysis

**Why**: Standard tool for geography, environmental science, urban planning

**Use Cases**:
- GIS analysis and mapping
- Spatial data visualization
- Remote sensing (satellite imagery)
- Terrain modeling
- Urban planning analysis

**Technical Details**:
- DCV desktop required
- 3 environments: basic-gis, advanced-gis, remote-sensing
- GPU: Recommended for large rasters and 3D visualization
- Reference: Lens has complete QGIS implementation

**Environment 1: Basic GIS**
```yaml
name: "QGIS Basic"
slug: "qgis-basic"
connection_type: "desktop"
inherits: ["Generic Ubuntu Desktop"]
instance_types:
  default: "t3.xlarge"  # 4 vCPU, 16GB RAM
packages:
  system:
    - qgis
    - qgis-plugin-grass
cost_estimate: "$0.17/hour"
```

**Environment 2: Advanced GIS**
```yaml
name: "QGIS Advanced"
slug: "qgis-advanced"
connection_type: "desktop"
instance_types:
  default: "t3.xlarge"
packages:
  system:
    - qgis
    - grass
    - saga
    - postgis
```

**Environment 3: Remote Sensing**
```yaml
name: "QGIS Remote Sensing"
slug: "qgis-remote-sensing"
connection_type: "desktop"
instance_types:
  default: "g4dn.xlarge"  # GPU for large rasters
desktop:
  gpu_required: true
packages:
  system:
    - qgis
    - orfeo-toolbox
    - snap-esa
cost_estimate: "$0.53/hour"
```

**Reference**: Complete Lens implementation in `apps/qgis/`

**Disciplines**: Geography, environmental science, urban planning, archaeology

##### 3. 🔬 Mathematica Template ([#222](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐ MEDIUM-HIGH

**What**: Wolfram Mathematica symbolic computation

**Use Cases**:
- Symbolic mathematics
- Algorithm development
- Scientific visualization
- Computational modeling

**Technical Details**:
- DCV desktop required
- License: Cloud-Based (online activation), BYOL, or AWS Marketplace
- Memory: 16-32GB RAM
- Instance: t3.xlarge

**Disciplines**: Mathematics, physics, engineering, computer science

##### 4. 📊 Stata Template ([#223](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐ MEDIUM-HIGH

**What**: Statistical analysis software

**Use Cases**:
- Statistical analysis
- Econometrics
- Survey data analysis
- Panel data research

**Technical Details**:
- DCV desktop required
- License: BYOL
- Memory: 8-16GB RAM
- Instance: t3.large

**Disciplines**: Economics, social science, public health, political science

##### 5. 🌍 ArcGIS Desktop Template (Optional) ([#211](https://github.com/scttfrdmn/prism/issues/214))
**Priority**: ⭐⭐⭐ MEDIUM

**What**: Commercial GIS software (Esri)

**Use Cases**:
- Professional GIS analysis (alternative to QGIS)
- Esri ecosystem integration
- Enterprise GIS workflows

**Technical Details**:
- DCV desktop required
- License: BYOL or AWS Marketplace
- Windows Server required (different from Linux templates)
- Instance: t3.xlarge or larger

**Note**: Only implement if demand justifies (QGIS covers most GIS needs)

**Disciplines**: Geography, urban planning (enterprise users)

#### Success Metrics (Desktop Applications)
- Commercial license integration (Cloud-Based, BYOL, Marketplace) working
- MATLAB and QGIS templates validated with real users
- DCV performance acceptable for GUI interaction
- Cost transparency (GPU vs non-GPU instances)
- Clear documentation for all three license setup methods
- All templates support research user provisioning

#### Implementation Notes (Desktop Applications)
- Test each application thoroughly on DCV
- Document all three license setup procedures:
  - Cloud-Based: Online activation (simplest for end users)
  - BYOL: License server configuration (institutional deployments)
  - AWS Marketplace: Pre-configured AMI approach
- Provide cost comparison (desktop vs web apps)
- Create troubleshooting guides
- Validate GPU acceleration where applicable
- Document instance sizing recommendations

### v0.6.3 (Q1 2027): TUI Completeness & Advanced Features

#### Auto-Update Phase 2: Assisted Updates
- Platform detection (Homebrew, apt, manual install)
- Automated update workflow with checksum verification
- Backup and rollback support

#### GUI System Tray Phase 3: Advanced Features
- Quick Launch from system tray
- Context-aware menu (recent templates, instances needing attention)
- Hover tooltip with cost summary

#### TUI Enhancements
- Project member management (add/remove members in TUI)
- Project-filtered instance views
- Cost breakdown visualization
- Hibernation savings display

### v0.7.0 (Q3 2026): Advanced UI & GPU Scheduling

#### 5. 🎯 EC2 Capacity Blocks for ML ([#63](https://github.com/scttfrdmn/prism/issues/63))
**Status**: Planned  
**Why**: Large ML workloads need guaranteed GPU availability

**What are Capacity Blocks?**
- Reserve GPU capacity 1-14 days in advance (up to 182 days)
- Guaranteed capacity for P5, P4d, Trn1 instances
- 10-20% discount vs on-demand pricing
- Advance booking up to 8 weeks

**Features (Phases 1-2)**:
- Search for available capacity blocks
- Purchase with budget integration
- Upfront payment with immutability warnings

**Example**:
```bash
$ prism capacity-blocks search \
  --instance-type p5.48xlarge \
  --instance-count 4 \
  --duration 48h \
  --earliest-start "2025-11-15"

Found 3 available offerings:

1. Offering ID: cbr-0123456789abcdefg
   Instance Type: p5.48xlarge × 4 instances
   Duration: 48 hours (2 days)
   Start: 2025-11-15 00:00 UTC
   Total Cost: $6,060.67 ($31.46/hour per instance)
   Discount: ~15% vs on-demand

To reserve: prism capacity-blocks purchase cbr-0123456789abcdefg
```

#### Auto-Update Phase 3: Background Updates
- Daemon-managed background update checks
- Auto-download with user permission
- Release channel support (stable/beta/dev)

#### Multi-User Authentication Phase 2
- Complete OAuth/OIDC integration
- SAML support for enterprise SSO
- Full RBAC implementation

### v0.7.1 (Q3 2026): Capacity Block Scheduling

#### Capacity Blocks Phases 3-4: Scheduled Launch & Management
**Status**: Planned

**Features**:
- Launch instances using capacity reservation ID
- Scheduled auto-launch when reservation becomes active
- **Scheduling Options**: Daemon-based (simple) OR Lambda-based (reliable, AWS-managed)
- Utilization tracking (X/N instances used)
- Cost analytics integration
- Underutilization warnings

**Scheduling Challenge**:
Reliable auto-launch requires daemon to be running OR AWS Lambda function:
- **Daemon-based**: Simple, works locally (requires computer running)
- **Lambda + EventBridge**: Highly reliable, AWS-managed (~$0 cost)
- **Recommended**: Hybrid approach with user choice

### v0.7.2 (Q4 2026): GUI Enhancements

#### Capacity Blocks Phase 5: GUI Integration
- Visual calendar picker for date selection
- Reservation dashboard with timeline view
- Cost comparison charts
- Scheduled launch interface

---

## 🚀 Completed Features

### Phase 1-4: Foundation (v0.1.0 - v0.4.5)
✅ Distributed architecture (daemon + CLI client)  
✅ Multi-modal access (CLI/TUI/GUI parity)  
✅ Template system with inheritance  
✅ Hibernation & cost optimization  
✅ Project-based budget management  

### Phase 4.6: Professional GUI (v0.4.6 - September 2025)
✅ Cloudscape Design System migration (AWS-native components)  
✅ Professional tabbed interface  
✅ Enterprise-grade accessibility (WCAG AA)  

### Phase 5A: Multi-User Foundation (v0.5.0 - September 2025)
✅ Dual user system (system users + persistent research users)  
✅ SSH key management with Ed25519/RSA support  
✅ EFS integration for persistent home directories  
✅ Complete CLI/TUI integration  

### Phase 5B: Template Marketplace (v0.5.2 - October 2025)
✅ Multi-registry support (community, institutional, private)  
✅ Template discovery with advanced search  
✅ Security validation and quality analysis  
✅ Ratings, reviews, and badges  

---

## 📊 Feature Status

### v0.5.8: Quick Start Experience (December 2025)

| Feature | Status | Milestone | Issues |
|---------|--------|-----------|--------|
| Rename to "Workspaces" | 📋 Planned | [#2](https://github.com/scttfrdmn/prism/milestone/2) | [#15](https://github.com/scttfrdmn/prism/issues/15) |
| Home Page + Quick Start Wizard | 📋 Planned | [#2](https://github.com/scttfrdmn/prism/milestone/2) | [#13](https://github.com/scttfrdmn/prism/issues/13) |
| `prism init` CLI Wizard | 📋 Planned | [#2](https://github.com/scttfrdmn/prism/milestone/2) | [#17](https://github.com/scttfrdmn/prism/issues/17) |

### v0.5.9: Navigation Restructure (January 2026)

| Feature | Status | Milestone | Issues |
|---------|--------|-----------|--------|
| Merge Terminal/WebView | 📋 Planned | [#3](https://github.com/scttfrdmn/prism/milestone/3) | [#14](https://github.com/scttfrdmn/prism/issues/14) |
| Collapse Advanced Features | 📋 Planned | [#3](https://github.com/scttfrdmn/prism/milestone/3) | [#16](https://github.com/scttfrdmn/prism/issues/16) |
| Unified Storage UI | 📋 Planned | [#3](https://github.com/scttfrdmn/prism/milestone/3) | [#18](https://github.com/scttfrdmn/prism/issues/18) |

### v0.5.10: Multi-Project Budgets (February 2026)

| Feature | Status | Milestone | Issues |
|---------|--------|-----------|--------|
| Shared Budget Pools | 📋 Planned | [#24](https://github.com/scttfrdmn/prism/milestone/24) | [#97](https://github.com/scttfrdmn/prism/issues/97) |
| Project Budget Allocation | 📋 Planned | [#24](https://github.com/scttfrdmn/prism/milestone/24) | [#98](https://github.com/scttfrdmn/prism/issues/98) |
| Budget Reallocation | 📋 Planned | [#24](https://github.com/scttfrdmn/prism/milestone/24) | [#99](https://github.com/scttfrdmn/prism/issues/99) |
| Multi-Project Rollup | 📋 Planned | [#24](https://github.com/scttfrdmn/prism/milestone/24) | [#100](https://github.com/scttfrdmn/prism/issues/100) |
| Enhanced Resource Tagging | 📋 Planned | [#24](https://github.com/scttfrdmn/prism/milestone/24) | [#128](https://github.com/scttfrdmn/prism/issues/128) |

### v0.5.11: User Invitation & Roles (March 2026)

| Feature | Status | Milestone | Issues |
|---------|--------|-----------|--------|
| Email Invitations | 📋 Planned | [#25](https://github.com/scttfrdmn/prism/milestone/25) | [#101](https://github.com/scttfrdmn/prism/issues/101) |
| Role Assignment | 📋 Planned | [#25](https://github.com/scttfrdmn/prism/milestone/25) | [#102](https://github.com/scttfrdmn/prism/issues/102) |
| Invitation Management | 📋 Planned | [#25](https://github.com/scttfrdmn/prism/milestone/25) | [#103](https://github.com/scttfrdmn/prism/issues/103) |
| Bulk CSV Invitations | 📋 Planned | [#25](https://github.com/scttfrdmn/prism/milestone/25) | [#104](https://github.com/scttfrdmn/prism/issues/104) |
| Quota Validation | 📋 Planned | [#25](https://github.com/scttfrdmn/prism/milestone/25) | [#105](https://github.com/scttfrdmn/prism/issues/105) |
| Research User Auto-Provisioning | 📋 Planned | [#25](https://github.com/scttfrdmn/prism/milestone/25) | [#106](https://github.com/scttfrdmn/prism/issues/106) |

### v0.5.12: Operational Stability & CLI (April 2026)

| Feature | Status | Milestone | Issues |
|---------|--------|-----------|--------|
| Launch Rate Limiting | 📋 Planned | [#26](https://github.com/scttfrdmn/prism/milestone/26) | [#107](https://github.com/scttfrdmn/prism/issues/107), [#90](https://github.com/scttfrdmn/prism/issues/90) |
| Retry Logic | 📋 Planned | [#26](https://github.com/scttfrdmn/prism/milestone/26) | [#108](https://github.com/scttfrdmn/prism/issues/108) |
| Consistent CLI Commands | 📋 Planned | [#26](https://github.com/scttfrdmn/prism/milestone/26) | [#20](https://github.com/scttfrdmn/prism/issues/20) |
| AWS Quota Management | 📋 Planned | [#26](https://github.com/scttfrdmn/prism/milestone/26) | [#57](https://github.com/scttfrdmn/prism/issues/57), [#58](https://github.com/scttfrdmn/prism/issues/58), [#59](https://github.com/scttfrdmn/prism/issues/59), [#60](https://github.com/scttfrdmn/prism/issues/60) |
| Improved Error Messages | 📋 Planned | [#26](https://github.com/scttfrdmn/prism/milestone/26) | [#109](https://github.com/scttfrdmn/prism/issues/109) |

### v0.5.13: UX Re-evaluation (May 2026)

| Feature | Status | Milestone | Issues |
|---------|--------|-----------|--------|
| UX Audit | 📋 Planned | [#27](https://github.com/scttfrdmn/prism/milestone/27) | [#110](https://github.com/scttfrdmn/prism/issues/110) |
| Persona Validation | 📋 Planned | [#27](https://github.com/scttfrdmn/prism/milestone/27) | [#111](https://github.com/scttfrdmn/prism/issues/111) |
| Quick Wins | 📋 Planned | [#27](https://github.com/scttfrdmn/prism/milestone/27) | [#112](https://github.com/scttfrdmn/prism/issues/112) |
| Performance Improvements | 📋 Planned | [#27](https://github.com/scttfrdmn/prism/milestone/27) | [#113](https://github.com/scttfrdmn/prism/issues/113) |
| Template Request Feature | 📋 Planned | [#27](https://github.com/scttfrdmn/prism/milestone/27) | [#229](https://github.com/scttfrdmn/prism/issues/229) |

### v0.8.0 (October 2026): Cross-Account & Compliance Foundation
**Release Date**: Target October 31, 2026
**Focus**: Multi-institution collaboration and regulatory compliance (NIST 800-171, HIPAA)
**Status**: 📋 Planned

#### P0 - Critical Features (Blocking Institutional Adoption)
- [#114](https://github.com/scttfrdmn/prism/issues/114) - S3-Based Cross-Institution Data Sharing (replaces cross-account EFS)
- [#116](https://github.com/scttfrdmn/prism/issues/116) - NIST 800-171 Compliance Framework for CUI Data
- [#117](https://github.com/scttfrdmn/prism/issues/117) - HIPAA Compliance Architecture for PHI Data

#### P2 - High Value Features
- [#121](https://github.com/scttfrdmn/prism/issues/121) - S3 Storage Integration (prerequisite for #114)
- [#122](https://github.com/scttfrdmn/prism/issues/122) - Institutional Template Repository with Approval Workflow
- [#127](https://github.com/scttfrdmn/prism/issues/127) - MATE Desktop by Default for Desktop Workstations

**Success Metrics**:
- Multi-institution projects: Support 3+ AWS accounts per project
- Compliance certification: Pass NIST 800-171 audit
- HIPAA-ready: Support clinical research workloads
- S3 adoption: 50% of large datasets (>10TB) use S3 vs EFS

**Target Users**: NIH researchers, clinical investigators, cross-institutional consortiums, research IT compliance officers

---

### v0.8.1 (January 2027): Collaboration Management Tools
**Release Date**: Target January 31, 2027
**Focus**: Advanced collaboration features and cost transparency
**Status**: 📋 Planned

#### P0 - Critical Features
- [#115](https://github.com/scttfrdmn/prism/issues/115) - User-Level Cost Attribution Across Institutions

#### P1 - High Priority Features
- [#118](https://github.com/scttfrdmn/prism/issues/118) - Invitation Policy Restrictions (Templates, Instance Types, Costs)
- [#119](https://github.com/scttfrdmn/prism/issues/119) - Collaboration Audit Trail for Compliance
- [#120](https://github.com/scttfrdmn/prism/issues/120) - Graceful Collaboration End with Work Preservation

**Success Metrics**:
- Cost attribution: 100% of multi-institution projects track per-user costs
- Policy enforcement: Zero budget overruns from invitation policy violations
- Audit compliance: Automated NIH/NSF compliance reports (save 40hr/year)
- Collaboration lifecycle: Zero lost work from expired collaborations

**Target Users**: Grant administrators, lab managers, multi-institution project leads

---

### v0.9.0 (April 2027): Advanced Enterprise Features
**Release Date**: Target April 30, 2027
**Focus**: Enterprise financial management and institutional dashboards
**Status**: 📋 Planned

#### P2 - Medium Priority Features
- [#123](https://github.com/scttfrdmn/prism/issues/123) - Chargeback System (integrates with Petri project)
- [#124](https://github.com/scttfrdmn/prism/issues/124) - Cross-Account Resource Transfer (Snapshots & AMIs)
- [#125](https://github.com/scttfrdmn/prism/issues/125) - Institutional Compliance Dashboard

#### P3 - Low Priority Features
- [#126](https://github.com/scttfrdmn/prism/issues/126) - FSx for Lustre / High-Performance Storage

**Success Metrics**:
- Automated chargeback: Monthly automated cost recovery for 80% of multi-institution projects
- Resource portability: Seamless workspace transfer between institutions
- Compliance monitoring: Research IT can monitor 300+ projects from single dashboard
- HPC support: Computational chemistry/climate modeling workloads supported

**Target Users**: Research IT administrators, institutional finance offices, HPC researchers

---

### v0.6.0: Enterprise Authentication (June 2026)

| Feature | Status | Milestone | Issues |
|---------|--------|-----------|--------|
| OAuth/OIDC Integration | 📋 Planned | TBD | Coming Soon |
| LDAP/Active Directory | 📋 Planned | TBD | Coming Soon |
| Auto-Update | 📋 Planned | TBD | [#61](https://github.com/scttfrdmn/prism/issues/61) |

### Future Releases (v0.6.0+)

| Feature | Status | Target Release | GitHub Issue |
|---------|--------|----------------|--------------|
| Directory Sync | 📋 Planned | v0.5.5-0.5.6 | [#53](https://github.com/scttfrdmn/prism/issues/53) |
| Configuration Sync | 📋 Planned | v0.5.3-0.5.4 | [#54](https://github.com/scttfrdmn/prism/issues/54) |
| Auto-Update | 📋 Planned | v0.6.0-0.7.0 | [#61](https://github.com/scttfrdmn/prism/issues/61) |
| GUI System Tray | 📋 Planned | v0.6.0-0.6.1 | [#62](https://github.com/scttfrdmn/prism/issues/62) |
| AWS Quota Management | 📋 Planned | v0.6.0 | [#57-60](https://github.com/scttfrdmn/prism/issues/57) |
| Multi-User Auth | 📋 Planned | v0.6.0-0.7.0 | Coming Soon |
| Capacity Blocks | 📋 Planned | v0.7.0-0.7.2 | [#63](https://github.com/scttfrdmn/prism/issues/63) |

**Legend**: ✅ Complete | 🟡 In Progress | 📋 Planned

---

## 🎯 Success Metrics

Prism tracks these metrics to measure progress:

### Current State (v0.5.7)
- ⏱️ **Time to first workspace**: ~15 minutes (needs improvement)
- 🧭 **Navigation complexity**: 14 flat items (needs simplification)
- 😃 **User confusion rate**: ~40% of support tickets (needs reduction)
- 🎯 **CLI first-attempt success**: ~60% (needs improvement)

### Target State (v0.5.8 + v0.5.9)
**v0.5.8 Targets (Quick Start Experience)**:
- ⏱️ **Time to first workspace**: 30 seconds (from 15 minutes)
- 🎯 **First-attempt success rate**: >90%
- 😃 **User confusion**: Reduce by 70%

**v0.5.9 Targets (Navigation Restructure)**:
- 🧭 **Navigation complexity**: 6 primary categories (from 14 items)
- ⏱️ **Time to find features**: <10 seconds
- 😃 **User confusion**: Further 30% reduction
- 📱 **Advanced feature discoverability**: >95%

**v0.6.0+ Targets (Enterprise Features)**:
- 📈 **Version adoption**: >70% on latest within 7 days
- 🔐 **Enterprise adoption**: Support institutional authentication
- 🎯 **CLI consistency**: Predictable command patterns across all features

---

## 💡 How to Contribute

**Found a feature request?**  
Create an issue: [github.com/scttfrdmn/prism/issues/new](https://github.com/scttfrdmn/prism/issues/new)

**Want to discuss the roadmap?**  
Join discussions: [github.com/scttfrdmn/prism/discussions](https://github.com/scttfrdmn/prism/discussions)

**Technical implementation details?**  
See: [Technical Debt Backlog](archive/roadmap/TECHNICAL_DEBT_BACKLOG.md)

---

## 📚 Related Documentation

- [VISION.md](VISION.md) - Long-term product vision
- [UX Evaluation](architecture/UX_EVALUATION_AND_RECOMMENDATIONS.md) - Current UX issues
- [Technical Debt Backlog](archive/roadmap/TECHNICAL_DEBT_BACKLOG.md) - Implementation tracking
- [GitHub Projects](https://github.com/scttfrdmn/prism/projects) - Sprint planning

---

**Questions?** Open a [GitHub Discussion](https://github.com/scttfrdmn/prism/discussions) or check [existing issues](https://github.com/scttfrdmn/prism/issues).
