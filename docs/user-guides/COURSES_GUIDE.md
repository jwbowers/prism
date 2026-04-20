# Courses Guide

Prism Courses enable instructors to provision research workspaces for entire university classes — with per-student budget limits, template enforcement, and bulk enrollment.

---

## Overview

A **Course** in Prism represents a class or semester-long research computing allocation. It provides:

- **Template whitelist**: restrict students to approved environments
- **Budget management**: per-student spending limits with centralized tracking
- **Bulk enrollment**: enroll students individually, in bulk via CSV, or via invitation tokens
- **Workspace monitoring**: view all student workspaces, usage, and spending from a single dashboard

---

## Creating a Course

### Via GUI

1. Navigate to **Courses** in the sidebar
2. Click **Create Course**
3. Fill in: name, code, description, term, institution
4. Set a total budget (distributed to students later)
5. Click **Create**

### Via CLI

```bash
prism course create \
  --name "BIOL 5100: Genomics" \
  --code BIOL5100-F26 \
  --description "Graduate genomics computing lab" \
  --term "Fall 2026"
```

---

## Enrolling Students

### Individual enrollment

```bash
prism course members enroll <course-id>
# → Interactive: enter email, user ID, display name, role, budget limit
```

### Bulk enrollment via CSV

Prepare a CSV with columns: `email,user_id,display_name,role,budget_limit`

```bash
prism course members import <course-id> students.csv
```

### Via invitation tokens

Create a shared enrollment token and distribute to the class:

1. GUI: Course → Members → **Create Enrollment Token**
2. Students redeem it: `prism invitation redeem <token>`

---

## Managing Templates

Enforce which templates students can launch (template whitelist):

```bash
# Add allowed templates
prism course templates add <course-id> python-ml
prism course templates add <course-id> r-research

# List allowed templates
prism course templates list <course-id>

# Remove a template
prism course templates remove <course-id> genomics
```

When a whitelist is set, students can only launch workspaces using approved templates. When no whitelist is set, all templates are available.

---

## Budget Management

### Distributing budget to students

```bash
# Set per-student budget
prism course budget distribute <course-id> --per-student 50.00

# Or set individual limits
prism course members enroll <course-id> --budget 75.00
```

### Monitoring spending

In the GUI, the **Budget** tab on a course shows:
- Total budget vs. spent
- Per-student breakdown with status (ok / warning / critical)
- Running cost projections

Students receive alerts when approaching their limit. Workspaces are stopped if a student exceeds their budget.

---

## Monitoring Student Workspaces

### Overview tab

The **Overview** tab shows all enrolled students with:
- Current spend vs. limit
- Active workspace count
- Budget status (green / yellow / red)

### Reset a student's workspace

```bash
prism course members provision <course-id> <user-id>
```

This terminates the student's current workspace and provisions a fresh one using the course's default template.

---

## Course Settings

| Setting | Description |
|---------|-------------|
| Template whitelist | Restrict launches to approved templates |
| Per-student budget | Maximum spend per student |
| Default template | Template used on workspace provision/reset |
| Enrollment period | Start and end dates for enrollment |

---

## CLI Reference

```bash
prism course                                    # List your courses
prism course create                             # Create a course (interactive)
prism course members list <course-id>           # List enrolled students
prism course members enroll <course-id>         # Enroll a student
prism course members unenroll <course-id> <uid> # Remove a student
prism course members import <course-id> <file>  # Bulk CSV import
prism course templates list <course-id>         # View template whitelist
prism course templates add <course-id> <slug>   # Add to whitelist
prism course templates remove <course-id> <slug>
prism course budget distribute <course-id>      # Set per-student budgets
prism course members provision <course-id> <uid>  # Reset student workspace
```
