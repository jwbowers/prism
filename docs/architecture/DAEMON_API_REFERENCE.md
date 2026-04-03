# Prism Daemon API Reference

## Version: v0.30.0
**Last Updated**: 2026-04-02
**Port**: 8947
**Base URL**: `http://localhost:8947`
**Protocol**: REST API with JSON

> **Note**: This API reference is maintained alongside the codebase. The version number reflects the current Prism release. All endpoints are stable unless marked as "Future Enhancement" or "Planned".

---

## 🔌 **API Architecture**

The Prism daemon provides a unified REST API that serves all client interfaces:
- **CLI Client** (`cmd/prism`) - Command-line interface
- **GUI Client** (`cmd/prism-gui`) - Desktop application (Wails 3.x)

All clients use the same API endpoints through the `pkg/api/client` library for consistent functionality across interfaces.

---

## 🔧 **Core API Endpoints**

### **Templates**

#### `GET /api/v1/templates`
Retrieve all available Prism templates.

**Response**:
```json
[
  {
    "name": "Python Machine Learning (Simplified)",
    "description": "Conda + Jupyter + ML packages (scikit-learn, pandas, matplotlib)",
    "package_manager": "conda",
    "packages": ["jupyter", "scikit-learn", "pandas", "matplotlib", "numpy"],
    "users": ["datascientist"],
    "services": ["jupyter"],
    "ports": [22, 8888],
    "validation_status": "validated"
  },
  {
    "name": "R Research Environment (Simplified)", 
    "description": "Conda + RStudio + tidyverse packages for statistical analysis",
    "package_manager": "conda",
    "packages": ["r-base", "rstudio", "r-tidyverse", "r-ggplot2"],
    "users": ["researcher"],
    "services": ["rstudio"],
    "ports": [22, 8787],
    "validation_status": "validated"
  }
]
```

#### `GET /api/v1/templates/validate`
Validate all templates for syntax, dependencies, and AWS deployment readiness.

**Response**:
```json
{
  "total_templates": 6,
  "validated": 6,
  "errors": [],
  "validation_results": [
    {
      "template": "Python Machine Learning (Simplified)",
      "status": "valid",
      "checks_passed": 8,
      "issues": []
    }
  ]
}
```

#### `GET /api/v1/templates/{name}/info`
Get detailed information about a specific template including dependency chains and inheritance.

**Response**:
```json
{
  "name": "Rocky Linux 9 + Conda Stack",
  "inherits": ["Rocky Linux 9 Base"],
  "resolved_packages": ["dnf", "system-tools", "jupyter", "scikit-learn", "pandas"],
  "resolved_users": ["rocky", "datascientist"],
  "resolved_ports": [22, 8888],
  "dependency_chain": [
    "Rocky Linux 9 Base → Rocky Linux 9 + Conda Stack"
  ],
  "validation_status": "validated",
  "troubleshooting_guide": "For conda environment issues, ensure PATH includes /home/datascientist/miniconda3/bin"
}
```

---

### **Workspaces**

#### `GET /api/v1/instances`
List all Prism workspaces with current status and metadata.

**Response**:
```json
[
  {
    "name": "my-ml-research",
    "id": "i-1234567890abcdef0",
    "state": "running",
    "public_ip": "54.123.45.67",
    "private_ip": "172.31.1.123", 
    "instance_type": "t3.medium",
    "region": "us-west-2",
    "template": "Python Machine Learning (Simplified)",
    "launch_time": "2024-06-15T10:30:00Z",
    "hourly_rate": 0.0416,
    "current_spend": 2.45,
    "hibernation_capable": true,
    "attached_volumes": ["shared-research-data"],
    "attached_ebs_volumes": ["project-storage-L"],
    "tags": {
      "Project": "cancer-research",
      "Budget": "grant-nih-2024"
    }
  }
]
```

#### `POST /api/v1/instances/launch`
Launch a new Prism workspace.

**Request**:
```json
{
  "name": "gpu-training-workstation",
  "template": "Python Machine Learning (Simplified)",
  "size": "XL",
  "instance_type": "g4dn.xlarge",
  "spot": true,
  "region": "us-west-2",
  "storage": {
    "ebs_volume_size": 100,
    "ebs_volume_type": "gp3"
  },
  "project": "cancer-research",
  "budget": "grant-nih-2024"
}
```

**Response**:
```json
{
  "name": "gpu-training-workstation",
  "instance_id": "i-0987654321fedcba0",
  "state": "launching",
  "estimated_ready_time": "2024-06-15T10:35:00Z",
  "hourly_rate": 0.526,
  "launch_progress": 15
}
```

#### `POST /api/v1/instances/{name}/stop`
Stop a running workspace (preserves EBS storage).

**Response**:
```json
{
  "name": "my-ml-research",
  "previous_state": "running", 
  "new_state": "stopping",
  "message": "Workspace stopping - all data preserved"
}
```

#### `POST /api/v1/instances/{name}/start`  
Start a stopped workspace.

**Response**:
```json
{
  "name": "my-ml-research",
  "previous_state": "stopped",
  "new_state": "starting", 
  "estimated_ready_time": "2024-06-15T11:05:00Z"
}
```

#### `POST /api/v1/instances/{name}/terminate`
Permanently terminate a workspace (destroys all data).

**Request**:
```json
{
  "confirm": "DELETE",
  "backup_ebs": false
}
```

**Response**:
```json
{
  "name": "my-ml-research",
  "state": "terminating",
  "message": "Workspace terminating - all data will be permanently lost"
}
```

#### `GET /api/v1/instances/{name}/connect`
Get connection information for accessing a workspace.

**Response**:
```json
{
  "ssh": {
    "command": "ssh -i ~/.ssh/prism.pem ec2-user@54.123.45.67",
    "host": "54.123.45.67", 
    "user": "ec2-user",
    "port": 22
  },
  "services": {
    "jupyter": {
      "url": "http://54.123.45.67:8888",
      "token": "a1b2c3d4e5f6g7h8i9j0",
      "local_forward": "ssh -L 8888:localhost:8888 -i ~/.ssh/prism.pem ec2-user@54.123.45.67"
    },
    "rstudio": {
      "url": "http://54.123.45.67:8787", 
      "username": "researcher",
      "password": "generated_password_123"
    }
  }
}
```

---

### **Hibernation System** (Phase 3)

#### `POST /api/v1/instances/{name}/hibernate`
Hibernate a workspace (preserves RAM state + EBS storage).

**Response**:
```json
{
  "name": "my-ml-research",
  "previous_state": "running",
  "new_state": "hibernating",
  "hibernation_supported": true,
  "estimated_savings": "$0.0416/hour during hibernation"
}
```

#### `POST /api/v1/instances/{name}/resume`  
Resume a hibernated workspace (restores RAM state).

**Response**:
```json
{
  "name": "my-ml-research",
  "previous_state": "hibernated", 
  "new_state": "resuming",
  "estimated_ready_time": "2024-06-15T11:02:00Z",
  "message": "Resuming from hibernation - RAM state preserved"
}
```

#### `GET /api/v1/instances/{name}/hibernation-status`
Check hibernation capability and status for a workspace.

**Response**:
```json
{
  "hibernation_supported": true,
  "current_state": "hibernated",
  "hibernation_time": "2024-06-15T15:30:00Z",
  "estimated_savings": 4.2,
  "hibernation_duration": "4h 30m"
}
```

---

### **Idle Detection & Automated Hibernation**

#### `GET /api/v1/idle/profiles`
List available idle detection profiles for automated hibernation.

**Response**:
```json
[
  {
    "name": "batch",
    "description": "Long-running research jobs", 
    "idle_threshold_minutes": 60,
    "action": "hibernate",
    "cpu_threshold": 5.0,
    "memory_threshold": 10.0,
    "network_threshold": 1024,
    "gpu_threshold": 5.0
  },
  {
    "name": "gpu",
    "description": "GPU workstations", 
    "idle_threshold_minutes": 15,
    "action": "stop",
    "cpu_threshold": 10.0,
    "gpu_threshold": 10.0
  }
]
```

#### `POST /api/v1/idle/profiles`
Create a new idle detection profile.

**Request**:
```json
{
  "name": "cost-optimized",
  "description": "Maximum cost savings",
  "idle_threshold_minutes": 10,
  "action": "hibernate",
  "cpu_threshold": 2.0,
  "memory_threshold": 5.0,
  "network_threshold": 512,
  "disk_threshold": 1024,
  "gpu_threshold": 2.0
}
```

#### `POST /api/v1/idle/instances/{name}/configure`
Configure idle detection for a specific workspace.

**Request**:
```json
{
  "profile": "gpu",
  "enabled": true,
  "custom_threshold_minutes": 20
}
```

#### `GET /api/v1/idle/history`
Get history of automated hibernation actions.

**Response**:
```json
[
  {
    "instance_name": "my-gpu-workstation",
    "action": "hibernate",
    "trigger_time": "2024-06-15T14:20:00Z",
    "idle_duration": "25 minutes", 
    "profile": "gpu",
    "estimated_savings": 0.52
  }
]
```

---

### **Enterprise Project Management** (Phase 4)

#### `GET /api/v1/projects`
List all research projects with budget and member information.

**Response**:
```json
[
  {
    "id": "proj_cancer_research_2024",
    "name": "Cancer Research Initiative",
    "description": "Multi-year cancer genomics research project",
    "budget": {
      "total_allocated": 50000.0,
      "current_spend": 12450.30,
      "remaining": 37549.70,
      "monthly_burn_rate": 4150.10
    },
    "members": [
      {
        "email": "dr.smith@university.edu",
        "role": "owner",
        "join_date": "2024-01-15T00:00:00Z"
      },
      {
        "email": "researcher.jones@university.edu", 
        "role": "admin",
        "join_date": "2024-02-01T00:00:00Z"
      }
    ],
    "active_instances": 3,
    "total_instances": 12,
    "created_at": "2024-01-15T00:00:00Z"
  }
]
```

#### `POST /api/v1/projects`
Create a new research project.

**Request**:
```json
{
  "name": "Climate Modeling Project", 
  "description": "Large-scale climate simulation research",
  "budget": {
    "total_allocated": 75000.0,
    "alert_threshold": 80.0,
    "auto_hibernate_threshold": 90.0,
    "prevent_launch_threshold": 95.0
  },
  "members": [
    {
      "email": "climate.lead@university.edu",
      "role": "owner"
    }
  ]
}
```

#### `GET /api/v1/projects/{project_id}/budget`
Get detailed budget information and cost breakdown for a project.

**Response**:
```json
{
  "total_allocated": 50000.0,
  "current_spend": 12450.30,
  "cost_breakdown": {
    "compute": 8920.15,
    "storage": 2130.45,
    "data_transfer": 1399.70
  },
  "hibernation_savings": 3240.80,
  "top_spending_instances": [
    {
      "name": "gpu-training-cluster",
      "cost": 4520.30,
      "percentage": 36.3
    }
  ],
  "budget_alerts": {
    "threshold_80_percent": false,
    "threshold_90_percent": false,
    "overspend_risk": "low"
  }
}
```

#### `GET /api/v1/projects/{project_id}/cost-analysis`
Get comprehensive cost analysis and optimization recommendations.

**Response**:
```json
{
  "current_monthly_burn": 4150.10,
  "projected_monthly_burn": 4850.25,
  "hibernation_potential_savings": 1240.50,
  "optimization_recommendations": [
    {
      "type": "hibernation",
      "instance": "data-processing-large",
      "potential_savings": 520.30,
      "recommendation": "Enable automated hibernation after 30 minutes idle"
    },
    {
      "type": "rightsizing",
      "instance": "web-scraper-micro",
      "potential_savings": 180.40,
      "recommendation": "Downsize from t3.medium to t3.small - current CPU usage <10%"
    }
  ]
}
```

---

### **Budget Management** (v0.5.10+)

Budget pools allow institutions to allocate research funding across multiple projects with automated tracking, alerts, and reallocation capabilities.

#### `GET /api/v1/budgets`
List all budget pools with allocation and spending summaries.

**Response**:
```json
{
  "budgets": [
    {
      "id": "budget_nsf_grant_2024",
      "name": "NSF Grant CISE-2024-12345",
      "description": "3-year NSF grant for ML research",
      "total_amount": 150000.0,
      "allocated_amount": 120000.0,
      "spent_amount": 45320.50,
      "remaining_amount": 104679.50,
      "period": "project",
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2026-12-31T23:59:59Z",
      "alert_threshold": 0.80,
      "created_by": "admin@university.edu",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-12-10T15:30:00Z"
    }
  ],
  "count": 1
}
```

#### `POST /api/v1/budgets`
Create a new budget pool.

**Request**:
```json
{
  "name": "NIH Grant R01-2025-67890",
  "description": "5-year NIH R01 for cancer genomics research",
  "total_amount": 250000.0,
  "period": "project",
  "start_date": "2025-01-01T00:00:00Z",
  "end_date": "2029-12-31T23:59:59Z",
  "alert_threshold": 0.85,
  "created_by": "pi@medical-school.edu"
}
```

**Response**:
```json
{
  "id": "budget_nih_r01_2025",
  "name": "NIH Grant R01-2025-67890",
  "description": "5-year NIH R01 for cancer genomics research",
  "total_amount": 250000.0,
  "allocated_amount": 0.0,
  "spent_amount": 0.0,
  "remaining_amount": 250000.0,
  "period": "project",
  "start_date": "2025-01-01T00:00:00Z",
  "end_date": "2029-12-31T23:59:59Z",
  "alert_threshold": 0.85,
  "created_by": "pi@medical-school.edu",
  "created_at": "2025-01-01T10:00:00Z",
  "updated_at": "2025-01-01T10:00:00Z"
}
```

#### `GET /api/v1/budgets/{budget_id}`
Get detailed information about a specific budget pool.

**Response**:
```json
{
  "id": "budget_nsf_grant_2024",
  "name": "NSF Grant CISE-2024-12345",
  "description": "3-year NSF grant for ML research",
  "total_amount": 150000.0,
  "allocated_amount": 120000.0,
  "spent_amount": 45320.50,
  "remaining_amount": 104679.50,
  "utilization_rate": 0.3021,
  "period": "project",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2026-12-31T23:59:59Z",
  "alert_threshold": 0.80,
  "created_by": "admin@university.edu",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-12-10T15:30:00Z"
}
```

#### `PUT /api/v1/budgets/{budget_id}`
Update budget pool details (name, description, total amount, alert threshold).

**Request**:
```json
{
  "total_amount": 175000.0,
  "alert_threshold": 0.90,
  "description": "3-year NSF grant (supplemental funding added)"
}
```

**Response**: Updated budget object (same format as GET).

#### `DELETE /api/v1/budgets/{budget_id}`
Delete a budget pool. All project allocations must be removed first.

**Response**:
```json
{
  "message": "Budget deleted successfully"
}
```

#### `GET /api/v1/budgets/{budget_id}/summary`
Get comprehensive budget summary with allocations and spending breakdown.

**Response**:
```json
{
  "budget": {
    "id": "budget_nsf_grant_2024",
    "name": "NSF Grant CISE-2024-12345",
    "total_amount": 150000.0,
    "allocated_amount": 120000.0,
    "spent_amount": 45320.50,
    "remaining_amount": 104679.50,
    "utilization_rate": 0.3021
  },
  "allocations": [
    {
      "id": "alloc_cancer_research",
      "project_id": "proj_cancer_research",
      "project_name": "Cancer Genomics Study",
      "allocated_amount": 75000.0,
      "spent_amount": 32150.25,
      "remaining": 42849.75,
      "utilization_rate": 0.4287
    },
    {
      "id": "alloc_ml_platform",
      "project_id": "proj_ml_platform",
      "project_name": "ML Infrastructure",
      "allocated_amount": 45000.0,
      "spent_amount": 13170.25,
      "remaining": 31829.75,
      "utilization_rate": 0.2927
    }
  ],
  "status": "ok",
  "alerts": [],
  "top_spending_projects": [
    {
      "project_id": "proj_cancer_research",
      "project_name": "Cancer Genomics Study",
      "spent": 32150.25,
      "percentage": 70.9
    }
  ]
}
```

#### `GET /api/v1/budgets/{budget_id}/allocations`
List all project allocations for a budget pool.

**Response**:
```json
{
  "allocations": [
    {
      "id": "alloc_cancer_research",
      "budget_id": "budget_nsf_grant_2024",
      "project_id": "proj_cancer_research",
      "allocated_amount": 75000.0,
      "spent_amount": 32150.25,
      "alert_threshold": 0.80,
      "allocated_at": "2024-01-15T00:00:00Z",
      "allocated_by": "admin@university.edu"
    }
  ],
  "count": 2
}
```

---

### **Project Invitations** (v0.5.11+)

Invitation system allows project owners to invite collaborators via email with role-based access control.

#### `POST /api/v1/invitations/send`
Send an invitation to join a project.

**Request**:
```json
{
  "project_id": "proj_cancer_research",
  "email": "researcher@partner-university.edu",
  "role": "admin",
  "message": "We'd like you to join our cancer genomics research project. You'll have admin access to manage instances and storage.",
  "invited_by": "pi@university.edu"
}
```

**Response**:
```json
{
  "id": "inv_abc123def456",
  "project_id": "proj_cancer_research",
  "email": "researcher@partner-university.edu",
  "role": "admin",
  "token": "secure-invitation-token-xyz789",
  "invited_by": "pi@university.edu",
  "invited_at": "2025-01-15T10:30:00Z",
  "expires_at": "2025-01-22T10:30:00Z",
  "status": "pending",
  "message": "We'd like you to join our cancer genomics research project..."
}
```

#### `GET /api/v1/invitations/project/{project_id}`
List all invitations for a specific project (all statuses: pending, accepted, declined, expired).

**Query Parameters**:
- `status` (optional): Filter by status (pending, accepted, declined, expired, revoked)

**Response**:
```json
{
  "invitations": [
    {
      "id": "inv_abc123def456",
      "email": "researcher@partner-university.edu",
      "role": "admin",
      "status": "pending",
      "invited_by": "pi@university.edu",
      "invited_at": "2025-01-15T10:30:00Z",
      "expires_at": "2025-01-22T10:30:00Z"
    },
    {
      "id": "inv_def789ghi012",
      "email": "student@university.edu",
      "role": "viewer",
      "status": "accepted",
      "invited_by": "pi@university.edu",
      "invited_at": "2025-01-10T14:00:00Z",
      "accepted_at": "2025-01-10T16:30:00Z"
    }
  ],
  "count": 2
}
```

#### `GET /api/v1/invitations/me`
List all invitations for the current user (invitations where you are the invitee).

**Query Parameters**:
- `status` (optional): Filter by status (pending, accepted, declined)

**Response**:
```json
{
  "invitations": [
    {
      "id": "inv_abc123def456",
      "project_id": "proj_cancer_research",
      "project_name": "Cancer Genomics Study",
      "role": "admin",
      "status": "pending",
      "invited_by": "pi@university.edu",
      "invited_at": "2025-01-15T10:30:00Z",
      "expires_at": "2025-01-22T10:30:00Z",
      "message": "We'd like you to join..."
    }
  ],
  "count": 1
}
```

#### `GET /api/v1/invitations/{invitation_id}`
Get detailed information about a specific invitation.

**Response**:
```json
{
  "id": "inv_abc123def456",
  "project_id": "proj_cancer_research",
  "project_name": "Cancer Genomics Study",
  "email": "researcher@partner-university.edu",
  "role": "admin",
  "token": "secure-invitation-token-xyz789",
  "invited_by": "pi@university.edu",
  "invited_at": "2025-01-15T10:30:00Z",
  "expires_at": "2025-01-22T10:30:00Z",
  "status": "pending",
  "resend_count": 0,
  "message": "We'd like you to join our cancer genomics research project..."
}
```

#### `POST /api/v1/invitations/{invitation_id}/accept`
Accept a project invitation.

**Request**:
```json
{
  "token": "secure-invitation-token-xyz789"
}
```

**Response**:
```json
{
  "id": "inv_abc123def456",
  "project_id": "proj_cancer_research",
  "status": "accepted",
  "accepted_at": "2025-01-15T14:00:00Z",
  "message": "Successfully joined project as admin"
}
```

#### `POST /api/v1/invitations/{invitation_id}/decline`
Decline a project invitation.

**Request**:
```json
{
  "token": "secure-invitation-token-xyz789",
  "reason": "Currently unavailable due to other commitments"
}
```

**Response**:
```json
{
  "id": "inv_abc123def456",
  "status": "declined",
  "declined_at": "2025-01-15T14:00:00Z",
  "decline_reason": "Currently unavailable due to other commitments"
}
```

#### `POST /api/v1/invitations/{invitation_id}/resend`
Resend an invitation email (for pending invitations only).

**Response**:
```json
{
  "id": "inv_abc123def456",
  "status": "pending",
  "resend_count": 1,
  "last_resent": "2025-01-16T10:00:00Z",
  "message": "Invitation resent successfully"
}
```

#### `DELETE /api/v1/invitations/{invitation_id}`
Revoke a pending invitation (project owner/admin only).

**Response**:
```json
{
  "message": "Invitation revoked successfully"
}
```

#### `POST /api/v1/invitations/bulk`
Send invitations to multiple users at once.

**Request**:
```json
{
  "project_id": "proj_ml_course",
  "invitations": [
    {
      "email": "student1@university.edu",
      "role": "viewer"
    },
    {
      "email": "student2@university.edu",
      "role": "viewer"
    },
    {
      "email": "ta@university.edu",
      "role": "admin",
      "message": "TA access for ML course"
    }
  ],
  "default_role": "viewer",
  "default_message": "Welcome to the ML course project!",
  "invited_by": "professor@university.edu"
}
```

**Response**:
```json
{
  "summary": {
    "total": 3,
    "sent": 3,
    "failed": 0,
    "skipped": 0
  },
  "results": [
    {
      "email": "student1@university.edu",
      "status": "sent",
      "invitation_id": "inv_student1_abc"
    },
    {
      "email": "student2@university.edu",
      "status": "sent",
      "invitation_id": "inv_student2_def"
    },
    {
      "email": "ta@university.edu",
      "status": "sent",
      "invitation_id": "inv_ta_ghi"
    }
  ]
}
```

---

### **Storage Management**

#### `GET /api/v1/storage/volumes`
List all EFS and EBS storage volumes.

**Response**:
```json
{
  "efs_volumes": [
    {
      "name": "shared-research-data",
      "filesystem_id": "fs-1234567890abcdef0", 
      "state": "available",
      "size_gb": 250.5,
      "creation_time": "2024-06-01T10:00:00Z",
      "monthly_cost": 76.65,
      "attached_instances": ["ml-workstation-1", "data-processor-2"]
    }
  ],
  "ebs_volumes": [
    {
      "name": "project-storage-L", 
      "volume_id": "vol-0987654321fedcba0",
      "state": "in-use",
      "size_gb": 100,
      "volume_type": "gp3",
      "attached_instance": "my-ml-research",
      "monthly_cost": 8.0,
      "iops": 3000,
      "throughput": 125
    }
  ]
}
```

#### `POST /api/v1/storage/volumes/create`
Create new EFS or EBS storage volume.

**Request**:
```json
{
  "type": "ebs",
  "name": "large-dataset-storage",
  "size_gb": 500,
  "volume_type": "gp3",
  "iops": 4000,
  "throughput": 250,
  "project": "cancer-research"
}
```

---

### **Cost Tracking & Analytics**

#### `GET /api/v1/costs/current`
Get current AWS costs across all Prism resources.

**Response**:
```json
{
  "total_current_hourly": 2.45,
  "total_daily_projection": 58.80,
  "total_monthly_projection": 1764.00,
  "cost_breakdown": {
    "instances": 1.89,
    "ebs_storage": 0.32,
    "efs_storage": 0.24
  },
  "hibernation_savings": {
    "current_month": 320.50,
    "projected_monthly": 485.25
  },
  "top_cost_drivers": [
    {
      "resource": "gpu-training-workstation", 
      "hourly_cost": 0.526,
      "percentage": 21.5
    }
  ]
}
```

#### `GET /api/v1/costs/history`
Get historical cost data with trends and analysis.

**Request Parameters**:
- `start_date`: ISO 8601 date (default: 30 days ago)
- `end_date`: ISO 8601 date (default: now)
- `granularity`: `daily`, `weekly`, `monthly` (default: daily)

**Response**:
```json
{
  "period": "30_days",
  "total_cost": 1247.50,
  "daily_costs": [
    {
      "date": "2024-06-15",
      "cost": 45.30,
      "instances": 8,
      "hibernation_savings": 12.40
    }
  ],
  "cost_trends": {
    "trend_direction": "increasing",
    "percentage_change": 15.3,
    "primary_driver": "additional GPU workspaces"
  }
}
```

---

### **Profile Management**

#### `GET /api/v1/profiles/current`
Get current active AWS profile and region configuration.

**Response**:
```json
{
  "name": "research-profile",
  "aws_profile": "my-research-account",
  "region": "us-west-2",
  "default_project": "cancer-research",
  "created_at": "2024-01-15T00:00:00Z",
  "last_used": "2024-06-15T10:30:00Z"
}
```

#### `GET /api/v1/profiles`
List all available Prism profiles.

**Response**:
```json
[
  {
    "name": "research-profile",
    "aws_profile": "my-research-account", 
    "region": "us-west-2",
    "default_project": "cancer-research",
    "active": true
  },
  {
    "name": "teaching-profile",
    "aws_profile": "university-teaching",
    "region": "us-east-1", 
    "default_project": "cs101-labs",
    "active": false
  }
]
```

#### `POST /api/v1/profiles/switch`
Switch to a different Prism profile.

**Request**:
```json
{
  "profile_name": "teaching-profile"
}
```

---

### **System Management**

#### `GET /api/v1/daemon/status`
Get daemon health status and version information.

**Response**:
```json
{
  "status": "healthy",
  "version": "0.5.5",
  "uptime": "2h 45m 30s",
  "api_version": "v1",
  "aws_connectivity": "connected",
  "active_profiles": 2,
  "total_instances": 12,
  "active_instances": 8
}
```

#### `POST /api/v1/daemon/shutdown`
Gracefully shutdown the daemon service.

**Response**:
```json
{
  "message": "Daemon shutting down gracefully",
  "active_operations": 0,
  "shutdown_timeout": 30
}
```

---

## 🔒 **Authentication & Security**

### **Local API Security**
- **Port Binding**: Daemon binds to `localhost:8947` only (no external access)
- **No Authentication**: Local-only daemon requires no authentication
- **AWS Credentials**: Uses standard AWS profile-based authentication
- **Process Isolation**: Each client creates independent daemon connection

### **AWS Integration**
```json
{
  "aws_credential_sources": [
    "AWS_PROFILE environment variable",
    "~/.aws/credentials file", 
    "~/.aws/config file",
    "IAM instance profile (when running on EC2)",
    "AWS SSO profiles"
  ],
  "required_permissions": [
    "ec2:*",
    "efs:*", 
    "iam:PassRole",
    "ssm:GetParameter"
  ]
}
```

---

## 📡 **Real-Time Updates**

### **Polling Strategy** (Current)
- **Instance Status**: 30-second polling interval
- **Cost Updates**: 5-minute polling interval  
- **Template Validation**: On-demand only

### **WebSocket Support** (Future Enhancement)
```json
{
  "endpoint": "ws://localhost:8947/api/v1/ws",
  "events": [
    "instance_state_changed",
    "hibernation_triggered", 
    "budget_threshold_exceeded",
    "template_validation_completed"
  ]
}
```

---

## 🚨 **Error Handling**

### **Standard Error Response Format**
```json
{
  "error": {
    "code": "INSTANCE_NOT_FOUND",
    "message": "Workspace 'my-research' not found in current region",
    "details": "Check workspace name and ensure correct AWS profile/region",
    "remediation": "Use 'prism workspace list' to see available workspaces",
    "timestamp": "2024-06-15T10:30:00Z"
  }
}
```

### **Common Error Codes**
- `INSTANCE_NOT_FOUND` - Workspace doesn't exist
- `TEMPLATE_INVALID` - Template validation failed  
- `AWS_PERMISSION_DENIED` - Insufficient AWS permissions
- `BUDGET_EXCEEDED` - Project budget limits exceeded
- `HIBERNATION_NOT_SUPPORTED` - Workspace type doesn't support hibernation
- `DAEMON_UNREACHABLE` - Cannot connect to daemon service

---

## 🔧 **Client Integration Examples**

### **Go API Client Usage**
```go
// pkg/api/client integration example
client := api.NewClientWithOptions("http://localhost:8947", client.Options{
    AWSProfile: "research-profile", 
    AWSRegion:  "us-west-2",
})

// List workspaces
instances, err := client.GetInstances()
if err != nil {
    log.Fatal(err)
}

// Launch new workspace
launchReq := &api.LaunchInstanceRequest{
    Name:     "new-ml-workstation",
    Template: "Python Machine Learning (Simplified)",
    Size:     "L",
    Project:  "cancer-research",
}
instance, err := client.LaunchInstance(launchReq)
```

### **JavaScript Frontend Integration** (Wails 3.x GUI)
```javascript
// Frontend service integration example
async function loadInstances() {
    try {
        const instances = await window.wails.PrismService.GetInstances();
        renderWorkspaces(workspaces);
    } catch (error) {
        console.error('Failed to load workspaces:', error);
        showError(`Failed to load workspaces: ${error.message}`);
    }
}

async function launchInstance(templateName, instanceName, size) {
    const request = {
        Template: templateName,
        Name: instanceName, 
        Size: size
    };
    
    return await window.wails.PrismService.LaunchInstance(request);
}
```

---

**Total API Endpoints**: 35+ endpoints across templates, workspaces, hibernation, projects, storage, costs, profiles, and system management.

This comprehensive API reference provides complete documentation for integrating with the Prism daemon across all client interfaces (CLI, TUI, GUI) and supports the full enterprise research platform feature set.