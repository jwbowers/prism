# Multi-User Instance Setup Guide

## Overview

This guide explains how to add collaborators to running Prism instances for shared research environments. Perfect for team projects, teaching, workshops, and cross-institutional collaboration.

### Who Is This For?

- **Instance Owners**: Researchers who launched the instance and want to add collaborators
- **Collaborators**: Team members who need access to shared research environments
- **IT Administrators**: Staff setting up multi-user research environments

### What You'll Learn

- Adding users to running instances
- Setting up web-only access (RStudio Server, Jupyter)
- Configuring SSH access (optional)
- Creating shared workspaces
- Security best practices
- Troubleshooting common issues

## Use Cases

### 1. Basic Two-Person Collaboration
Researcher in US collaborates with colleague in Chile on R analysis:
- **Owner**: Launches R environment, adds collaborator
- **Collaborator**: Opens browser to RStudio Server, starts analysis
- **Result**: Both work on same datasets, see each other's changes

### 2. Teaching & Coursework
Professor sets up environment for 30 students:
- **Owner**: Launches instance, creates read-only student accounts
- **Students**: Access Jupyter notebooks via browser
- **Result**: Everyone has consistent environment, no installation needed

### 3. Conference Workshop
Workshop organizer serves 40 attendees:
- **Owner**: Pre-configures environment with example data
- **Attendees**: Get instant access via shared credentials
- **Result**: No setup time lost, workshop starts immediately

### 4. International Collaboration
Multi-timezone team shares analysis environment:
- **Owner**: Sets up 24/7 instance with shared data directory
- **Team**: Members in US, Europe, Asia access async
- **Result**: Work continues across time zones, no data transfer

### 5. Client Review
Consultant shares analysis with client:
- **Owner**: Launches instance with results
- **Client**: Browser-only access to review findings
- **Result**: Client sees live analysis, no data export needed

---

## Quick Start (5 Minutes)

### For Instance Owners

1. **Connect to your instance**:
   ```bash
   prism connect my-r-env
   ```

2. **Add a collaborator**:
   ```bash
   sudo adduser collaborator
   ```

3. **Share access details**:
   - **URL**: `http://<instance-public-ip>:8787` (RStudio) or `:8888` (Jupyter)
   - **Username**: `collaborator`
   - **Password**: (what you set during adduser)

### For Collaborators

1. Open browser to shared URL
2. Enter username and password
3. Start working!

---

## Detailed Setup Instructions

### Part 1: For Instance Owners (CLI Users)

#### Step 1: Connect to Your Instance

First, connect to the running instance where you'll add users:

```bash
prism connect my-r-env
```

Or using the instance ID:

```bash
prism connect i-0123456789abcdef0
```

#### Step 2: Create User Account

On the instance, create a new user account:

```bash
# Interactive user creation (recommended for first-time)
sudo adduser collaborator

# You'll be prompted for:
# - Password (enter twice)
# - Full name (optional)
# - Room number, work phone, etc. (optional, press Enter to skip)
```

**Alternative**: Automated user creation (for scripting):

```bash
# Create user non-interactively
sudo useradd -m -s /bin/bash collaborator

# Set password
echo "collaborator:SecurePassword123" | sudo chpasswd
```

#### Step 3: Grant Appropriate Permissions

**For RStudio Server access**:

```bash
# Add user to RStudio users group
sudo usermod -aG rstudio-users collaborator
```

**For Jupyter access**:

```bash
# No special group needed - all system users can access Jupyter
# Optionally limit to specific group:
sudo usermod -aG jupyter-users collaborator
```

**For administrative access** (use sparingly):

```bash
# Grant sudo privileges
sudo usermod -aG sudo collaborator
```

#### Step 4: Set Up Shared Workspace

Create a shared directory where all collaborators can work:

```bash
# Create shared project directory
sudo mkdir -p /shared/projects

# Create researchers group
sudo groupadd researchers

# Add users to group
sudo usermod -aG researchers $(whoami)
sudo usermod -aG researchers collaborator

# Set directory ownership and permissions
sudo chgrp researchers /shared/projects
sudo chmod 2775 /shared/projects

# Set default ACLs for new files
sudo setfacl -d -m g::rwx /shared/projects
```

**What this does**:
- `2775`: Set-GID bit ensures new files inherit group
- `setfacl`: New files automatically get group write permissions
- All researchers can read/write, others can't access

#### Step 5: Get Instance Public IP

Find the instance's public IP to share with collaborators:

```bash
# From your local machine (before connecting)
prism list

# Or from within the instance
curl -s http://169.254.169.254/latest/meta-data/public-ipv4
```

#### Step 6: Share Access Details

Provide your collaborator with:

**For RStudio Server**:
- **URL**: `http://54.123.45.67:8787`
- **Username**: `collaborator`
- **Password**: `SecurePassword123`
- **Shared directory**: `/shared/projects`

**For Jupyter Lab**:
- **URL**: `http://54.123.45.67:8888`
- **Username**: `collaborator`
- **Password**: `SecurePassword123`
- **Shared directory**: `/shared/projects`

---

### Part 2: For Collaborators (Web-Only Users)

#### Step 1: Open Your Browser

Navigate to the URL shared by the instance owner:

- **RStudio**: `http://54.123.45.67:8787`
- **Jupyter**: `http://54.123.45.67:8888`

**Note**: These are HTTP URLs (not HTTPS). Your browser may show a warning - this is expected for development environments. See [Security Considerations](#security-considerations) for production use.

#### Step 2: Login

Enter the credentials provided by the instance owner:

- **Username**: The username they created for you
- **Password**: The password they set

**RStudio Server**:
- Login screen appears automatically
- Enter credentials and click "Sign In"
- RStudio IDE loads in your browser

**Jupyter Lab**:
- Login screen appears automatically
- Enter credentials
- Jupyter interface loads

#### Step 3: Navigate to Shared Workspace

**In RStudio**:
1. Files pane → Navigate to `/shared/projects`
2. Or use console: `setwd("/shared/projects")`

**In Jupyter**:
1. File browser → Navigate to `/shared/projects`
2. Create new notebooks there
3. All team members see the same files

#### Step 4: Start Working!

You now have full access to:
- The RStudio IDE or Jupyter Lab interface
- All R/Python packages installed on the instance
- Shared data and analysis files
- Computing resources (CPU, RAM, GPU)

**Tips**:
- Save your work frequently
- Use meaningful file names (e.g., `sarah-analysis-2024-01-15.R`)
- Coordinate with team members to avoid editing same files simultaneously
- Check `/shared/projects` regularly for updates from collaborators

---

## Advanced Topics

### SSH Key Setup (Optional)

For technical collaborators who want command-line access:

#### On Collaborator's Machine

1. **Generate SSH key** (if you don't have one):
   ```bash
   ssh-keygen -t ed25519 -C "your.email@example.com"
   ```

2. **Copy your public key**:
   ```bash
   cat ~/.ssh/id_ed25519.pub
   ```

#### On Instance Owner's Machine

1. **Connect to instance**:
   ```bash
   prism connect my-r-env
   ```

2. **Switch to collaborator's account**:
   ```bash
   sudo su - collaborator
   ```

3. **Set up SSH directory**:
   ```bash
   mkdir -p ~/.ssh
   chmod 700 ~/.ssh
   ```

4. **Add collaborator's public key**:
   ```bash
   echo "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5..." >> ~/.ssh/authorized_keys
   chmod 600 ~/.ssh/authorized_keys
   ```

5. **Exit back to your account**:
   ```bash
   exit
   ```

#### Collaborator Can Now SSH

```bash
ssh collaborator@54.123.45.67
```

**Advantages of SSH access**:
- Use local terminal instead of web interface
- Transfer files with `scp` or `rsync`
- Run long-running jobs
- Access from command-line tools

---

## Security Considerations

### 1. Strong Passwords

**Minimum requirements**:
- At least 12 characters
- Mix of uppercase, lowercase, numbers, symbols
- Not based on dictionary words
- Different from other passwords

**Setting strong passwords**:
```bash
# Generate random password
openssl rand -base64 16

# Set user password
echo "collaborator:$(openssl rand -base64 16)" | sudo chpasswd
```

### 2. SSH Keys vs Passwords

**When to use SSH keys**:
- ✅ Technical collaborators who understand SSH
- ✅ Long-term collaborations
- ✅ Automated access (scripts, CI/CD)
- ✅ Higher security requirements

**When to use passwords**:
- ✅ Non-technical collaborators
- ✅ Web-only access (RStudio, Jupyter)
- ✅ Short-term collaborations (workshops, demos)
- ✅ Easier onboarding

### 3. User Permission Levels

**Regular user** (default, recommended):
```bash
# No sudo access
# Can only modify own files
sudo adduser collaborator
```

**Group member** (shared workspace):
```bash
# Can read/write shared directories
sudo usermod -aG researchers collaborator
```

**Administrator** (use sparingly):
```bash
# Full system access
sudo usermod -aG sudo collaborator
```

**Recommendation**: Start with regular user + group access. Grant sudo only when necessary.

### 4. Network Security

**Security groups**:
- Only open ports you need (22 for SSH, 8787 for RStudio, 8888 for Jupyter)
- Restrict source IPs if possible (e.g., only your institution's IP range)
- Use Prism's automatic security group configuration

**Check current ports**:
```bash
sudo netstat -tlnp | grep -E ':(22|8787|8888)'
```

### 5. Data Privacy Considerations

**For sensitive data**:
- ⚠️ Do NOT use plain HTTP in production
- ✅ Use VPN to encrypt all traffic
- ✅ Enable HTTPS with proper certificates
- ✅ Encrypt data at rest (EBS encryption)
- ✅ Review AWS compliance certifications (HIPAA, etc.)

**For public data**:
- ✅ Plain HTTP is acceptable for development
- ✅ Still use strong passwords
- ✅ Monitor for unauthorized access

### 6. Audit and Monitoring

**Track user activity**:
```bash
# See who's currently logged in
who

# View login history
last

# Monitor active processes by user
ps aux | grep collaborator
```

**Review file access**:
```bash
# Recent file modifications in shared directory
find /shared/projects -mtime -7 -ls
```

---

## Shared Workspace Best Practices

### Directory Structure

Organize shared workspace for clarity:

```
/shared/
├── projects/
│   ├── project-alpha/
│   │   ├── data/          # Raw data (read-only)
│   │   ├── analysis/      # R/Python scripts
│   │   ├── results/       # Output files
│   │   └── docs/          # Documentation
│   └── project-beta/
├── data/
│   └── reference-datasets/  # Shared datasets
└── software/
    └── custom-tools/        # Shared scripts/tools
```

**Setting up**:
```bash
# Create structure
sudo mkdir -p /shared/{projects,data,software}
sudo chgrp -R researchers /shared
sudo chmod -R 2775 /shared

# Make data read-only
sudo chmod 2755 /shared/data
```

### File Naming Conventions

**Good naming**:
- `alice-exploratory-analysis-2024-01-15.R`
- `bob-regression-model-v2.py`
- `team-meeting-notes-2024-01-15.md`

**Avoid**:
- `analysis.R` (whose? when?)
- `script1.py` (what does it do?)
- `temp.txt` (will be deleted?)

### Coordination Strategies

**For small teams (2-5 people)**:
- Use descriptive file names
- Add comments in code: `# Alice: working on this section 2024-01-15`
- Quick Slack/email before editing shared files

**For larger teams (6+ people)**:
- Consider Git for version control
- Use branches for different analyses
- Schedule who works when (if needed)

**Git setup** (optional):
```bash
# In shared directory
cd /shared/projects/project-alpha
git init
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Make commits visible to all
git config core.sharedRepository group
```

---

## Application-Specific Configuration

### RStudio Server

#### Custom R Library Paths

Allow each user to install their own packages:

```bash
# Add to ~/.Rprofile
.libPaths(c("/home/collaborator/R/library", .libPaths()))
```

#### Shared R Library

Install packages once for all users:

```bash
# As admin user
sudo R
install.packages("tidyverse", lib="/usr/local/lib/R/site-library")
```

#### RStudio Server Configuration

**Session timeout** (in /etc/rstudio/rsession.conf):
```
session-timeout-minutes=60
```

**User limits** (in /etc/rstudio/rserver.conf):
```
rsession-memory-limit-mb=4096
rsession-stack-limit-mb=16
```

### Jupyter Lab

#### Custom Kernels

Each user can install their own Python environments:

```bash
# Create virtual environment
python3 -m venv ~/myenv

# Activate it
source ~/myenv/bin/activate

# Install packages
pip install pandas numpy jupyter

# Register kernel
python -m ipykernel install --user --name=myenv
```

#### Shared Jupyter Extensions

Install extensions for all users:

```bash
sudo pip install jupyterlab-git
sudo jupyter labextension install @jupyterlab/git
```

#### Jupyter Configuration

**Increase output limit** (in ~/.jupyter/jupyter_notebook_config.py):
```python
c.NotebookApp.iopub_data_rate_limit = 10000000
```

---

## Troubleshooting

### Issue: Can't Login to RStudio/Jupyter

**Symptoms**:
- "Invalid username or password"
- Login page refreshes without error

**Checks**:
1. **Verify user exists**:
   ```bash
   id collaborator
   ```

2. **Test password**:
   ```bash
   su - collaborator
   ```

3. **Check service status**:
   ```bash
   # RStudio
   sudo systemctl status rstudio-server

   # Jupyter
   sudo systemctl status jupyter
   ```

4. **Review logs**:
   ```bash
   # RStudio
   sudo tail /var/log/syslog | grep rstudio

   # Jupyter
   journalctl -u jupyter -n 50
   ```

**Solutions**:
- Reset password: `sudo passwd collaborator`
- Restart service: `sudo systemctl restart rstudio-server`
- Check group membership: `groups collaborator`

### Issue: Can't Access Shared Directory

**Symptoms**:
- "Permission denied" when opening `/shared/projects`
- Can see directory but can't create files

**Checks**:
1. **Verify group membership**:
   ```bash
   groups collaborator
   # Should show: collaborator researchers
   ```

2. **Check directory permissions**:
   ```bash
   ls -ld /shared/projects
   # Should show: drwxrwsr-x root researchers
   ```

3. **Test file creation**:
   ```bash
   sudo su - collaborator
   touch /shared/projects/test.txt
   ```

**Solutions**:
- Add to group: `sudo usermod -aG researchers collaborator`
- Fix permissions: `sudo chmod 2775 /shared/projects`
- Fix ownership: `sudo chgrp researchers /shared/projects`
- **Important**: User must logout and login again for group changes to take effect

### Issue: SSH Connection Refused

**Symptoms**:
- `ssh: connect to host 54.123.45.67 port 22: Connection refused`
- SSH works for owner but not collaborator

**Checks**:
1. **Verify SSH service**:
   ```bash
   sudo systemctl status ssh
   ```

2. **Check security group**:
   ```bash
   # From local machine
   prism list
   # Look for security group rules allowing port 22
   ```

3. **Test SSH key**:
   ```bash
   cat ~/.ssh/authorized_keys
   # Should contain collaborator's public key
   ```

4. **Check SSH logs**:
   ```bash
   sudo tail /var/log/auth.log
   ```

**Solutions**:
- Restart SSH: `sudo systemctl restart ssh`
- Fix authorized_keys permissions: `chmod 600 ~/.ssh/authorized_keys`
- Verify public key format (should be single line)
- Check AWS security group allows port 22 from collaborator's IP

### Issue: RStudio/Jupyter Service Not Running

**Symptoms**:
- Browser shows "Connection refused" or "Can't reach this page"
- Port is not responding

**Checks**:
1. **Verify service status**:
   ```bash
   sudo systemctl status rstudio-server
   sudo systemctl status jupyter
   ```

2. **Check if port is listening**:
   ```bash
   sudo netstat -tlnp | grep 8787  # RStudio
   sudo netstat -tlnp | grep 8888  # Jupyter
   ```

3. **Review recent logs**:
   ```bash
   sudo journalctl -u rstudio-server -n 100
   sudo journalctl -u jupyter -n 100
   ```

**Solutions**:
- Start service: `sudo systemctl start rstudio-server`
- Enable auto-start: `sudo systemctl enable rstudio-server`
- Check configuration: `sudo rstudio-server verify-installation`
- Verify AWS security group allows port 8787/8888

### Issue: Collaborator Sees Wrong Files

**Symptoms**:
- Collaborator can't see files owner created
- Files appear with wrong ownership

**Checks**:
1. **Check file ownership**:
   ```bash
   ls -la /shared/projects/
   ```

2. **Verify working directory**:
   ```bash
   # In RStudio
   getwd()

   # In Jupyter
   import os; os.getcwd()
   ```

**Solutions**:
- Fix ownership: `sudo chown :researchers /shared/projects/*`
- Make sure everyone works in `/shared/projects`
- Set default directory in RStudio: `setwd("/shared/projects")`

---

## Example Scenarios

### Scenario 1: Biology Lab with 3 Researchers

**Setup**:
```bash
# Instance owner launches R environment
prism launch "R Research Environment (Simplified)" bio-lab

# Connect and add team members
prism connect bio-lab

# Add users
sudo adduser alice
sudo adduser bob
sudo adduser charlie

# Create lab group
sudo groupadd biolab
sudo usermod -aG biolab $(whoami)
sudo usermod -aG biolab alice
sudo usermod -aG biolab bob
sudo usermod -aG biolab charlie

# Create shared workspace
sudo mkdir -p /shared/biolab/{data,analysis,results}
sudo chgrp -R biolab /shared/biolab
sudo chmod -R 2775 /shared/biolab

# Add to RStudio
sudo usermod -aG rstudio-users alice
sudo usermod -aG rstudio-users bob
sudo usermod -aG rstudio-users charlie
```

**Team members access**:
- Alice: `http://<ip>:8787` (username: alice)
- Bob: `http://<ip>:8787` (username: bob)
- Charlie: `http://<ip>:8787` (username: charlie)

**Result**: All three can access RStudio, see shared data, run analyses, and share results.

### Scenario 2: University Course with 20 Students

**Setup**:
```bash
# Professor launches Jupyter environment
prism launch python-ml course-ml-101

# Connect
prism connect course-ml-101

# Create student accounts (script)
for i in {1..20}; do
    username="student$i"
    password=$(openssl rand -base64 12)
    sudo useradd -m -s /bin/bash "$username"
    echo "$username:$password" | sudo chpasswd
    echo "$username,$password" >> ~/student-credentials.csv
done

# Create course directory with read-only materials
sudo mkdir -p /shared/course/{lectures,assignments,student-work}
sudo cp -r ~/course-materials/* /shared/course/lectures/
sudo chmod -R 755 /shared/course/lectures  # Read-only

# Student work area
sudo chmod 1777 /shared/course/student-work  # Sticky bit
```

**Students access**:
- Each student gets unique username/password
- Access via `http://<ip>:8888`
- Can read lecture materials, write own work
- Can't see other students' work (sticky bit)

### Scenario 3: Conference Workshop with 40 Attendees

**Setup**:
```bash
# Workshop organizer launches instance
prism launch python-ml workshop-pandas-2024

# Create single shared account for attendees
sudo adduser workshop
echo "workshop:WorkshopPass2024" | sudo chpasswd
sudo usermod -aG jupyter-users workshop

# Pre-load workshop materials
sudo mkdir -p /home/workshop/workshop-materials
sudo cp -r ~/notebooks /home/workshop/workshop-materials/
sudo chown -R workshop:workshop /home/workshop/workshop-materials
```

**Attendees access**:
- All use same credentials: workshop / WorkshopPass2024
- Access via `http://<ip>:8888`
- Can follow along with instructor
- Can save notebooks to their own directories

**Note**: Shared account works for workshops because:
- Short duration (hours, not days)
- Everyone works on same materials
- No need to track individual progress
- Easy to communicate single password

### Scenario 4: International Collaboration (US + Chile)

**Setup**:
```bash
# US researcher launches R environment
prism launch "R Research Environment (Simplified)" intl-collab

# Connect and add Chilean collaborator
prism connect intl-collab
sudo adduser maria

# Create shared workspace
sudo mkdir -p /shared/cancer-genomics/{raw-data,processed,analysis,papers}
sudo groupadd genomics
sudo usermod -aG genomics $(whoami)
sudo usermod -aG genomics maria
sudo chgrp -R genomics /shared/cancer-genomics
sudo chmod -R 2775 /shared/cancer-genomics

# Grant RStudio access
sudo usermod -aG rstudio-users maria
```

**Workflow**:
- US researcher (8am-5pm PST): Morning data processing
- Chilean researcher (4pm-1am PST / 8am-5pm Chile): Evening analysis
- Both: Collaborate async via shared RStudio notebooks
- Weekly video calls to discuss findings

**Benefits**:
- No data transfer delays (work on same instance)
- No version conflicts (shared files, not copies)
- 24/7 research cycle (work continues overnight)
- Cost-effective (single instance instead of two)

---

## Next Steps

### For Instance Owners

1. **Try it out**: Add a test user and verify access
2. **Document your setup**: Keep notes on usernames, passwords, directory structure
3. **Regular backups**: Create EBS snapshots of shared workspace
4. **Monitor usage**: Check disk space, CPU, memory usage
5. **Plan for growth**: Consider larger instance if team expands

### For Collaborators

1. **Bookmark the URL**: Save RStudio/Jupyter link
2. **Explore the interface**: Familiarize yourself with the environment
3. **Check shared directories**: See what data is available
4. **Communicate with team**: Let others know what you're working on
5. **Ask questions**: Don't hesitate to ask instance owner for help

### Advanced Topics

Ready to level up? Explore:
- **Git integration**: Version control for analysis scripts
- **Custom environments**: Conda, virtualenv for specific packages
- **Scheduled jobs**: Cron for automated data processing
- **Resource monitoring**: Track CPU/memory usage per user
- **HTTPS setup**: Secure web access with SSL certificates

---

## Related Documentation

- [Custom AMI Workflow](CUSTOM_AMI_WORKFLOW.md) - Save configured multi-user instances as AMIs
- [AMI Best Practices](AMI_BEST_PRACTICES.md) - Versioning and management for team AMIs
- [AWS Setup Guide](AWS_SETUP_GUIDE.md) - Initial AWS account configuration
- [Collaboration Quickstart](COLLABORATION_QUICKSTART.md) - Simplified 5-minute guide

---

## Support

Need help?
- **GitHub Issues**: [Report issues](https://github.com/scttfrdmn/prism/issues)
- **Documentation**: [Full docs site](https://scttfrdmn.github.io/prism/)
- **Community**: [GitHub Discussions](https://github.com/scttfrdmn/prism/discussions)

---

**Last Updated**: January 2026 | **Version**: 1.0
