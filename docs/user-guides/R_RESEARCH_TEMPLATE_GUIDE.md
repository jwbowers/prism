# R Research Full Stack Template Guide

## Overview

The **R Research Full Stack** template provides a complete, production-ready R research environment designed for collaborative data analysis. It includes everything needed for modern R-based research: RStudio Server (web-based IDE), Quarto for publishing, LaTeX for documents, Python integration, and essential data science tools.

**Perfect for:**
- Collaborative research projects with remote team members
- Publishing research papers with R Markdown/Quarto
- Mixed R and Python data science workflows
- Teaching and coursework (web-based access)
- Multi-user research environments

## Quick Start (⏱️ 5 minutes)

### 1. Launch the Environment

```bash
# Launch R research environment
prism launch r-research-full-stack my-r-project

# Wait for installation (this takes ~15-20 minutes first time)
# The template installs R, RStudio Server, Quarto, LaTeX, and 40+ R packages
```

**Installation Components:**
- R 4.4.2 with base packages (2-3 min)
- RStudio Server 2024.12.0 (1-2 min)
- Quarto 1.6.33 (1 min)
- TeX Live 2024 full distribution (5-7 min)
- Essential R packages: tidyverse, rmarkdown, knitr, ggplot2, devtools, etc. (10-15 min)
- Python 3.12 + Jupyter Lab (2-3 min)
- Database clients and utilities (1-2 min)

### 2. Access RStudio Server

```bash
# Get connection info
prism connect my-r-project

# Output shows:
# RStudio Server: http://54.123.45.67:8787
# Username: researcher
# Password: [your instance password]
```

### 3. Open RStudio in Your Browser

1. Navigate to the RStudio Server URL (port 8787)
2. Login with your credentials
3. Start analyzing data in R!

**You now have access to:**
- Full RStudio IDE in your browser
- All tidyverse packages pre-installed
- Quarto for document publishing
- LaTeX for PDF generation
- Git integration for version control

## What's Included

### Core R Environment
- **R 4.4.2**: Latest stable R release
- **RStudio Server 2024.12.0**: Web-based IDE on port 8787
- **40+ R packages pre-installed**:
  - **Data manipulation**: dplyr, tidyr, purrr, stringr
  - **Visualization**: ggplot2, plotly, viridis, scales
  - **Publishing**: rmarkdown, knitr, bookdown, blogdown, xaringan
  - **Tables**: gt, gtsummary
  - **Database**: DBI, RSQLite, RPostgres, RMySQL
  - **Web**: httr, jsonlite, xml2, rvest, shiny
  - **Development**: devtools, usethis, testthat, pkgdown
  - **Python integration**: reticulate
  - **Utilities**: here, fs, glue, lubridate, forcats

### Publishing Tools
- **Quarto 1.6.33**: Modern scientific publishing system
- **Pandoc 3.5**: Universal document converter
- **TeX Live 2024**: Full LaTeX distribution with all packages
  - pdflatex, xelatex, lualatex
  - All fonts and packages for academic publishing
- **Document tools**: ghostscript, pdftk, ImageMagick

### Python Integration
- **Python 3.12**: Latest Python for mixed workflows
- **Jupyter Lab**: Interactive notebooks (port 8888)
- **Scientific packages**: numpy, pandas, matplotlib, seaborn, scikit-learn, scipy
- **Reticulate**: Seamless R-Python integration in RStudio

### Database Support
- **PostgreSQL client**: Connect to PostgreSQL databases
- **MySQL client**: Connect to MySQL/MariaDB databases
- **SQLite**: Embedded database for local data
- **R database packages**: DBI, RSQLite, RPostgres, RMySQL

### Development Tools
- **Git 2.43+**: Version control with LFS support for large files
- **Text editors**: vim, nano, emacs-nox
- **Terminal multiplexers**: tmux, screen for persistent sessions
- **System monitoring**: htop, tree, ncdu

### Data Processing Utilities
- **csvkit**: Command-line CSV tools
- **jq**: JSON processor
- **xmlstarlet**: XML toolkit
- **Compression tools**: zip, unzip, bzip2, p7zip
- **File transfer**: rsync, wget, curl

## Usage Examples

### Example 1: Create and Render Quarto Document

```bash
# SSH into your instance
prism connect my-r-project

# Create new Quarto project
cd ~/documents
quarto create-project my-analysis --type manuscript

# Edit the document
cd my-analysis
nano index.qmd

# Render to PDF
quarto render

# The PDF is now in _output/my-analysis.pdf
```

### Example 2: Mixed R and Python Workflow

In RStudio Server (http://your-ip:8787):

```r
# Install reticulate if not already installed
# library(reticulate)

# Use Python from R
library(reticulate)
use_python("/usr/bin/python3")

# Import Python libraries
pd <- import("pandas")
np <- import("numpy")

# Create DataFrame in Python, use in R
py_data <- pd$DataFrame(list(
  x = np$array(c(1, 2, 3, 4, 5)),
  y = np$array(c(2, 4, 6, 8, 10))
))

# Convert to R data frame
r_data <- py_to_r(py_data)

# Use ggplot2 for visualization
library(ggplot2)
ggplot(r_data, aes(x = x, y = y)) +
  geom_point() +
  geom_smooth(method = "lm")
```

### Example 3: Connect to PostgreSQL Database

```r
# Load database packages
library(DBI)
library(RPostgres)

# Connect to database
con <- dbConnect(
  RPostgres::Postgres(),
  host = "your-db-host.amazonaws.com",
  port = 5432,
  dbname = "research_data",
  user = "researcher",
  password = Sys.getenv("DB_PASSWORD")
)

# Query data
data <- dbGetQuery(con, "
  SELECT *
  FROM experiments
  WHERE experiment_date > '2024-01-01'
")

# Analyze with tidyverse
library(dplyr)
summary_stats <- data %>%
  group_by(treatment) %>%
  summarise(
    mean_response = mean(response),
    sd_response = sd(response),
    n = n()
  )

# Disconnect
dbDisconnect(con)
```

### Example 4: Create Interactive Shiny Dashboard

```r
# Create new Shiny app
library(shiny)
library(ggplot2)
library(dplyr)

# app.R
ui <- fluidPage(
  titlePanel("Research Data Explorer"),

  sidebarLayout(
    sidebarPanel(
      selectInput("variable", "Variable:",
                 choices = c("Sepal.Length", "Sepal.Width",
                           "Petal.Length", "Petal.Width")),
      sliderInput("bins", "Number of bins:",
                 min = 5, max = 50, value = 30)
    ),

    mainPanel(
      plotOutput("distPlot")
    )
  )
)

server <- function(input, output) {
  output$distPlot <- renderPlot({
    ggplot(iris, aes_string(x = input$variable)) +
      geom_histogram(bins = input$bins, fill = "steelblue") +
      theme_minimal() +
      labs(title = paste("Distribution of", input$variable))
  })
}

shinyApp(ui = ui, server = server)

# Run the app
# Access at http://your-ip:3838
```

### Example 5: Generate Research Paper with Quarto

Create `paper.qmd`:

```markdown
---
title: "My Research Paper"
author: "Researcher Name"
date: today
format:
  pdf:
    toc: true
    number-sections: true
    colorlinks: true
bibliography: references.bib
---

## Introduction

This paper analyzes...

## Methods

```{r}
#| label: setup
#| include: false
library(tidyverse)
library(knitr)
library(gt)
```

## Results

```{r}
#| label: fig-analysis
#| fig-cap: "Distribution of experimental results"

data <- read_csv("data/results.csv")

ggplot(data, aes(x = treatment, y = response)) +
  geom_boxplot() +
  theme_minimal()
```

## Conclusion

Our findings show...

## References
```

Render:
```bash
quarto render paper.qmd
```

## Collaboration Setup

### Add a Collaborator

```bash
# SSH into your instance
prism connect my-r-project

# Create user account
sudo adduser colleague
sudo usermod -aG sudo colleague

# Set RStudio Server password (same as Linux password)
# User can now login at http://your-ip:8787
```

### Share Project Access

```r
# In RStudio Server, set project permissions
# File > New Project > Existing Directory
# Select ~/projects/shared-analysis

# Set directory permissions for collaboration
system("chmod -R 775 ~/projects/shared-analysis")
system("chgrp -R sudo ~/projects/shared-analysis")
```

### Concurrent Work

Multiple users can:
- Work simultaneously in RStudio Server (separate sessions)
- Share R projects in `/home/shared/` or specific project directories
- Use Git for version control and collaboration
- Access the same data files in shared directories

## Performance Optimization

### Create Custom AMI for Faster Launch

After first launch and full installation (15-20 minutes):

```bash
# Create AMI from configured instance
prism ami create my-r-project --name "R Research Full Stack AMI"

# Future launches from AMI: < 2 minutes!
prism launch --ami ami-abc123def456 quick-r-instance
```

**Benefits:**
- Launch time: 15-20 minutes → < 2 minutes
- All packages pre-installed and cached
- Custom configurations preserved
- Share AMI with colleagues or students

See [Custom AMI Workflow Guide](CUSTOM_AMI_WORKFLOW.md) for details.

### Instance Sizing Recommendations

**Small Projects (< 5GB data)**:
```bash
prism launch r-research-full-stack my-project --size M
# Instance: t3.medium (2 vCPU, 4GB RAM)
# Cost: ~$0.05/hour
```

**Medium Projects (5-20GB data)**:
```bash
prism launch r-research-full-stack my-project --size L
# Instance: t3.large (2 vCPU, 8GB RAM)
# Cost: ~$0.10/hour
```

**Large Projects (20GB+ data, complex models)**:
```bash
prism launch r-research-full-stack my-project --size XL
# Instance: t3.xlarge (4 vCPU, 16GB RAM)
# Cost: ~$0.20/hour
```

### Cost Optimization with Hibernation

```bash
# Hibernate when not in use (preserves all state)
prism hibernate my-r-project

# Resume when needed (< 2 minutes)
prism resume my-r-project

# Savings: ~90% reduction in compute costs
```

## Troubleshooting

### RStudio Server Not Accessible

**Check if service is running:**
```bash
prism connect my-r-project
sudo systemctl status rstudio-server

# If not running, start it
sudo systemctl start rstudio-server
```

**Check firewall (security group):**
```bash
# Verify port 8787 is open
prism list my-r-project
# Look for "Ports: [22, 8787, 8888]"
```

**Can't login to RStudio Server:**
- Username is the system username (default: `researcher`)
- Password is the Linux user password
- Set/reset password: `sudo passwd researcher`

### R Package Installation Fails

**Insufficient memory:**
```bash
# Launch larger instance
prism launch r-research-full-stack my-project --size L
```

**Missing system dependencies:**
```bash
# Install additional dev libraries
sudo apt-get install -y libgdal-dev libproj-dev
```

**Install R package with dependencies:**
```r
# In RStudio Server or R console
install.packages("sf", dependencies = TRUE)
```

### Quarto Render Fails

**LaTeX errors:**
```bash
# Verify TeX Live installation
which pdflatex
pdflatex --version

# If needed, reinstall
sudo apt-get install --reinstall texlive-full
```

**Missing Quarto:**
```bash
# Verify installation
quarto --version

# If needed, reinstall
wget https://github.com/quarto-dev/quarto-cli/releases/download/v1.6.33/quarto-1.6.33-linux-amd64.deb
sudo apt-get install -y ./quarto-1.6.33-linux-amd64.deb
```

### Python/Jupyter Integration Issues

**Jupyter not found:**
```bash
# Verify installation
which jupyter
jupyter --version

# If needed, reinstall
pip3 install --upgrade jupyter jupyterlab
```

**Reticulate can't find Python:**
```r
library(reticulate)
use_python("/usr/bin/python3", required = TRUE)
py_config()
```

### Database Connection Issues

**PostgreSQL connection fails:**
```bash
# Test connection from command line
psql -h your-db-host -U username -d database_name

# Check security groups allow outbound connections
# Check database firewall allows your instance IP
```

## Advanced Features

### Using Git for Version Control

```bash
# Configure Git
git config --global user.name "Your Name"
git config --global user.email "you@example.com"

# Initialize repository
cd ~/projects/my-analysis
git init
git add .
git commit -m "Initial commit"

# Connect to GitHub (or GitLab, Bitbucket)
git remote add origin https://github.com/yourusername/my-analysis.git
git push -u origin main
```

### Large File Support with Git LFS

```bash
# Track large data files with LFS
cd ~/projects/my-analysis
git lfs track "*.csv"
git lfs track "*.rds"
git lfs track "*.RData"

# Add and commit
git add .gitattributes
git add data/large-file.csv
git commit -m "Add large data file"
git push
```

### Running R Scripts in Background

```bash
# Run R script in background
nohup Rscript my_analysis.R > output.log 2>&1 &

# Check progress
tail -f output.log

# Find process
ps aux | grep Rscript
```

### Scheduling R Scripts with Cron

```bash
# Edit crontab
crontab -e

# Run script daily at 2 AM
0 2 * * * /usr/bin/Rscript /home/researcher/scripts/daily_analysis.R >> /home/researcher/logs/daily.log 2>&1
```

## Best Practices

### Project Organization

```
~/projects/my-research/
├── data/
│   ├── raw/           # Original, immutable data
│   ├── processed/     # Cleaned, processed data
│   └── external/      # External reference data
├── R/                 # R scripts and functions
├── notebooks/         # Jupyter/R notebooks for exploration
├── reports/           # Quarto/RMarkdown reports
├── figures/           # Generated figures and plots
├── results/           # Analysis results
├── docs/              # Documentation
├── .gitignore
├── README.md
└── my-research.Rproj  # RStudio project file
```

### Data Management

1. **Keep raw data immutable**: Never modify original data files
2. **Document data processing**: Use R Markdown/Quarto notebooks
3. **Use relative paths**: Use the `here` package for portable code
4. **Version control code, not data**: Use Git LFS for large data files
5. **Back up regularly**: Use `rsync` or cloud storage

### Reproducible Research

```r
# Use renv for package management
install.packages("renv")
renv::init()           # Initialize project environment
renv::snapshot()       # Save package versions
renv::restore()        # Restore exact package versions

# Document session info
sessionInfo()
```

## Resources

### Official Documentation
- **RStudio**: https://www.rstudio.com/
- **Quarto**: https://quarto.org/
- **Tidyverse**: https://www.tidyverse.org/
- **R for Data Science**: https://r4ds.had.co.nz/
- **Jupyter**: https://jupyter.org/

### Prism Documentation
- [Multi-User Instance Setup](MULTI_USER_INSTANCE_SETUP.md) - Detailed collaboration guide
- [Custom AMI Workflow](CUSTOM_AMI_WORKFLOW.md) - Create reusable AMIs for faster launches
- [Template Format](TEMPLATE_FORMAT.md) - Customize templates
- [Community Template Guide](../development/COMMUNITY_TEMPLATE_GUIDE.md) - Contribute templates

### Getting Help
- **Prism Issues**: https://github.com/scttfrdmn/prism/issues
- **RStudio Community**: https://community.rstudio.com/
- **Stack Overflow**: Tag questions with `[r]`, `[rstudio]`, `[quarto]`

## FAQ

**Q: How long does the initial launch take?**
A: First launch: 15-20 minutes (installs everything). Create an AMI after first launch, then future launches take < 2 minutes.

**Q: Can multiple users work simultaneously?**
A: Yes! Each user gets their own RStudio Server session. Add users with `sudo adduser username`.

**Q: What's the difference between R Markdown and Quarto?**
A: Quarto is the next-generation of R Markdown with better multi-language support, consistent syntax, and enhanced features. Both are included.

**Q: Can I use R packages not pre-installed?**
A: Yes! Install any CRAN, Bioconductor, or GitHub package: `install.packages("package")` or `devtools::install_github("user/package")`.

**Q: How do I share my environment with colleagues?**
A: Create an AMI, share the AMI ID, or use template inheritance to create a customized version.

**Q: Does this work on ARM instances?**
A: Currently x86_64 only. ARM support planned for future versions.

**Q: Can I customize the template?**
A: Yes! Copy the template to `~/.prism/templates/my-r-env.yml` and modify as needed. See [Template Format](TEMPLATE_FORMAT.md).

## Related Templates

- **R Research Minimal**: Lightweight R environment (R + RStudio only)
- **Python ML**: Python-focused data science environment
- **Ultimate Research Workstation**: Multi-language research platform

## Version History

- **v1.0.0** (January 2026): Initial release
  - R 4.4.2 + RStudio Server 2024.12.0
  - Quarto 1.6.33 + TeX Live 2024
  - Python 3.12 + Jupyter Lab
  - 40+ pre-installed R packages
  - Full collaboration support

---

**Template**: `r-research-full-stack`
**Last Updated**: January 16, 2026
**Version**: 1.0.0
