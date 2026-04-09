// Type definitions
export interface Project {
  id: string;
  name: string;
  description?: string;
  budget_limit?: number;
  current_spend?: number;
  owner_id?: string;
  owner_email?: string;
  owner?: string;
  created_at: string;
  updated_at?: string;
  status: string;
  member_count?: number;
  active_instances?: number;
  total_cost?: number;
  budget_status?: {
    total_budget: number;
    spent_amount: number;
    spent_percentage: number;
    alert_count: number;
  };
  last_activity?: string;
}

export interface User {
  username: string;
  uid: number;
  display_name?: string;  // Frontend sends this
  full_name?: string;     // Backend returns this
  email: string;
  ssh_keys: number;
  created_at: string;
  provisioned_instances?: string[];
  status?: string;
  enabled?: boolean;       // User account status (enabled/disabled)
}

export interface Template {
  Name: string;  // API returns capital N
  Slug: string;  // API returns capital S
  Description?: string;  // API returns capital D
  name?: string;  // Keep lowercase for backward compatibility
  slug?: string;  // Keep lowercase for backward compatibility
  description?: string;  // Keep lowercase for backward compatibility
  category?: string;
  complexity?: string;
  package_manager?: string;
  features?: string[];
  // Additional fields that might come from API
  [key: string]: unknown;
}

export interface Instance {
  id: string;
  name: string;
  template: string;
  state: string;
  public_ip?: string;
  dns_name?: string;
  dns_hostname?: string;
  instance_type?: string;
  launch_time?: string;
  region?: string;
  username?: string;
  project?: string;
  web_services?: WebService[];
}

// Unified StorageVolume interface matching backend API
export interface StorageVolume {
  name: string;
  type: 'workspace' | 'shared' | 'cloud';
  aws_service: 'ebs' | 'efs' | 's3';
  region: string;
  state: string;
  creation_time: string;

  // Size fields (varies by type)
  size_gb?: number;      // EBS
  size_bytes?: number;   // EFS

  // EBS-specific fields
  volume_id?: string;
  volume_type?: string;
  iops?: number;
  throughput?: number;
  attached_to?: string;

  // EFS-specific fields
  filesystem_id?: string;
  mount_targets?: string[];
  performance_mode?: string;
  throughput_mode?: string;

  // Cost
  estimated_cost_gb: number;
}

// Legacy interfaces for backward compatibility
export interface EFSVolume {
  name: string;
  filesystem_id: string;
  region: string;
  creation_time: string;
  state: string;
  performance_mode: string;
  throughput_mode: string;
  estimated_cost_gb: number;
  size_bytes: number;
  attached_to?: string;
}

export interface EBSVolume {
  name: string;
  volume_id: string;
  region: string;
  creation_time: string;
  state: string;
  volume_type: string;
  size_gb: number;
  estimated_cost_gb: number;
  attached_to?: string;
}

export interface InstanceSnapshot {
  snapshot_id: string;
  snapshot_name: string;
  source_instance: string;
  source_instance_id: string;
  source_template: string;
  description: string;
  state: string;
  architecture: string;
  storage_cost_monthly: number;
  created_at: string;
  size_gb?: number;
}

export interface BudgetData {
  project_id: string;
  project_name: string;
  total_budget: number;
  spent_amount: number;
  spent_percentage: number;
  remaining: number;
  alert_count: number;
  status: 'ok' | 'warning' | 'critical';
  projected_monthly_spend?: number;
  days_until_exhausted?: number;
  active_alerts?: Array<{
    threshold: number;
    action: string;
    triggered_at: string;
  }>;
}

export interface CostBreakdown {
  ec2_compute?: number;
  ebs_storage?: number;
  efs_storage?: number;
  data_transfer: number;
  other?: number;
  total: number;
  instances?: number;
  storage?: number;
}

export interface Budget {
  id: string;
  name: string;
  description: string;
  total_amount: number;
  allocated_amount: number;
  spent_amount: number;
  period: string;
  start_date: string;
  end_date?: string;
  alert_threshold: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface BudgetAllocation {
  id: string;
  budget_id: string;
  project_id: string;
  allocated_amount: number;
  spent_amount: number;
  alert_threshold?: number;
  notes?: string;
  allocated_at: string;
  allocated_by: string;
}

export interface BudgetSummary {
  budget: Budget;
  allocations: BudgetAllocation[];
  project_names: Record<string, string>;
  remaining_amount: number;
  utilization_rate: number;
}

export interface CreateBudgetRequest {
  name: string;
  description: string;
  total_amount: number;
  period: string;
  start_date: string;
  end_date?: string;
  alert_threshold: number;
  created_by: string;
}

export interface AMI {
  id: string;
  name: string;
  template_name: string;
  region: string;
  state: string;
  architecture: string;
  size_gb: number;
  description?: string;
  created_at: string;
  tags?: Record<string, string>;
}

export interface AMIBuild {
  id: string;
  template_name: string;
  status: string;
  progress: number;
  current_step?: string;
  error?: string;
  started_at: string;
  completed_at?: string;
}

export interface AMIRegion {
  name: string;
  ami_count: number;
  total_size_gb: number;
  monthly_cost: number;
}

export interface RightsizingRecommendation {
  instance_name: string;
  current_type: string;
  recommended_type: string;
  cpu_utilization: number;
  memory_utilization: number;
  current_cost: number;
  recommended_cost: number;
  monthly_savings: number;
  savings_percentage: number;
  confidence: 'high' | 'medium' | 'low';
  reason?: string;
}

export interface RightsizingStats {
  total_recommendations: number;
  total_monthly_savings: number;
  average_cpu_utilization: number;
  average_memory_utilization: number;
  over_provisioned_count: number;
  optimized_count: number;
}

export interface PolicyStatus {
  enabled: boolean;
  status: string;
  status_icon: string;
  assigned_policies: string[];
  message?: string;
}

export interface PolicySet {
  id: string;
  name: string;
  description: string;
  policies: number;
  status: string;
  tags?: Record<string, string>;
}

export interface PolicyCheckResult {
  allowed: boolean;
  template_name: string;
  reason: string;
  matched_policies?: string[];
  suggestions?: string[];
}

export interface MarketplaceTemplate {
  id: string;
  name: string;
  display_name: string;
  author: string;
  publisher: string;
  category: string;
  description: string;
  rating: number;
  downloads: number;
  verified: boolean;
  featured: boolean;
  version: string;
  tags?: string[];
  badges?: string[];
  created_at: string;
  updated_at: string;
  ami_available?: boolean;
}

export interface MarketplaceCategory {
  id: string;
  name: string;
  count: number;
}

export interface IdlePolicy {
  id: string;
  name: string;
  idle_minutes: number;
  action: 'hibernate' | 'stop' | 'notify';
  cpu_threshold: number;
  memory_threshold: number;
  network_threshold: number;
  description?: string;
  enabled: boolean;
}

export interface IdleSchedule {
  instance_name: string;
  policy_name: string;
  enabled: boolean;
  last_checked: string;
  idle_minutes: number;
  status: string;
}

export interface CachedInvitation {
  token: string;
  invitation_id: string;
  project_id: string;
  project_name: string;
  email: string;
  role: string;
  invited_by: string;
  invited_at: string;
  expires_at: string;
  status: 'pending' | 'accepted' | 'declined' | 'expired' | 'revoked';
  message?: string;
  added_at: string;
}

export interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  header?: string;
  content: string;
  dismissible?: boolean;
  onDismiss?: () => void;
}

export interface ProjectData {
  name: string;
  description?: string;
  budget_limit?: number;
  [key: string]: unknown;
}

export interface MemberData {
  user_id: string;
  username?: string;
  role: string;
  joined_at?: string;
  [key: string]: unknown;
}

export interface UserData {
  username: string;
  display_name: string;
  email: string;
  [key: string]: unknown;
}

export interface SharedTokenConfig {
  name: string;
  role: string;
  redemption_limit?: number;
  expires_at?: string;
  [key: string]: unknown;
}

export interface BulkInviteResponse {
  total: number;
  sent: number;
  failed: number;
  skipped?: number;
  errors?: Array<{email: string; error: string}>;
}

// Additional type definitions for API integration
export interface Invitation {
  id: string;
  project_id: string;
  project_name: string;
  email: string;
  role: 'viewer' | 'member' | 'admin';
  token: string;
  status: 'pending' | 'accepted' | 'declined' | 'expired';
  invited_by: string;
  invited_at: string;
  expires_at: string;
  responded_at?: string;
  message?: string;
}

export interface ProjectDetails extends Project {
  members: ProjectMember[];
  cost_breakdown: CostBreakdown;
}

export interface ProjectMember {
  user_id: string;
  username: string;
  role: string;
  joined_at: string;
}


export interface SharedInvitationToken {
  token: string;
  project_id: string;
  project_name: string;
  name: string;
  role: 'viewer' | 'member' | 'admin';
  redemption_limit: number;
  redemptions: number;
  redemption_count?: number;
  created_at: string;
  created_by: string;
  expires_at: string;
  revoked: boolean;
  qr_code_url?: string;
  status?: string;
}

// v0.13.0 Governance Interfaces
export interface RoleQuota {
  role: string;
  max_instances: number;
  max_instance_type: string;
  max_spend_daily: number;
}
export interface GrantPeriod {
  name: string;
  start_date: string;
  end_date: string;
  auto_freeze: boolean;
  frozen_at?: string;
}
export interface ApprovalRequest {
  id: string;
  project_id: string;
  requested_by: string;
  type: string;
  status: string;
  details: Record<string, unknown>;
  reason: string;
  reviewed_by?: string;
  review_note?: string;
  created_at: string;
  expires_at: string;
  reviewed_at?: string;
}
export interface BudgetShareRequest {
  from_project_id: string;
  to_project_id?: string;
  to_member_id?: string;
  amount: number;
  reason?: string;
  expires_at?: string;
}
export interface BudgetShareRecord {
  id: string;
  request: BudgetShareRequest;
  approved_by: string;
  created_at: string;
  expires_at?: string;
}
export interface OnboardingTemplate {
  id?: string;
  name: string;
  description?: string;
  templates?: string[];
  budget_limit?: number;
}

// v0.14.0/v0.16.0 University Education System Interfaces
export interface ClassMember {
  user_id: string;
  email: string;
  display_name: string;
  role: string;
  budget_spent: number;
  budget_limit: number;
  added_at: string;
  expires_at?: string;
}
export interface Course {
  id: string;
  code: string;
  title: string;
  department: string;
  semester: string;
  semester_start: string;
  semester_end: string;
  owner: string;
  status: string; // "pending"|"active"|"closed"|"archived"
  members: ClassMember[];
  approved_templates: string[];
  per_student_budget: number;
  total_budget: number;
  default_template: string;
  auto_provision_on_enroll: boolean;
  created_at: string;
  updated_at: string;
}
export interface CourseBudgetSummary {
  total_budget: number;
  per_student_default: number;
  total_spent: number;
  students: StudentBudgetInfo[];
}
export interface StudentBudgetInfo {
  user_id: string;
  email: string;
  display_name: string;
  budget_limit: number;
  budget_spent: number;
  remaining: number;
}
export interface CourseOverview {
  course_id: string;
  course_code: string;
  total_students: number;
  active_instances: number;
  total_budget_spent: number;
  students: StudentOverviewStatus[];
}
export interface StudentOverviewStatus {
  user_id: string;
  email: string;
  display_name: string;
  instances: Instance[];
  budget_spent: number;
  budget_limit: number;
  budget_status: string; // "ok"|"warning"|"exceeded"
}
export interface UsageReport {
  course_id: string;
  course_code: string;
  semester: string;
  total_spent: number;
  total_budget: number;
  students: StudentUsageRecord[];
  generated_at: string;
}
export interface StudentUsageRecord {
  user_id: string;
  email: string;
  display_name: string;
  total_hours: number;
  total_cost: number;
  instance_count: number;
}
export interface CourseAuditEntry {
  timestamp: string;
  course_id: string;
  actor: string;
  target: string;
  action: string;
  detail?: Record<string, unknown>;
}

// v0.19.0 Education Power Features interfaces
export interface SharedMaterialsVolume {
  course_id: string;
  efs_id: string;
  mount_path: string;
  state: string;
  size_gb: number;
  created_at: string;
  mounted_instance_count: number;
}
export interface WorkspaceResetResult {
  student_id: string;
  backup_snapshot_id?: string;
  backup_download_url?: string;
  status: string;
  backup_expires_at?: string;
}

// v0.18.0 Workshop & Event Management interfaces
export interface WorkshopParticipant {
  user_id: string;
  email?: string;
  display_name?: string;
  joined_at: string;
  instance_id?: string;
  instance_name?: string;
  status: string; // "pending"|"provisioned"|"running"|"stopped"
  progress?: number; // 0-100
}

export interface WorkshopEvent {
  id: string;
  title: string;
  description?: string;
  owner: string;
  template: string;
  approved_templates?: string[];
  max_participants: number;
  budget_per_participant?: number;
  start_time: string;
  end_time: string;
  early_access_hours?: number;
  status: string; // "draft"|"active"|"ended"|"archived"
  join_token?: string;
  participants?: WorkshopParticipant[];
  created_at: string;
  updated_at: string;
}

export interface WorkshopDashboard {
  workshop_id: string;
  title: string;
  total_participants: number;
  active_instances: number;
  stopped_instances: number;
  pending_instances: number;
  total_spent: number;
  time_remaining: string;
  status: string;
  participants: WorkshopParticipant[];
}

export interface WorkshopConfig {
  name: string;
  template: string;
  max_participants: number;
  budget_per_participant?: number;
  duration_hours: number;
  early_access_hours?: number;
  description?: string;
  created_at: string;
}

export interface AppState {
  activeView: 'dashboard' | 'templates' | 'workspaces' | 'storage' | 'backups' | 'projects' | 'users' | 'ami' | 'rightsizing' | 'policy' | 'marketplace' | 'idle' | 'invitations' | 'logs' | 'settings' | 'terminal' | 'webview' | 'budgets' | 'approvals' | 'courses' | 'workshops' | 'capacity-blocks';
  settingsSection: 'general' | 'profiles' | 'users' | 'ami' | 'rightsizing' | 'policy' | 'marketplace' | 'idle' | 'logs';
  templates: Record<string, Template>;
  instances: Instance[];
  efsVolumes: EFSVolume[];
  ebsVolumes: EBSVolume[];
  snapshots: InstanceSnapshot[];
  projects: Project[];
  users: User[];
  budgets: BudgetData[];
  budgetPools: Budget[];
  selectedBudgetId: string | null;
  amis: AMI[];
  amiBuilds: AMIBuild[];
  amiRegions: AMIRegion[];
  rightsizingRecommendations: RightsizingRecommendation[];
  rightsizingStats: RightsizingStats | null;
  policyStatus: PolicyStatus | null;
  policySets: PolicySet[];
  marketplaceTemplates: MarketplaceTemplate[];
  marketplaceCategories: MarketplaceCategory[];
  idlePolicies: IdlePolicy[];
  idleSchedules: IdleSchedule[];
  invitations: CachedInvitation[];
  courses: Course[];
  workshops: WorkshopEvent[];
  selectedTemplate: Template | null;
  selectedProject: Project | null;
  selectedTerminalInstance: string;
  loading: boolean;
  notifications: Notification[];
  connected: boolean;
  error: string | null;
  updateInfo: any | null;
  autoStartEnabled: boolean;
  pendingApprovalsCount: number;
}

// API Response Interfaces
export interface HibernationStatus {
  supported: boolean;
  enabled?: boolean;
  state?: string;
  message?: string;
}

export interface UserStatus {
  username: string;
  status: string;
  provisioned_instances: string[];
  ssh_keys_count: number;
  last_active?: string;
}

export interface UserProvisionResponse {
  success: boolean;
  instance: string;
  username: string;
  message?: string;
}

export interface SSHKeyResponse {
  public_key: string;
  private_key: string;
  fingerprint: string;
  generated_at: string;
}

export interface SSHKeyConfig {
  key_id: string;
  profile_id: string;
  username: string;
  key_type: string;
  fingerprint: string;
  public_key: string;
  comment: string;
  created_at: string;
  last_used?: string;
  from_profile: string;
  auto_generated: boolean;
}

export interface UserSSHKeysResponse {
  username: string;
  keys: SSHKeyConfig[];
}

export interface ProjectUsageResponse {
  project_id: string;
  period: string;
  instance_hours: number;
  storage_gb_hours: number;
  data_transfer_gb: number;
  [key: string]: string | number;
}

export interface SharedToken {
  token: string;
  project_id: string;
  project_name: string;
  name: string;
  role: string;
  redemption_limit: number;
  redemptions_remaining: number;
  created_by: string;
  created_at: string;
  expires_at?: string;
  status: 'active' | 'expired' | 'exhausted';
}

export interface QRCodeData {
  qr_code: string;  // base64 encoded image or URL
  url: string;
}

export interface RedeemTokenResponse {
  success: boolean;
  project_id: string;
  project_name: string;
  role: string;
  message?: string;
}

export interface WebService {
  name: string;
  port: number;
  url?: string;
  description?: string;
  type?: string;
}

export interface UserUpdateRequest {
  email?: string;
  display_name?: string;
  role?: string;
}

// ── v0.20.0 Storage Power types ──────────────────────────────────────────────

export interface FileEntry {
  path: string;
  size_bytes: number;
  is_dir: boolean;
  modified_at: string;
  permissions: string;
}

export interface CapacityBlock {
  id: string;
  instance_type: string;
  instance_count: number;
  availability_zone: string;
  start_time: string;
  end_time: string;
  duration_hours: number;
  state: string;
  total_cost: number;
}

export interface CapacityBlockRequest {
  instance_type: string;
  instance_count: number;
  availability_zone?: string;
  start_time: string;
  duration_hours: number;
}

export interface S3Mount {
  instance_name: string;
  bucket_name: string;
  mount_path: string;
  method: string;
  read_only: boolean;
  status: string;
}

export interface StorageAnalyticsSummary {
  storage_name: string;
  type: string;
  period: string;
  usage_percent: number;
  total_cost: number;
  daily_cost: number;
  recommendations: string[];
}

