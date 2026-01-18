# R Research Full Stack Template

## Overview

**Template Name:** `r-research-full-stack`
**Type:** Monolithic (all-in-one)
**Purpose:** Complete R research environment for collaborative data science and scientific publishing

Perfect for research projects requiring web-based collaboration with non-technical colleagues who need zero command-line experience.

## What's Included

### Core Research Tools
- ✅ **R 4.4.2** - Latest R with full ecosystem
- ✅ **RStudio Server 2024.12.0** - Web-based IDE (port 8787)
- ✅ **Quarto 1.6.33** - Next-generation scientific publishing
- ✅ **LaTeX (TeX Live 2024 Full)** - Complete LaTeX distribution for PDF generation
- ✅ **Pandoc** - Universal document converter

### Programming Languages
- ✅ **Python 3.12** - For mixed R/Python workflows
- ✅ **Jupyter Lab** - Interactive Python notebooks (port 8888)

### Development Tools
- ✅ **Git + Git LFS** - Version control with large file support
- ✅ **Database clients** - PostgreSQL, MySQL, SQLite

### R Package Ecosystem (Pre-installed)

**Data Science Core:**
- `tidyverse` - Complete data science toolkit
- `dplyr`, `tidyr`, `readr` - Data manipulation
- `ggplot2`, `plotly` - Visualization
- `lubridate` - Date/time handling

**Document Publishing:**
- `rmarkdown`, `knitr` - Dynamic documents
- `bookdown` - Books and long documents
- `blogdown` - Websites and blogs
- `xaringan` - Presentations
- `gt`, `gtsummary` - Professional tables

**Development:**
- `devtools`, `usethis`, `pkgdown` - Package development
- `testthat` - Unit testing

**Data Access:**
- `httr`, `xml2`, `rvest` - Web scraping and APIs
- `jsonlite` - JSON handling
- `DBI`, `RSQLite`, `RPostgres`, `RMySQL` - Databases

**Interoperability:**
- `reticulate` - Python integration
- `shiny` - Interactive web apps

## Launch Instructions

### Basic Launch
```bash
# Launch with default settings (M size recommended)
prism workspace launch r-research-full-stack my-research --size M --wait

# Check launch progress
prism workspace list

# Get instance details
prism workspace describe my-research
```

### Advanced Launch Options
```bash
# Launch with hibernation for cost savings
prism workspace launch r-research-full-stack my-research \
  --size M \
  --hibernation \
  --wait

# Launch with project association
prism workspace launch r-research-full-stack my-research \
  --size M \
  --project "Chile Collaboration" \
  --wait

# Launch with spot instances (70% cost savings)
prism workspace launch r-research-full-stack my-research \
  --size L \
  --spot \
  --wait
```

## Size Recommendations

| Size | vCPU | RAM | Use Case | Monthly Cost* |
|------|------|-----|----------|---------------|
| **M** (Recommended) | 2 | 8GB | Most research work, light ML | ~$60 |
| **L** | 4 | 16GB | Heavy data analysis, medium datasets | ~$120 |
| **XL** | 8 | 32GB | Large datasets, complex models | ~$240 |

*Approximate costs for on-demand instances running 24/7. Use hibernation to reduce costs by ~70% when idle.

## Initial Setup

### 1. Wait for Provisioning
The initial launch takes **10-15 minutes** due to extensive package installation:
- R packages installation: ~8 minutes
- LaTeX (TeX Live Full): ~3 minutes
- RStudio Server + other tools: ~2 minutes

**Check status:**
```bash
# Watch launch progress
prism workspace launch r-research-full-stack my-research --size M --wait

# Or check status separately
prism workspace list
prism workspace describe my-research
```

### 2. Get Access Information
```bash
# Get instance IP and connection details
prism workspace describe my-research

# Output includes:
# - Public IP address
# - RStudio Server URL: http://YOUR_IP:8787
# - SSH connection info
```

### 3. Set Researcher Password
Before accessing RStudio Server, set a password:

```bash
# SSH to instance
prism workspace connect my-research

# Set password for researcher user
sudo passwd researcher
# Enter new password (e.g., "SecurePassword123!")

# Exit SSH
exit
```

### 4. Access RStudio Server
1. **Open browser:** Navigate to `http://YOUR_IP:8787`
2. **Login:**
   - Username: `researcher`
   - Password: (what you set in step 3)
3. **Start working!** Full R environment ready to use

## Testing Checklist

### Test 1: R and Packages
Open RStudio Server and run in the R Console:

```r
# Test tidyverse
library(tidyverse)

# Create a quick plot
iris %>%
  ggplot(aes(Sepal.Length, Petal.Length, color = Species)) +
  geom_point() +
  theme_minimal()

# Test data manipulation
mtcars %>%
  group_by(cyl) %>%
  summarise(avg_mpg = mean(mpg))

# Verify key packages are installed
installed_packages <- c("tidyverse", "rmarkdown", "knitr", "ggplot2",
                        "dplyr", "DBI", "reticulate", "shiny")
sapply(installed_packages, function(pkg) {
  if(require(pkg, character.only = TRUE)) {
    cat(sprintf("✓ %s: %s\n", pkg, packageVersion(pkg)))
    TRUE
  } else {
    cat(sprintf("✗ %s: NOT FOUND\n", pkg))
    FALSE
  }
})
```

### Test 2: Quarto Publishing
In RStudio: **File → New File → Quarto Document**

```yaml
---
title: "Test Document"
format: pdf
---

## Introduction
This is a test of Quarto + LaTeX.

## R Code
```{r}
summary(mtcars)
plot(mtcars$mpg, mtcars$hp)
```

## Conclusion
If this renders to PDF, everything works!
```

Click **"Render"** - should generate a PDF document.

**Test from R Console:**
```r
# Verify Quarto is available
system("quarto --version")  # Should show: 1.6.33

# Verify LaTeX is available
system("pdflatex --version")  # Should show TeX Live 2024
```

### Test 3: Git and Version Control
In RStudio Terminal or R Console:

```r
# Test Git
system("git --version")  # Should show git version

# Test Git LFS
system("git lfs version")  # Should show git-lfs version

# Configure git (first time only)
system("git config --global user.name 'Your Name'")
system("git config --global user.email 'your.email@example.com'")
```

**Create a test repository:**
```bash
# In RStudio Terminal
cd ~
mkdir test-repo && cd test-repo
git init
echo "# Test Repository" > README.md
git add README.md
git commit -m "Initial commit"
git log  # Should show your commit
```

### Test 4: Python Integration
Test Python and Jupyter:

```r
# In R Console
library(reticulate)

# Check Python configuration
py_config()

# Test Python from R
py_run_string("import sys; print(sys.version)")

# Test NumPy
py_run_string("import numpy as np; print(np.__version__)")

# Test pandas
py_run_string("import pandas as pd; print(pd.__version__)")
```

**Test Jupyter Lab (Optional):**
```bash
# In RStudio Terminal or SSH
jupyter lab --ip=0.0.0.0 --no-browser --port=8888

# Access in browser: http://YOUR_IP:8888
# Use token shown in terminal output
```

### Test 5: Database Connectivity
```r
# Test SQLite
library(DBI)
library(RSQLite)

# Create in-memory database
con <- dbConnect(RSQLite::SQLite(), ":memory:")

# Create test table
dbWriteTable(con, "mtcars", mtcars)

# Query
result <- dbGetQuery(con, "SELECT * FROM mtcars WHERE mpg > 25")
print(result)

# Cleanup
dbDisconnect(con)

# Test PostgreSQL and MySQL clients (if you have remote servers)
# library(RPostgres)
# library(RMySQL)
```

## Collaborative Workflow

### For Non-Technical Colleagues

**What They Need:**
- ✅ Web browser (Chrome, Firefox, Safari, Edge)
- ✅ Internet connection
- ✅ Login credentials (username/password)
- ❌ NO command line needed
- ❌ NO local R installation needed
- ❌ NO local software installation needed

**Access Steps:**
1. Open browser
2. Navigate to: `http://YOUR_IP:8787`
3. Login with credentials you provide
4. Start working in RStudio (just like RStudio Desktop!)

**Collaboration Features:**
- Multiple users can work on same instance (with different sessions)
- Shared file system: `/home/researcher/` accessible to all
- Git integration for version control
- Real-time file updates

### Setting Up Shared Storage

For true collaboration with shared files:

```bash
# Create EFS storage for project
prism storage create project-shared-data --type efs

# Attach to workspace
prism storage attach project-shared-data my-research

# Now accessible at: /mnt/efs/project-shared-data/
# All collaborators see the same files
```

### Project-Based Collaboration

```bash
# 1. Create project with budget
prism project create "Chile Collaboration" \
  --owner "your-email@example.com" \
  --budget-limit 500 \
  --budget-period monthly

# 2. Launch workspace in project
prism workspace launch r-research-full-stack chile-workspace \
  --size M \
  --project "Chile Collaboration" \
  --hibernation

# 3. Invite collaborator
prism project invite "Chile Collaboration" \
  --email "colleague@university.cl" \
  --role member \
  --message "Join our R research project! Access RStudio at http://IP:8787"

# 4. They get email → Click link → Access workspace → Work in RStudio
```

## Cost Optimization

### Hibernation (Recommended)
```bash
# Hibernate when not in use (saves ~70% cost)
prism workspace hibernate my-research

# Resume when needed (takes 2-3 minutes)
prism workspace resume my-research

# Schedule automatic hibernation
prism workspace schedule my-research \
  --hibernate-after-idle 2h \
  --hibernate-weekdays "18:00-08:00" \
  --hibernate-weekends
```

### Spot Instances
```bash
# Launch with spot for 70% savings (may be interrupted)
prism workspace launch r-research-full-stack my-research \
  --size M \
  --spot \
  --hibernation
```

### Rightsizing
```bash
# Start small, upgrade if needed
prism workspace launch r-research-full-stack my-research --size M

# Upgrade later if you need more power
prism workspace resize my-research --size L
```

## File Organization

The workspace includes pre-created directories:

```
/home/researcher/
├── projects/     # Research projects
├── data/         # Data files
├── notebooks/    # Jupyter notebooks
├── documents/    # Quarto/RMarkdown documents
├── scripts/      # R and Python scripts
└── WELCOME.txt   # Welcome guide
```

**Recommended Structure:**
```
/home/researcher/projects/my-project/
├── data/         # Project data
│   ├── raw/      # Original data
│   └── processed/  # Cleaned data
├── R/            # R scripts
├── analysis/     # Analysis notebooks
├── reports/      # Generated reports
├── figures/      # Plots and figures
└── README.md     # Project documentation
```

## Troubleshooting

### RStudio Server Won't Load
```bash
# SSH to instance
prism workspace connect my-research

# Check RStudio status
sudo systemctl status rstudio-server

# Restart if needed
sudo systemctl restart rstudio-server

# Check logs
sudo tail -50 /var/log/rstudio/rstudio-server/rserver.log
```

### Can't Login to RStudio
```bash
# Reset researcher password
prism workspace connect my-research
sudo passwd researcher
```

### Quarto/LaTeX Not Working
```r
# In R Console
# Verify Quarto
system("quarto check")

# Verify LaTeX
system("pdflatex --version")

# If missing, SSH and reinstall
```

### Python Integration Issues
```r
# Reset Python configuration
library(reticulate)
py_discover_config()

# Use specific Python
use_python("/usr/bin/python3.12")
```

### Out of Disk Space
```bash
# Check disk usage
prism workspace connect my-research
df -h

# Clean package caches
sudo apt-get clean
sudo apt-get autoremove -y

# Or resize volume
prism workspace resize-volume my-research --size 100GB
```

### SSH Connection Issues

**Problem:** `Permission denied (publickey)` when trying to SSH to workspace

**Cause:** The SSH key management system has been improved to automatically sync your local SSH keys with AWS. This issue should rarely occur with recent versions of Prism.

**Solution:**
```bash
# Verify your local SSH key exists
ls -la ~/.ssh/cws-aws-default-key*

# If missing, Prism will auto-generate on next launch
# The key is automatically uploaded to AWS during workspace launch

# Test SSH connection with explicit key
ssh -i ~/.ssh/cws-aws-default-key ubuntu@YOUR_IP

# Or use the workspace connect command (recommended)
prism workspace connect my-research
```

**Verify Key Fingerprints Match:**
```bash
# Check local key fingerprint
ssh-keygen -lf ~/.ssh/cws-aws-default-key.pub

# Check AWS key fingerprint
aws ec2 describe-key-pairs --key-names cws-aws-default-key --query 'KeyPairs[0].KeyFingerprint'

# Should match! If not, re-launch the workspace to force key sync
```

**Force Key Refresh:**
```bash
# If keys get out of sync (rare), just launch a new workspace
# The key management system will automatically delete old AWS keys
# and upload your current local key

# Or manually remove AWS key
aws ec2 delete-key-pair --key-name cws-aws-default-key

# Next workspace launch will re-upload automatically
```

**Profile-Specific Keys:**
If using Prism profiles, each profile can have its own SSH key:
```bash
# Check current profile
prism profile list

# Profile-specific keys are named: cws-{profile-name}-key
# Located at: ~/.ssh/cws-{profile-name}-key
```

## Maintenance

### Keep System Updated
```bash
# SSH to instance
prism workspace connect my-research

# Update system packages
sudo apt-get update && sudo apt-get upgrade -y

# Update R packages
# In R Console:
update.packages(ask = FALSE, checkBuilt = TRUE)
```

### Backup Your Work
```bash
# Create snapshot before major changes
prism snapshot create my-research --name "before-major-update"

# Create regular backups
prism backup create my-research --name "weekly-backup-$(date +%Y%m%d)"

# List backups
prism backup list

# Restore if needed
prism restore my-research --from backup-id
```

## Comparison: Full Stack vs Publishing Stack

| Feature | r-research-full-stack (This) | r-publishing-stack |
|---------|------------------------------|---------------------|
| **Approach** | Monolithic (all-in-one) | Stacked (layered inheritance) |
| **Launch Time** | 10-15 minutes | 12-18 minutes |
| **Inheritance** | Direct from Ubuntu 24.04 | R Base → RStudio → Publishing |
| **Customization** | Harder to modify | Easier to extend layers |
| **Use When** | You want everything now | You want modular approach |

**Recommendation:** Use `r-research-full-stack` (this template) for production work. Use `r-publishing-stack` if you're experimenting with template inheritance or want to customize layers.

## Advanced Topics

### Adding Custom R Packages
```r
# In RStudio
install.packages("packagename")

# For Bioconductor packages
if (!require("BiocManager")) install.packages("BiocManager")
BiocManager::install("packagename")

# For GitHub packages
devtools::install_github("user/repo")
```

### Using Quarto for Different Formats
```yaml
---
title: "My Analysis"
format:
  html: default
  pdf: default
  docx: default
---
```

Render all formats:
```r
quarto::quarto_render("document.qmd")
```

### Setting Up RStudio Projects
1. **File → New Project → New Directory**
2. Choose directory location
3. Check "Create a git repository"
4. Click "Create Project"

Now you have:
- Project file (`.Rproj`)
- Git repository
- Organized workspace

### Collaborative Git Workflow
```bash
# Clone existing repository
git clone https://github.com/user/repo.git
cd repo

# Work in RStudio, make changes

# Commit and push
git add .
git commit -m "Descriptive message"
git push

# Pull collaborator changes
git pull
```

## Performance Tips

1. **Use data.table for large datasets** (faster than dplyr)
2. **Enable parallel processing:**
   ```r
   library(future)
   plan(multisession, workers = 4)  # Use 4 cores
   ```
3. **Use RStudio Server performance settings:**
   - Tools → Global Options → General → Advanced
   - Reduce max memory and history size if needed

4. **Monitor resource usage:**
   ```r
   # In R
   memory.size()
   memory.limit()

   # In Terminal
   htop  # Interactive process viewer
   ```

## Support and Documentation

- **Quarto:** https://quarto.org/docs/guide/
- **RStudio Server:** https://docs.posit.co/ide/server-pro/
- **Tidyverse:** https://www.tidyverse.org/
- **R for Data Science:** https://r4ds.hadley.nz/
- **Prism Documentation:** https://github.com/scttfrdmn/prism/tree/main/docs

## Example Workflows

### Workflow 1: Data Analysis Report
```r
# 1. Load data
library(tidyverse)
data <- read_csv("data/raw/dataset.csv")

# 2. Clean and explore
data_clean <- data %>%
  filter(!is.na(value)) %>%
  mutate(category = as.factor(category))

summary(data_clean)

# 3. Visualize
ggplot(data_clean, aes(x = category, y = value)) +
  geom_boxplot() +
  theme_minimal()

# 4. Create Quarto report
# File → New File → Quarto Document
# Write analysis, include code chunks
# Click "Render" to generate PDF
```

### Workflow 2: Reproducible Research
```r
# Use here package for relative paths
library(here)

# Load data relative to project root
data <- read_csv(here("data", "raw", "experiment1.csv"))

# Run analysis
results <- data %>% analyze()

# Save results
write_csv(results, here("data", "processed", "results.csv"))

# Generate figures
plot <- create_plot(results)
ggsave(here("figures", "result_plot.pdf"), plot)
```

### Workflow 3: Mixed R/Python
```r
# R analysis
library(tidyverse)
library(reticulate)

# Load data in R
data_r <- read_csv("data.csv")

# Pass to Python for ML
py_run_string("
import pandas as pd
from sklearn.ensemble import RandomForestRegressor

# Get data from R
data_py = r.data_r

# Train model
model = RandomForestRegressor()
model.fit(X, y)
predictions = model.predict(X_test)
")

# Get predictions back in R
predictions_r <- py$predictions
```

## Template Maintenance

This template is maintained in the Prism community repository.

**Location:** `templates/community/r-research-full-stack.yml`

**To report issues or suggest improvements:**
1. Open issue: https://github.com/scttfrdmn/prism/issues
2. Include template name and version
3. Describe issue or enhancement request

**To contribute:**
1. Fork repository
2. Modify template
3. Test thoroughly
4. Submit pull request

## Version History

- **v1.0.0** (2024-12) - Initial release
  - R 4.4.2, RStudio Server 2024.12.0
  - Quarto 1.6.33, TeX Live 2024
  - Python 3.12, Jupyter Lab
  - Complete tidyverse and publishing packages
