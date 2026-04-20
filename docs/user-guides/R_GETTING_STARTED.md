# Getting Started with R on Prism

This tutorial walks you through Prism's R template family — from a minimal scripting environment
all the way to a published Shiny app — with practical R workflows at each step. By the end
you'll have a reproducible research environment and a custom AMI that launches in under 2 minutes.

**Prerequisites**: Prism installed, AWS credentials configured, `prism init` completed.

---

## The R Template Family

Prism's R templates form two inheritance trees. Choose your path before launching:

```
R Base (Ubuntu 24.04)          ← command-line R, no IDE
 ├── R + RStudio Server         ← web IDE, interactive work       ⭐ start here
 │    └── R Research Publishing Stack  ← + Quarto, LaTeX, Python
 └── R Shiny Server             ← share analyses as web apps

R Research Full Stack           ← monolithic: RStudio + Quarto + Python, independent
```

| Template | Best for | Launch time | Disk | Instance |
|----------|----------|-------------|------|----------|
| R Base | Scripts, CI, building blocks | ~5 min | 20 GB | any |
| R + RStudio Server | Interactive analysis, teaching | ~8 min | 25 GB | t3.medium |
| R Research Full Stack | Complete lab environment | ~15 min | 80 GB | m7i.xlarge |
| R Research Publishing Stack | Papers, reproducible reports | ~18 min | 80 GB | m7i.xlarge |
| R Shiny Server | Sharing dashboards with others | ~10 min | 30 GB | t3.medium |

---

## Part 1: Your First R Workspace

**R + RStudio Server** is the right starting point for most researchers. It gives you a full
web-based IDE without the weight of a complete publishing stack.

### Launch

```bash
prism workspace launch r-rstudio-server my-analysis
```

Prism provisions the instance, installs R 4.4+, RStudio Server, and all base packages. Watch the
output — it takes about 8 minutes. When it completes you'll see:

```
✅ Instance ready: my-analysis
   Public IP: 54.123.45.67
   Services:
     RStudio Server: http://54.123.45.67:8787
```

### Connect

**SSH** (works immediately — uses your SSH key):

```bash
prism workspace connect my-analysis          # opens SSH terminal
```

**RStudio Server** requires a system password for the `researcher` user. Templates that
include RStudio generate a random password during provisioning and print it in the launch
log. Look for lines like:

```
=============================================
  RStudio Server Login Credentials
  Username: researcher
  Password: <random-password>
=============================================
```

If you missed the password or need to reset it, SSH in and run:

```bash
sudo passwd researcher
```

Then open `http://<public-ip>:8787` and log in with username `researcher` and the password
you set.

### Verify your environment

In the RStudio console:

```r
R.version.string     # should show R 4.4.x
installed.packages()[, "Package"]  # tidyverse, rmarkdown, knitr, devtools already here
```

---

## Part 2: Installing Packages Efficiently

The templates use **Posit Package Manager** — a binary package repository for Ubuntu 24.04.
Packages install in seconds instead of minutes because they come pre-compiled.

### Set Posit Package Manager as your default

In RStudio, run once per session or add to `~/.Rprofile`:

```r
options(repos = c(CRAN = "https://packagemanager.posit.co/cran/__linux__/noble/latest"))
```

This is already configured for the initial install — packages you add later will also use it.

### Install packages the fast way

```r
# Examples: pre-built binaries, no compilation
install.packages("lme4")       # ~3 seconds vs ~4 minutes from source
install.packages("brms")       # Bayesian modeling
install.packages("targets")    # Pipeline toolkit
```

Compare: on a standard CRAN mirror without binaries, `lme4` compiles C++ code and takes several
minutes. With Posit Package Manager it's a simple download.

### Bioconductor packages

```r
if (!require("BiocManager")) install.packages("BiocManager")
BiocManager::install("DESeq2")
```

---

## Part 3: Reproducible Environments with renv

`renv` locks your package versions so collaborators and future-you get the same environment.

### Initialize renv in a project

```r
# In RStudio: File > New Project > New Directory, or in the console:
setwd("~/projects/my-study")
renv::init()
```

`renv` creates:
- `renv.lock` — the exact version of every package
- `.Rprofile` — loads renv automatically when you open the project
- `renv/` — local package library (not shared with other projects)

### Snapshot after installing packages

```r
install.packages(c("lme4", "emmeans", "ggplot2"))
renv::snapshot()    # writes versions to renv.lock
```

### Share with collaborators

Send them your project (including `renv.lock`). They run:

```r
renv::restore()     # installs exactly your versions
```

### Recommended workflow

```r
# Start of session
renv::status()      # check if lock file is current

# After adding packages
renv::snapshot()

# Before sharing / archiving
renv::status()
git add renv.lock
git commit -m "lock package versions"
```

---

## Part 4: Cost-Effective Workflows — Hibernate and Resume

Cloud instances cost money while idle. Prism's hibernation feature preserves your work and
stops billing (you pay only for EBS storage, ~$0.10/GB/month).

### Check your instance cost

```bash
prism workspace list                 # shows instance type and estimated hourly cost
```

A t3.medium (default for RStudio) costs ~$0.042/hour. Left running for a week = ~$7.

### Hibernate when you're done for the day

From your local terminal:

```bash
prism workspace hibernate my-analysis
```

RStudio Server saves its state. The instance stops. Your files, installed packages, and running R
sessions (if you save `.RData`) are preserved on the EBS volume.

### Resume the next day

```bash
prism workspace resume my-analysis
```

In ~2 minutes, your instance is back at the same IP with all files intact.

### Save your R session before hibernating

In RStudio:

```r
save.image("~/projects/my-study/session.RData")
```

Or use RStudio's built-in "Save workspace" option. On resume:

```r
load("~/projects/my-study/session.RData")
```

### Auto-hibernation

The RStudio template auto-hibernates after **60 minutes** of idle (no active R processes,
no RStudio sessions). Adjust this in the instance settings if you run long background jobs.

---

## Part 5: Scaling Up — Full Research Environments

When your analysis outgrows the base RStudio environment, move to one of the full stacks.

### When to upgrade

| Situation | Template to use |
|-----------|-----------------|
| Need Quarto for papers/reports | R Research Publishing Stack |
| Need Python + R in same environment | R Research Full Stack or Publishing Stack |
| Need LaTeX for PDF output | R Research Publishing Stack |
| Need Jupyter notebooks | R Research Full Stack |
| Heavy computation (>4 GB data in memory) | Either full stack (m7i.xlarge = 16 GB RAM) |

### R Research Full Stack — the monolithic approach

A standalone environment with everything pre-installed:

```bash
prism workspace launch r-research-full-stack my-big-study
```

Launches on **m7i.xlarge** (4 vCPU, 16 GB RAM) by default — a non-burstable instance suitable
for sustained computation. Available at both ports 8787 (RStudio) and 8888 (Jupyter Lab).

```bash
# Access RStudio
open http://<ip>:8787

# Access Jupyter Lab
open http://<ip>:8888
```

Includes 80 GB disk — enough for medium-sized datasets and LaTeX.

### R Research Publishing Stack — the layered approach

Builds on top of R + RStudio Server, adding Quarto, full TeX Live, Python, and Jupyter:

```bash
prism workspace launch r-publishing-stack my-paper
```

Also launches on m7i.xlarge with 80 GB disk. This is the recommended template for
writing papers — Quarto handles both HTML and PDF output from the same `.qmd` source.

**Quick Quarto example** (in RStudio terminal):

```bash
# Create a new article
quarto create article my-paper

# Preview in browser
quarto preview my-paper.qmd

# Render to PDF (requires LaTeX — included)
quarto render my-paper.qmd --to pdf
```

### Mixed R + Python workflows

Both full stacks include Python 3.12 and Jupyter. Switch between languages in a single
Quarto document:

```r
# my-analysis.qmd
# ```{r}
library(reticulate)
use_python("/usr/bin/python3")
# ```

# ```{python}
import pandas as pd
df = pd.read_csv("data.csv")
# ```

# ```{r}
py$df |> as_tibble()   # access Python's df from R
# ```
```

---

## Part 6: Sharing Your Work with Shiny

Once your analysis is complete, use the **R Shiny Server** template to share it as an
interactive web app with colleagues or students.

### Launch a Shiny server

```bash
prism workspace launch r-shiny my-shiny-dashboard
```

Installs R, Shiny Server, and a sample demo app. Accessible at port 3838.

### Deploy your app

SSH in and copy your app directory:

```bash
prism workspace connect my-shiny-dashboard

# On the instance:
sudo mkdir -p /srv/shiny-server/my-app
sudo cp -r ~/projects/my-app/* /srv/shiny-server/my-app/
sudo chown -R shiny:shiny /srv/shiny-server/my-app
sudo systemctl restart shiny-server
```

Access at: `http://<ip>:3838/my-app/`

### Install additional R packages for your app

```bash
prism workspace connect my-shiny-dashboard
```

```r
# On the instance, as researcher
options(repos = c(CRAN = "https://packagemanager.posit.co/cran/__linux__/noble/latest"))
install.packages(c("shinydashboard", "plotly", "leaflet"))
sudo systemctl restart shiny-server
```

### Transfer an app from your analysis instance

If you developed the app on `my-analysis` (RStudio), get both IPs then transfer:

```bash
# Get IP addresses
prism workspace list

# Copy app from analysis instance to shiny instance
ANALYSIS_IP=<analysis-instance-ip>
SHINY_IP=<shiny-instance-ip>

ssh researcher@${ANALYSIS_IP} "tar czf /tmp/my-app.tar.gz ~/projects/my-app"
scp researcher@${ANALYSIS_IP}:/tmp/my-app.tar.gz /tmp/
scp /tmp/my-app.tar.gz researcher@${SHINY_IP}:/tmp/
ssh researcher@${SHINY_IP} "
  sudo mkdir -p /srv/shiny-server/my-app
  sudo tar xzf /tmp/my-app.tar.gz -C /srv/shiny-server/my-app --strip-components=2
  sudo chown -R shiny:shiny /srv/shiny-server/my-app
  sudo systemctl restart shiny-server
"
```

---

## Part 7: Lock In Your Environment with an AMI

After spending 15-90 minutes configuring an environment, save it as an AMI (Amazon Machine Image).
Future launches from that AMI take **under 2 minutes** instead of 15-90 minutes.

### When to create an AMI

- After installing your domain-specific packages (e.g., a full bioinformatics stack)
- After configuring your `.Rprofile`, SSH keys, and project structure
- Before a course or workshop (saves time for all participants)
- Any time you want to share a ready-to-use environment with collaborators

### Create an AMI

```bash
prism ami save my-analysis "R 4.4 Genomics - March 2026"
```

This takes ~5 minutes. The instance keeps running while the snapshot is taken.

### Launch from your AMI

```bash
prism workspace launch --ami "R 4.4 Genomics - March 2026" my-new-instance
```

Or in the GUI, select your saved AMI from the launch dialog.

### AMI naming tips

Include the date and key packages in the name:
- `"R 4.4 + Seurat 5 + DESeq2 - 2026-03"` — genomics stack
- `"R + Quarto + tinytex - Stats 510 Fall 2026"` — course environment
- `"R Shiny + leaflet + DT - Lab Dashboard"` — shared dashboard base

### List and manage AMIs

```bash
prism ami list                                  # all saved AMIs
prism ami status <ami-id>                       # details for a specific AMI
prism ami delete <ami-id>                       # remove when no longer needed
```

---

## Quick Reference

### Common commands

```bash
# Launch templates
prism workspace launch r-base-ubuntu24 my-scripts          # minimal, SSH only
prism workspace launch r-rstudio-server my-analysis        # RStudio web IDE
prism workspace launch r-research-full-stack my-lab        # full stack, m7i.xlarge
prism workspace launch r-publishing-stack my-paper         # + Quarto + LaTeX
prism workspace launch r-shiny my-dashboard                # Shiny Server

# Instance management
prism workspace list                                       # all instances + IPs
prism workspace connect my-analysis                        # SSH into instance
prism workspace hibernate my-analysis                      # stop + preserve
prism workspace resume my-analysis                         # restart from hibernate
prism workspace delete my-analysis                         # destroy (irreversible)

# AMI management
prism ami save my-analysis "name"                          # save environment
prism ami list                                             # list saved AMIs
prism workspace launch --ami "name" new-instance           # launch from AMI
prism ami delete <ami-id>                                  # remove AMI
```

### Access URLs

| Template | URL |
|----------|-----|
| R + RStudio Server | `http://<ip>:8787` |
| R Research Full Stack | `http://<ip>:8787` (RStudio) / `http://<ip>:8888` (Jupyter) |
| R Research Publishing Stack | `http://<ip>:8787` (RStudio) / `http://<ip>:8888` (Jupyter) |
| R Shiny Server | `http://<ip>:3838` |

**RStudio login**: Username `researcher`, password from provisioning log (or reset via `sudo passwd researcher` over SSH)

### Posit Package Manager URL

```r
options(repos = c(CRAN = "https://packagemanager.posit.co/cran/__linux__/noble/latest"))
```

Add to `~/.Rprofile` on your instance to make it permanent.

---

## Next Steps

- **Reference guide**: [R Research Template Guide](R_RESEARCH_TEMPLATE_GUIDE.md) — deep dive into
  the Full Stack template with worked examples (Quarto documents, R+Python, database connections)
- **Shared storage**: `prism volume create shared-data --size 100` — attach an EFS volume to share
  data between your analysis and Shiny instances
- **Collaboration**: Add colleagues as users with `sudo adduser colleague` on the instance; they
  log in at `http://<ip>:8787` with their own credentials
- **Cost visibility**: `prism budget` — see what your R instances are spending per day
