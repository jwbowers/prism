---
template: home.html
title: Prism — Cloud Research Workstations
hide:
  - navigation
  - toc
---

<div class="hero">
  <div class="hero-inner">
    <img src="images/prism-icon.png" alt="Prism" class="hero-icon">
    <h1>Cloud research workstations<br><span class="hero-gradient">in seconds, not hours</span></h1>
    <p class="hero-sub">Pre-configured AWS environments for researchers, labs, and classrooms — launch with one command, stop paying when you're done.</p>
    <div class="hero-buttons">
      <a href="user-guides/GETTING_STARTED/" class="btn-hero-primary">Get Started</a>
      <a href="https://github.com/scttfrdmn/prism" class="btn-hero-secondary" target="_blank">View on GitHub</a>
    </div>
    <div class="install-block">
      <code>brew install scttfrdmn/prism/prism</code>
    </div>
  </div>
</div>

<div class="features-section">
  <div class="features-grid">
    <div class="feature-card">
      <span class="feature-icon">⚡</span>
      <h3>Launch in seconds</h3>
      <p>One command and you're running. Prism picks the right instance, configures storage, and installs your tools automatically.</p>
    </div>
    <div class="feature-card">
      <span class="feature-icon">💰</span>
      <h3>Cost control built in</h3>
      <p>Budget limits per project, idle detection, and automatic hibernation prevent surprise AWS bills.</p>
    </div>
    <div class="feature-card">
      <span class="feature-icon">🧪</span>
      <h3>Research-ready templates</h3>
      <p>Python ML, R/RStudio, genomics, bioinformatics — pre-configured with the tools your field actually uses.</p>
    </div>
    <div class="feature-card">
      <span class="feature-icon">🖥️</span>
      <h3>CLI, TUI, and GUI</h3>
      <p>Script with the CLI, navigate with the terminal UI, or use the desktop app. All three share the same backend.</p>
    </div>
    <div class="feature-card">
      <span class="feature-icon">👥</span>
      <h3>Built for labs and classes</h3>
      <p>Provision workstations for an entire class via invitation codes. Budget pools, project isolation, and usage tracking included.</p>
    </div>
    <div class="feature-card">
      <span class="feature-icon">🔒</span>
      <h3>Secure by default</h3>
      <p>SSH key management, per-project IAM scoping, and encrypted storage. Your data stays in your own AWS account.</p>
    </div>
  </div>
</div>

<div class="quick-start-section">
  <h2>From zero to running in minutes</h2>

```bash
# Install
brew install scttfrdmn/prism/prism

# Connect your AWS credentials
prism profile add

# Launch a research environment
prism workspace launch python-ml my-project

# Connect via SSH — ready in under 2 minutes
prism workspace ssh my-project
```

  <div class="quick-start-links">
    <a href="user-guides/GETTING_STARTED/">Full installation guide →</a>
    <a href="admin-guides/AWS_IAM_PERMISSIONS/">IAM permissions →</a>
  </div>
</div>

<div class="templates-section">
  <h2>Environment templates</h2>
  <div class="templates-grid">
    <div class="template-card"><span>🐍</span><div><strong>python-ml</strong><br>PyTorch, TensorFlow, Jupyter</div></div>
    <div class="template-card"><span>📊</span><div><strong>r-research</strong><br>R, RStudio Server, Bioconductor</div></div>
    <div class="template-card"><span>🧬</span><div><strong>genomics</strong><br>GATK, BWA, samtools, STAR</div></div>
    <div class="template-card"><span>🔬</span><div><strong>bioinformatics</strong><br>Conda, Snakemake, BioPython</div></div>
    <div class="template-card"><span>🤖</span><div><strong>deep-learning</strong><br>CUDA, cuDNN, GPU-ready PyTorch</div></div>
    <div class="template-card"><span>📡</span><div><strong>data-science</strong><br>Pandas, scikit-learn, DuckDB</div></div>
    <div class="template-card"><span>⚙️</span><div><strong>hpc-base</strong><br>MPI, OpenMP, GCC, CMake</div></div>
    <div class="template-card"><span>🌐</span><div><strong>+ community</strong><br><a href="user-guides/TEMPLATE_MARKETPLACE_USER_GUIDE/">Browse marketplace</a></div></div>
  </div>
</div>

<div class="downloads-section">
  <h2>Downloads</h2>
  <div class="downloads-grid">
    <div class="download-card">
      <strong> macOS Apple Silicon</strong>
      <a href="https://github.com/scttfrdmn/prism/releases/latest/download/prism-darwin-arm64.tar.gz" class="btn-download">Download</a>
    </div>
    <div class="download-card">
      <strong> macOS Intel</strong>
      <a href="https://github.com/scttfrdmn/prism/releases/latest/download/prism-darwin-amd64.tar.gz" class="btn-download">Download</a>
    </div>
    <div class="download-card">
      <strong>🪟 Windows</strong>
      <a href="https://github.com/scttfrdmn/prism/releases/latest/download/prism-windows-amd64.zip" class="btn-download">Download</a>
    </div>
    <div class="download-card">
      <strong>🐧 Linux</strong>
      <a href="https://github.com/scttfrdmn/prism/releases/latest/download/prism-linux-amd64.tar.gz" class="btn-download">Download</a>
    </div>
  </div>
  <p class="downloads-note">All releases and changelogs on the <a href="https://github.com/scttfrdmn/prism/releases" target="_blank">GitHub releases page</a>.</p>
</div>
