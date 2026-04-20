# Collaboration Quickstart Guide

**🎯 Goal**: Get your collaborator access to your research environment in 5 minutes

**👥 Audience**: Researchers adding non-technical collaborators who only need web browser access

---

## Overview

This is the simplified guide for quickly adding collaborators to your Prism instance. They'll access RStudio or Jupyter through their web browser - no software installation needed!

**What you'll do**:
1. Connect to your instance (1 minute)
2. Create a user account (2 minutes)
3. Share the web link (1 minute)
4. Collaborator logs in and starts working! (1 minute)

**Need more details?** See the [comprehensive multi-user setup guide](MULTI_USER_INSTANCE_SETUP.md).

---

## For Instance Owners: Adding a Collaborator

### Step 1: Connect to Your Instance

From your terminal, connect to the running instance:

```bash
prism workspace connect my-research-env
```

Replace `my-research-env` with your instance name (see `prism workspace list` for all instances).

### Step 2: Create User Account

On the instance, run:

```bash
sudo adduser collaborator
```

You'll be prompted to:
- **Enter password** (twice): Choose something secure but memorable
- **Full name** (optional): Press Enter to skip
- **Other info** (optional): Press Enter to skip
- **Confirm**: Type `Y` and press Enter

**Example**:
```
Adding user `collaborator' ...
Enter new UNIX password: ********
Retype new UNIX password: ********
Full Name []: Maria Rodriguez
Room Number []:
Work Phone []:
Home Phone []:
Other []:
Is the information correct? [Y/n] Y
```

### Step 3: Add to Web Service Access

**For RStudio Server**:
```bash
sudo usermod -aG rstudio-users collaborator
```

**For Jupyter Lab**:
```bash
# No special setup needed! All users can access Jupyter
```

### Step 4: Find Your Instance IP Address

```bash
# Exit from the instance (Ctrl+D or type 'exit')
# Then run:
prism workspace list
```

Look for your instance and note the **Public IP** (e.g., `54.123.45.67`).

### Step 5: Share Access Details

Send your collaborator an email like this:

---

**Email Template**:

> Hi Maria,
>
> I've set up access to our research environment. Here's how to get started:
>
> **Web Link**: http://54.123.45.67:8787
> **Username**: collaborator
> **Password**: [the password you set]
>
> Just open that link in your web browser, enter the username/password, and you'll see the RStudio interface. All our shared data is in the `/shared/projects` folder.
>
> Let me know if you have any issues!

---

**That's it!** Your collaborator can now access the environment.

---

## For Collaborators: Getting Started

### Step 1: Open the Link

Click the web link shared by your colleague (e.g., `http://54.123.45.67:8787`).

**Note**: Your browser may show a "Not Secure" warning. This is normal for research environments. Click "Advanced" → "Proceed" to continue.

### Step 2: Login

You'll see a login screen:

- **Username**: The username they gave you (e.g., `collaborator`)
- **Password**: The password they set for you

Click **Sign In**.

### Step 3: You're In!

**RStudio Server**:
- You'll see the RStudio IDE interface
- Left: Console for running R commands
- Top-right: Environment with data objects
- Bottom-right: Files, plots, help
- Just like RStudio on your desktop!

**Jupyter Lab**:
- You'll see the Jupyter interface
- Left: File browser
- Right: Notebook area
- Click "New" → "Python 3" to create a notebook

### Step 4: Find Shared Data

**In RStudio**:
1. Bottom-right pane → **Files** tab
2. Navigate to `/shared/projects`
3. You'll see files your colleague shared

Or in the R console:
```r
setwd("/shared/projects")
list.files()
```

**In Jupyter**:
1. Left sidebar file browser
2. Navigate to `/shared/projects`
3. Double-click files to open

### Step 5: Start Working!

You can now:
- Run R/Python code
- Create new scripts/notebooks
- Access shared datasets
- Install packages (they'll only affect your account)
- Save your work (auto-saved!)

**Tips**:
- Use clear file names: `maria-analysis-2024-01-15.R`
- Save frequently (though auto-save is enabled)
- Check `/shared/projects` regularly for updates from teammates
- Let your colleague know before editing their files

---

## Quick Troubleshooting

### Can't Login

**Problem**: "Invalid username or password"

**Solutions**:
1. Double-check username (case-sensitive!)
2. Double-check password
3. Ask instance owner to reset password:
   ```bash
   sudo passwd collaborator
   ```

### Can't See Shared Files

**Problem**: `/shared/projects` folder is empty or missing

**Solutions**:
1. Make sure you're looking in the right place:
   - RStudio: Files pane → Navigate to root `/` → `shared` → `projects`
   - Jupyter: File browser → Navigate to `/shared/projects`

2. Ask instance owner to check permissions:
   ```bash
   ls -la /shared/projects
   sudo usermod -aG researchers collaborator
   ```

3. **Important**: Logout and login again after owner adds you to group

### Page Won't Load

**Problem**: Browser shows "Can't reach this page" or "Connection refused"

**Solutions**:
1. Verify the URL is correct (check for typos)
2. Confirm the instance is running (ask owner to check `prism workspace list`)
3. Try a different browser (Chrome, Firefox, Safari)
4. Check your internet connection
5. Ask owner to verify web service is running:
   ```bash
   sudo systemctl status rstudio-server
   # or
   sudo systemctl status jupyter
   ```

---

## Common Questions

### Do I need to install anything?

**No!** Everything runs in your web browser. Just click the link and login.

### What browsers work?

Chrome, Firefox, Safari, and Edge all work great. Use an updated version for best performance.

### Can I work offline?

No, you need internet access to connect to the instance. But the instance is always available (24/7).

### Will I lose my work if I close the browser?

No! Your work is saved on the remote instance. When you login again, everything will be there.

### Can I install R/Python packages?

Yes! Packages you install are available to your account only. They won't affect other users.

**RStudio**:
```r
install.packages("ggplot2")
```

**Jupyter**:
```python
!pip install pandas
```

### How do I save files?

Files are automatically saved when you:
- RStudio: Click **Save** icon or Ctrl+S
- Jupyter: **File** → **Save** or Ctrl+S

Save to `/shared/projects` so teammates can access them.

### Can multiple people edit the same file?

Technically yes, but be careful:
- **Best practice**: Coordinate who's editing what
- **Alternative**: Use version control (Git) for complex projects
- **Safe option**: Each person works on their own files, merge later

### What if I need more help?

Contact your instance owner or see the [full multi-user setup guide](MULTI_USER_INSTANCE_SETUP.md) for detailed instructions.

---

## Next Steps

Once you're comfortable with the basics:

### Learn RStudio Shortcuts

- `Ctrl+Enter`: Run current line
- `Ctrl+Shift+C`: Comment/uncomment code
- `Ctrl+L`: Clear console
- `Tab`: Auto-complete

### Organize Your Work

Create a personal folder in `/shared/projects`:
```r
dir.create("/shared/projects/maria")
setwd("/shared/projects/maria")
```

### Explore Installed Packages

**RStudio**:
```r
installed.packages()[,c(1,3)]
```

**Jupyter**:
```python
!pip list
```

### Share Your Findings

When you have results to share:
1. Save plots as PNG/PDF
2. Export data as CSV
3. Share file path with teammates
4. Consider writing R Markdown / Jupyter Notebook reports

---

## Advanced: Setting Up SSH Access (Optional)

If you're comfortable with command-line tools, you can also access via SSH for more powerful workflows.

### On Your Computer

1. **Generate SSH key** (if you don't have one):
   ```bash
   ssh-keygen -t ed25519 -C "your.email@example.com"
   ```
   Press Enter to accept defaults.

2. **View your public key**:
   ```bash
   cat ~/.ssh/id_ed25519.pub
   ```
   Copy this entire line.

3. **Send to instance owner**: Email them your public key.

### Instance Owner Adds Your Key

```bash
prism workspace connect my-research-env
sudo su - collaborator
mkdir -p ~/.ssh && chmod 700 ~/.ssh
echo "your-public-key-here" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
exit
```

### You Can Now SSH

```bash
ssh collaborator@54.123.45.67
```

**Benefits**:
- Use your favorite local editor (VS Code, Vim, Emacs)
- Transfer files with `scp` or `rsync`
- Run long-running jobs in `tmux` or `screen`
- Access full command-line tools

---

## Related Guides

- **[Multi-User Instance Setup](MULTI_USER_INSTANCE_SETUP.md)** - Comprehensive guide with advanced topics
- **[Custom AMI Workflow](CUSTOM_AMI_WORKFLOW.md)** - Save your configured environment
- **[AWS Setup Guide](AWS_SETUP_GUIDE.md)** - Initial setup for instance owners

---

## Support

Need help getting started?

- **Ask your instance owner** - They set up the environment and can help!
- **GitHub Issues**: [Report problems](https://github.com/scttfrdmn/prism/issues)
- **Full documentation**: [Complete docs](https://scttfrdmn.github.io/prism/)

---

**Last Updated**: January 2026 | **Version**: 1.0
