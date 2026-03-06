# Quick Start: R Research Environment for Non-Technical Users

## What You'll Get

A complete R research environment accessible through your web browser - no software installation needed!

- ✅ RStudio (just like RStudio Desktop, but in your browser)
- ✅ R with all data science packages pre-installed
- ✅ Document publishing with Quarto and LaTeX
- ✅ Version control with Git
- ✅ Python and Jupyter (bonus!)

## For Project Owners (The Person Setting This Up)

### Step 1: Launch the Workspace

```bash
# Basic launch (recommended)
prism launch r-research-full-stack chile-collab --size M --wait

# With cost-saving hibernation
prism launch r-research-full-stack chile-collab \
  --size M \
  --hibernation \
  --wait

# In a project (for budget tracking and collaboration)
prism project create "Chile Collaboration" \
  --owner "your-email@example.com" \
  --budget-limit 500 \
  --budget-period monthly

prism launch r-research-full-stack chile-collab \
  --size M \
  --project "Chile Collaboration" \
  --hibernation \
  --wait
```

**This will take 10-15 minutes.** Go get coffee!

### Step 2: Set Up Access

```bash
# Get the workspace IP address
prism workspace describe chile-collab

# You'll see output like:
# Name: chile-collab
# Public IP: 54.123.45.67
# RStudio URL: http://54.123.45.67:8787
```

```bash
# SSH to set password
prism workspace connect chile-collab

# Set password for researcher user
sudo passwd researcher
# Enter: YourSecurePassword123!

# Exit
exit
```

### Step 3: Invite Your Colleague

**Option A: Just send them the info**
```
Subject: R Research Environment Ready!

Hi [Colleague Name],

Your R research environment is ready to use!

🖥️  Access URL: http://54.123.45.67:8787

Login Credentials:
  Username: researcher
  Password: YourSecurePassword123!

Just open that link in your web browser (Chrome, Firefox, or Safari) and login. You'll see RStudio - it works exactly like RStudio Desktop, but in your browser. No software to install!

Everything is already installed:
- R 4.4.2 with tidyverse
- Quarto for making PDFs
- Git for version control
- All the packages you need

Your files are in /home/researcher/projects/

Questions? Just reply to this email.

Best,
[Your Name]
```

**Option B: Use Prism's project invitation system** (if using projects)
```bash
prism project invite "Chile Collaboration" \
  --email "colleague@university.cl" \
  --role member \
  --message "Join our R research project! No command line needed - just use RStudio in your browser."
```

They'll get an email with a link to the web dashboard where they can see the workspace and access RStudio.

## For Collaborators (The Person Who Got Invited)

### Accessing RStudio

1. **Open your web browser** (Chrome, Firefox, Safari, or Edge)

2. **Go to the URL your colleague sent you**
   - Example: `http://54.123.45.67:8787`

3. **Login**
   - Enter the username: `researcher`
   - Enter the password (your colleague will give you this)

4. **You're in!**
   - This is RStudio - it works exactly like RStudio Desktop
   - All your files are in the Files pane on the right
   - Create new R scripts: File → New File → R Script
   - Run code: Click "Source" or press Cmd+Enter (Mac) / Ctrl+Enter (Windows)

### First Steps in RStudio

**Create Your First Script:**
1. Click **File → New File → R Script**
2. Type some R code:
```r
# Load tidyverse
library(tidyverse)

# Load some data
data <- read_csv("data/mydata.csv")

# Make a plot
ggplot(data, aes(x = variable1, y = variable2)) +
  geom_point() +
  theme_minimal()
```
3. Click **"Source"** to run the whole script
4. Or select lines and press **Cmd+Enter** (Mac) / **Ctrl+Enter** (Windows)

**Working with Files:**
- Click **"Files"** tab (bottom right)
- Navigate to `/home/researcher/projects/`
- Click **"New Folder"** to create a project folder
- Upload data: Click **"Upload"** button

**Creating Reports:**
1. Click **File → New File → Quarto Document**
2. Fill in title and author
3. Write your analysis (mix text and R code)
4. Click **"Render"** to create a PDF

**Saving Your Work:**
- Everything saves automatically in RStudio
- Files are saved to the cloud workspace
- Your colleague can see your files (it's a shared workspace)

### Common Questions

**Q: Do I need to install anything on my computer?**
A: No! Everything runs in your browser. You just need internet access.

**Q: What if my IT department blocks certain websites?**
A: The workspace uses port 8787. If you can't access it, ask your IT to allow outbound connections to that port.

**Q: Can I work on this from my Mac at home and my Windows computer at work?**
A: Yes! Just open the same URL from any computer and login.

**Q: What happens if I close my browser?**
A: Your work is saved. Just open the URL again and login - everything will be there.

**Q: Can I download files to my computer?**
A: Yes! In the Files pane, check the boxes next to files, then click "More → Export"

**Q: Can I upload my data files?**
A: Yes! In the Files pane, click "Upload" and select your files.

**Q: My colleague and I both need to edit the same file. Can we?**
A: Not at the exact same time in the same file. But you can:
- Take turns editing
- Work on different files
- Use Git for version control (ask your colleague)

**Q: What if something goes wrong?**
A: Contact the person who set up the workspace. They can:
- Restart the workspace
- Reset your password
- Fix technical issues

**Q: How much does this cost?**
A: Ask the project owner. Typical cost is ~$60/month for the M size. Can be reduced with hibernation (~$20/month if only used part-time).

## For Both: Shared Workflows

### Organizing Your Project

Create this structure in `/home/researcher/projects/`:

```
my-project/
├── data/
│   ├── raw/           # Original data (don't edit!)
│   └── processed/     # Cleaned data
├── scripts/           # R scripts
├── analysis/          # Analysis files (.Rmd, .qmd)
├── figures/           # Generated plots
├── reports/           # Final reports/papers
└── README.txt         # Project notes
```

### Best Practices

1. **Use descriptive file names**
   - ✅ `analysis_climate_trends_2024.R`
   - ❌ `script1.R`

2. **Comment your code**
```r
# Load libraries
library(tidyverse)

# Read data from our field site
data <- read_csv("data/raw/field_measurements.csv")

# Calculate mean temperature by site
temp_summary <- data %>%
  group_by(site_id) %>%
  summarise(mean_temp = mean(temperature))
```

3. **Save intermediate results**
```r
# After cleaning data
write_csv(clean_data, "data/processed/clean_data.csv")

# After analysis
saveRDS(model_results, "results/model_output.rds")
```

4. **Create Quarto reports, not just scripts**
   - Reports mix text, code, and outputs
   - Easy to share with non-R users
   - Automatically generates PDFs

### Communication

**Leave notes for each other:**
Create a file called `NOTES.txt` in your project folder:

```
2024-12-15 (John): Uploaded new data file to data/raw/december_samples.csv
2024-12-16 (Maria): Analyzed December samples, see analysis/december_analysis.qmd
2024-12-17 (John): Great analysis! Can you add the seasonal comparison?
```

## For Project Owners: Maintenance

### Cost Monitoring
```bash
# Check spending
prism project status "Chile Collaboration"

# View detailed costs
prism billing report --project "Chile Collaboration" --month current
```

### Hibernation (Save Money)
```bash
# Hibernate when not in use (saves ~70% cost)
prism workspace hibernate chile-collab

# Resume before work session
prism workspace resume chile-collab

# Tell collaborator: "I'm hibernating the workspace to save costs.
# Let me know when you need it and I'll wake it up (takes 2-3 minutes)"
```

### Backups
```bash
# Create weekly backup
prism backup create chile-collab --name "weekly-$(date +%Y%m%d)"

# List backups
prism backup list chile-collab

# Restore if needed
prism restore chile-collab --from backup-id
```

### Troubleshooting

**Collaborator can't access RStudio:**
```bash
# Check workspace is running
prism workspace list

# Restart if needed
prism workspace stop chile-collab
prism workspace start chile-collab

# Check RStudio Server is running
prism workspace connect chile-collab
sudo systemctl status rstudio-server
sudo systemctl restart rstudio-server
```

**Forgot password:**
```bash
# Reset password
prism workspace connect chile-collab
sudo passwd researcher
```

**Need more power:**
```bash
# Upgrade to L size (more RAM/CPU)
prism workspace stop chile-collab
prism workspace resize chile-collab --size L
prism workspace start chile-collab
```

**Can't SSH to workspace:**
```bash
# SSH should "just work" with recent Prism versions
# Keys are automatically managed

# Try the built-in connect command (easiest)
prism workspace connect chile-collab

# Or with explicit key path
ssh -i ~/.ssh/prism-aws-default-key ubuntu@YOUR_IP

# If still failing, verify your local key exists
ls -la ~/.ssh/prism-aws-default-key*

# Prism auto-generates and syncs keys on workspace launch
# If keys get out of sync, just launch a new workspace
```

## Success Stories

### Example 1: Cross-Continental Collaboration
> "My colleague in Chile doesn't know the command line at all. With this setup, I launched the workspace, sent him the URL and password, and he was using RStudio in his browser within 5 minutes. He didn't have to install anything - his Mac is managed by IT and he can't install software. This solved that problem completely!" - Research Professor, USA

### Example 2: Teaching Remote Workshop
> "I needed to teach R to 20 graduate students, all with different computers (Mac, Windows, Chromebook). Instead of spending 2 hours on installation issues, I launched one workspace per student. They just opened a URL and had identical environments. Zero installation problems!" - Data Science Instructor

### Example 3: Field Research
> "Our team collects data in the field and analyzes it remotely. We have one shared workspace with our data pipeline. Team members in three countries all access the same RStudio environment. No more 'but it works on my computer!' problems." - Ecology Research Group

## Next Steps

Once you're comfortable with the basics:

1. **Learn Git for version control** - Prevents "final_FINAL_v3.R" filenames
2. **Explore Quarto** - Beautiful documents and presentations
3. **Try Shiny** - Interactive web applications from R
4. **Use Projects** - RStudio Projects keep work organized

See the full documentation: `docs/templates/R_RESEARCH_FULL_STACK.md`

## Getting Help

- **Prism Questions:** https://github.com/scttfrdmn/prism/issues
- **RStudio Help:** https://docs.posit.co/
- **R Help:** `?function_name` in R Console
- **Quarto Help:** https://quarto.org/docs/guide/
- **Tidyverse Help:** https://www.tidyverse.org/

Remember: Everyone starts as a beginner. The R community is friendly and helpful!
