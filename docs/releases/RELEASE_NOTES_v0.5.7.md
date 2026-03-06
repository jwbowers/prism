# Prism v0.5.7 Release Notes

**Release Date**: October 26, 2025
**Release Type**: Feature Release - Template File Provisioning & Test Infrastructure
**Status**: ✅ **RELEASED**

## 🎯 **Release Focus**

v0.5.7 introduces **S3-backed template file provisioning** for multi-GB dataset distribution, binary deployment, and pre-trained model delivery, along with critical test infrastructure improvements for production-ready CI/CD.

---

## 🚀 **MAJOR NEW FEATURES**

### **📦 S3-Backed Template File Provisioning**
**Status**: ✅ **COMPLETE**

**What It Enables**:
- **Multi-GB dataset distribution** to research environments
- **Binary deployment** for licensed software and tools
- **Pre-trained model distribution** for ML/AI workflows
- **Large file handling** with multipart uploads (up to 5TB)
- **Data integrity** via MD5 checksum verification
- **Real-time progress tracking** for large transfers

**Example Usage**:
```yaml
# Template with file provisioning
files:
  - source: "s3://my-bucket/datasets/imagenet-10gb.tar.gz"
    destination: "/data/imagenet"
    required: true
    md5: "a1b2c3d4e5f6..."
    extract: true

  - source: "s3://my-bucket/models/bert-large-uncased.bin"
    destination: "/models/bert"
    required: false
    architecture: "x86_64"  # Conditional provisioning

  - source: "s3://my-bucket/configs/jupyter_config.py"
    destination: "/home/researcher/.jupyter/"
    cleanup_after_download: true
```

**Technical Capabilities**:
- ✅ S3 multipart transfer support (files up to 5TB)
- ✅ MD5 checksum verification for data integrity
- ✅ Progress tracking with real-time updates
- ✅ Conditional provisioning (architecture-specific files)
- ✅ Required vs optional files with graceful fallback
- ✅ Auto-cleanup from S3 after download
- ✅ Complete documentation: [TEMPLATE_FILE_PROVISIONING.md](../TEMPLATE_FILE_PROVISIONING.md)

**GitHub Issues**:
- [#64](https://github.com/scttfrdmn/prism/issues/64) - S3-backed file transfer with progress tracking
- [#31](https://github.com/scttfrdmn/prism/issues/31) - Template asset management

---

## 🐛 **CRITICAL FIXES**

### **✅ Test Infrastructure Stability**
**Status**: ✅ **COMPLETE**

**Issue #83 Regression Fix**:
- **Problem**: API tests hitting AWS and timing out (97.961s test suite)
- **Root Cause**: Test mode check happened after AWS manager creation
- **Solution**: Restructured test mode handling to bypass AWS calls entirely
- **Result**: 206x faster tests (97.961s → 0.463s)

**Data Race Fix**:
- **Problem**: Concurrent cache access in system_metrics.go
- **Root Cause**: Unprotected concurrent reads/writes to CPU cache
- **Solution**: Added sync.Mutex protection to cache structures
- **Result**: Zero race conditions detected

**Test Performance**:
- ✅ All smoke tests passing (8/8)
- ✅ 206x faster test execution
- ✅ Zero race conditions
- ✅ Production-ready CI/CD pipeline
- ✅ Fast developer feedback loop (<1 second)

**GitHub Issue**:
- [#83](https://github.com/scttfrdmn/prism/issues/83) - API Test Stability

---

## 🔧 **IMPROVEMENTS**

### **📝 Script Cleanup**
- Completed CloudWorkStation → Prism rename across 19+ script files
- Updated build, service, and package management scripts
- Verified documentation consistency
- Consistent branding across entire codebase

### **⬆️ Dependency Updates**
- **Wails v3.0.0-alpha.36**: Verified latest version
- **AWS SDK Updates**:
  - aws-sdk-go-v2: 1.39.3 → 1.39.4
  - aws-config: 1.31.13 → 1.31.15
  - aws-sts: 1.38.7 → 1.38.9
- All dependencies updated to latest compatible versions

### **🐛 GUI Version Check Fix**
- Fixed smoke test failure for GUI version extraction
- Wails binaries don't support `--version` flag
- Now checks GUI package.json directly
- Handles both cmd/prism-gui and cmd/prism-gui locations
- All version checks passing

---

## 📊 **Impact & Benefits**

### **For Researchers**:
- 🚀 **Dataset Distribution**: Share multi-GB datasets across research teams
- 🔬 **Model Deployment**: Pre-trained models available immediately on launch
- 📦 **Binary Distribution**: Licensed software pre-installed and ready
- ⚡ **Faster Setup**: Large files downloaded once, cached on instance

### **For Developers**:
- ✅ **Reliable CI/CD**: 206x faster tests enable rapid iteration
- 🐛 **Zero Race Conditions**: Production-ready concurrent code
- 🔍 **Fast Feedback**: <1 second test suite for quick development
- 📈 **Quality Assurance**: All smoke tests passing before every push

---

## 📚 **Documentation**

### **New Documentation**:
- [Template File Provisioning Guide](../TEMPLATE_FILE_PROVISIONING.md) - Complete S3 provisioning documentation
- [Release Notes v0.5.7](../releases/RELEASE_NOTES_v0.5.7.md) - This document

### **Updated Documentation**:
- [CHANGELOG.md](../../CHANGELOG.md) - Complete v0.5.7 changelog entry
- [ROADMAP.md](../ROADMAP.md) - Updated to reflect v0.5.7 completion
- Scripts and build documentation - Consistent Prism branding

---

## 🔄 **Migration Guide**

### **No Breaking Changes**
This release is fully backward compatible. Existing templates and workflows continue to work without changes.

### **New Features (Optional)**
To use template file provisioning:

1. Add `files` section to your template YAML
2. Upload files to S3 bucket
3. Configure S3 permissions (read-only recommended)
4. Launch template as normal

See [TEMPLATE_FILE_PROVISIONING.md](../TEMPLATE_FILE_PROVISIONING.md) for detailed instructions.

---

## 📦 **Installation**

### **macOS**
```bash
# Using Homebrew
brew tap scttfrdmn/prism
brew upgrade prism  # If already installed
# or
brew install prism
```

### **Windows**
```powershell
# Using Scoop
scoop update prism  # If already installed
# or
scoop install prism
```

### **Linux**
```bash
# Using Homebrew on Linux
brew upgrade prism  # If already installed
# or
brew install prism
```

### **Direct Download**
Download from [GitHub Releases](https://github.com/scttfrdmn/prism/releases/tag/v0.5.7):
- [macOS Intel (x86_64)](https://github.com/scttfrdmn/prism/releases/download/v0.5.7/prism-darwin-amd64.tar.gz)
- [macOS Apple Silicon (M1/M2)](https://github.com/scttfrdmn/prism/releases/download/v0.5.7/prism-darwin-arm64.tar.gz)
- [Windows (x86_64)](https://github.com/scttfrdmn/prism/releases/download/v0.5.7/prism-windows-amd64.zip)
- [Linux (x86_64)](https://github.com/scttfrdmn/prism/releases/download/v0.5.7/prism-linux-amd64.tar.gz)
- [Linux (ARM64)](https://github.com/scttfrdmn/prism/releases/download/v0.5.7/prism-linux-arm64.tar.gz)

---

## 🙏 **Contributors**

This release includes contributions and fixes from the Prism development team.

---

## 🔗 **Links**

- **GitHub Release**: [v0.5.7](https://github.com/scttfrdmn/prism/releases/tag/v0.5.7)
- **Full Changelog**: [CHANGELOG.md](../../CHANGELOG.md)
- **Roadmap**: [ROADMAP.md](../ROADMAP.md)
- **Documentation**: [docs/index.md](../index.md)
- **Issue Tracker**: [GitHub Issues](https://github.com/scttfrdmn/prism/issues)
- **Project Board**: [GitHub Projects](https://github.com/scttfrdmn/prism/projects)

---

## 🚀 **What's Next?**

**v0.5.8 and Beyond** - See [ROADMAP.md](../ROADMAP.md) for upcoming features:
- Commercial software template support
- Configuration sync system
- Advanced storage integration (FSx, S3 mount points)
- Template marketplace enhancements
