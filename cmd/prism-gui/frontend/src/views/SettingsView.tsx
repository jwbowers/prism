import React from 'react';
import packageJson from '../../package.json';
import {
  Container,
  Header,
  SpaceBetween,
  Button,
  StatusIndicator,
  Badge,
  FormField,
  Input,
  Select,
  Alert,
  Spinner,
  Box,
  ColumnLayout,
  Toggle
} from '../lib/cloudscape-shim';
import { PlaceholderView } from './placeholder-view';

type SettingsSection = 'general' | 'profiles' | 'users' | 'ami' | 'rightsizing' | 'policy' | 'marketplace' | 'idle' | 'logs';

export interface UpdateInfo {
  current_version: string;
  latest_version: string;
  is_update_available: boolean;
  install_method?: string;
  update_command?: string;
  published_at?: string;
  release_url?: string;
}

export interface SettingsViewProps {
  settingsSection: SettingsSection;
  onSectionChange: (section: SettingsSection) => void;
  // General section data
  connected: boolean;
  loading: boolean;
  instanceCount: number;
  efsVolumeCount: number;
  ebsVolumeCount: number;
  updateInfo: UpdateInfo | null;
  autoStartEnabled: boolean;
  onRefresh: () => void;
  onSetAutoStart: (enabled: boolean) => Promise<void>;
  // Pre-rendered sub-views
  subViews: {
    profiles: React.ReactNode;
    users: React.ReactNode;
    ami: React.ReactNode;
    marketplace: React.ReactNode;
    idle: React.ReactNode;
    logs: React.ReactNode;
  };
}

const getStatusLabel = (context: string, status: string, additionalInfo?: string): string => {
  const labels: Record<string, Record<string, string>> = {
    connection: {
      success: 'Connected to daemon',
      error: 'Disconnected from daemon'
    },
    auth: {
      authenticated: 'Authenticated'
    },
    policy: {
      enabled: 'Policy enabled',
      disabled: 'Policy disabled'
    }
  };
  const label = labels[context]?.[status] || `${context} ${status}`;
  return additionalInfo ? `${label}: ${additionalInfo}` : label;
};

export const SettingsView: React.FC<SettingsViewProps> = ({
  settingsSection,
  onSectionChange,
  connected,
  loading,
  instanceCount,
  efsVolumeCount,
  ebsVolumeCount,
  updateInfo,
  autoStartEnabled,
  onRefresh,
  onSetAutoStart,
  subViews,
}) => {
  const navItems: Array<{ section: SettingsSection; label: string }> = [
    { section: 'general', label: 'General' },
    { section: 'profiles', label: 'Profiles' },
    { section: 'users', label: 'Users' },
  ];
  const advancedItems: Array<{ section: SettingsSection; label: string }> = [
    { section: 'ami', label: 'AMI Management' },
    { section: 'rightsizing', label: 'Rightsizing' },
    { section: 'policy', label: 'Policy Framework' },
    { section: 'marketplace', label: 'Template Marketplace' },
    { section: 'idle', label: 'Idle Detection' },
    { section: 'logs', label: 'Logs Viewer' },
  ];

  const renderSettingsContent = () => {
    switch (settingsSection) {
      case 'general':
        return (
          <SpaceBetween size="l">
            <Header
              variant="h1"
              description="Configure Prism preferences and system settings"
              actions={
                <SpaceBetween direction="horizontal" size="xs">
                  <Button onClick={onRefresh} disabled={loading}>
                    {loading ? <Spinner /> : 'Refresh'}
                  </Button>
                  <Button variant="primary">
                    Save Settings
                  </Button>
                </SpaceBetween>
              }
            >
              General Settings
            </Header>

      {/* System Status Section */}
      <Container
        header={
          <Header
            variant="h2"
            description="System status and daemon configuration"
          >
            System Status
          </Header>
        }
      >
        <ColumnLayout columns={3} variant="text-grid">
          <SpaceBetween size="m">
            <Box variant="awsui-key-label">Daemon Status</Box>
            <StatusIndicator
              type={connected ? 'success' : 'error'}
              iconAriaLabel={getStatusLabel('connection', connected ? 'success' : 'error')}
            >
              {connected ? 'Connected' : 'Disconnected'}
            </StatusIndicator>
            <Box color="text-body-secondary">
              Prism daemon on port 8947
            </Box>
          </SpaceBetween>
          <SpaceBetween size="m">
            <Box variant="awsui-key-label">API Version</Box>
            <Box fontSize="heading-m">v{packageJson.version}</Box>
            <Box color="text-body-secondary">
              Current Prism version
            </Box>
          </SpaceBetween>
          <SpaceBetween size="m">
            <Box variant="awsui-key-label">Active Resources</Box>
            <Box fontSize="heading-m">{instanceCount + efsVolumeCount + ebsVolumeCount}</Box>
            <Box color="text-body-secondary">
              Workspaces, EFS and EBS volumes
            </Box>
          </SpaceBetween>
        </ColumnLayout>
      </Container>

      {/* Update Information Section */}
      <Container
        header={
          <Header
            variant="h2"
            description="Check for and install Prism updates"
          >
            Update Information
          </Header>
        }
      >
        {updateInfo ? (
          <SpaceBetween size="m">
            <ColumnLayout columns={3} variant="text-grid">
              <SpaceBetween size="m">
                <Box variant="awsui-key-label">Current Version</Box>
                <Box fontSize="heading-m">{updateInfo.current_version}</Box>
              </SpaceBetween>
              <SpaceBetween size="m">
                <Box variant="awsui-key-label">Latest Version</Box>
                <Box fontSize="heading-m">
                  {updateInfo.latest_version}
                  {updateInfo.is_update_available && (
                    <span style={{ marginLeft: '8px' }}><Badge color="green">New</Badge></span>
                  )}
                </Box>
              </SpaceBetween>
              <SpaceBetween size="m">
                <Box variant="awsui-key-label">Status</Box>
                <StatusIndicator
                  type={updateInfo.is_update_available ? 'info' : 'success'}
                >
                  {updateInfo.is_update_available ? 'Update Available' : 'Up to Date'}
                </StatusIndicator>
              </SpaceBetween>
            </ColumnLayout>

            {updateInfo.is_update_available && (
              <Alert type="info" header="New version available">
                <SpaceBetween size="s">
                  <div><strong>Installation method:</strong> {updateInfo.install_method}</div>
                  <div><strong>Update command:</strong> <code>{updateInfo.update_command}</code></div>
                  <div><strong>Published:</strong> {new Date(updateInfo.published_at || '').toLocaleDateString()}</div>
                  <div>
                    <a href={updateInfo.release_url} target="_blank" rel="noopener noreferrer">
                      View release notes →
                    </a>
                  </div>
                </SpaceBetween>
              </Alert>
            )}
          </SpaceBetween>
        ) : (
          <Alert type="info">Checking for updates...</Alert>
        )}
      </Container>

      {/* Configuration Section */}
      <Container
        header={
          <Header
            variant="h2"
            description="Prism configuration and preferences"
          >
            Configuration
          </Header>
        }
      >
        <SpaceBetween size="l">
          <FormField
            label="Auto-refresh interval"
            description="How often the GUI should refresh data from the daemon"
          >
            <Select
              selectedOption={{ label: "30 seconds", value: "30" }}
              onChange={() => {}}
              options={[
                { label: "15 seconds", value: "15" },
                { label: "30 seconds", value: "30" },
                { label: "1 minute", value: "60" },
                { label: "2 minutes", value: "120" }
              ]}
            />
          </FormField>

          <FormField
            label="Default workspace size"
            description="Default size for new workspaces when launching templates"
          >
            <Select
              selectedOption={{ label: "Medium (M)", value: "M" }}
              onChange={() => {}}
              options={[
                { label: "Small (S)", value: "S" },
                { label: "Medium (M)", value: "M" },
                { label: "Large (L)", value: "L" },
                { label: "Extra Large (XL)", value: "XL" }
              ]}
            />
          </FormField>

          <FormField
            label="Show advanced features"
            description="Display advanced management options like hibernation policies and cost tracking"
          >
            <Select
              selectedOption={{ label: "Enabled", value: "enabled" }}
              onChange={() => {}}
              options={[
                { label: "Enabled", value: "enabled" },
                { label: "Disabled", value: "disabled" }
              ]}
            />
          </FormField>

          <FormField
            label="Start at login"
            description="Automatically start Prism GUI when you log in to your computer"
          >
            <Toggle
              checked={autoStartEnabled || false}
              onChange={async ({ detail }) => {
                await onSetAutoStart(detail.checked);
              }}
            >
              {autoStartEnabled ? 'Enabled' : 'Disabled'}
            </Toggle>
          </FormField>
        </SpaceBetween>
      </Container>

      {/* Profile and Authentication Section */}
      <Container
        header={
          <Header
            variant="h2"
            description="AWS profile and authentication settings"
          >
            AWS Configuration
          </Header>
        }
      >
        <ColumnLayout columns={2}>
          <SpaceBetween size="m">
            <FormField
              label="AWS Profile"
              description="Current AWS profile for authentication"
            >
              <Input
                value="aws"
                readOnly
                placeholder="AWS profile name"
              />
            </FormField>
            <FormField
              label="AWS Region"
              description="Current AWS region for resources"
            >
              <Input
                value="us-west-2"
                readOnly
                placeholder="AWS region"
              />
            </FormField>
          </SpaceBetween>
          <SpaceBetween size="m">
            <Box variant="strong">Authentication Status</Box>
            <StatusIndicator
              type="success"
              iconAriaLabel={getStatusLabel('auth', 'authenticated')}
            >
              Authenticated via AWS profile
            </StatusIndicator>
            <Box color="text-body-secondary">
              Using credentials from AWS profile "aws" in region us-west-2.
              Prism automatically manages authentication for all API calls.
            </Box>
          </SpaceBetween>
        </ColumnLayout>
      </Container>

      {/* Feature Management */}
      <Container
        header={
          <Header
            variant="h2"
            description="Enable or disable Prism features"
          >
            Feature Management
          </Header>
        }
      >
        <SpaceBetween size="m">
          {[
            { name: "Workspace Management", status: "enabled", description: "Launch, manage, and connect to cloud workspaces" },
            { name: "Storage Management", status: "enabled", description: "EFS and EBS volume operations" },
            { name: "Project Management", status: "enabled", description: "Multi-user collaboration and budget tracking" },
            { name: "User Management", status: "enabled", description: "Research users with persistent identity" },
            { name: "Hibernation Policies", status: "enabled", description: "Automated cost optimization through hibernation" },
            { name: "Cost Tracking", status: "partial", description: "Budget monitoring and expense analysis" },
            { name: "Template Marketplace", status: "partial", description: "Community template sharing and discovery" },
            { name: "Scaling Predictions", status: "partial", description: "Resource optimization recommendations" }
          ].map((feature) => (
            <Box key={feature.name}>
              <SpaceBetween direction="horizontal" size="s">
                <span style={{ fontWeight: 'bold', minWidth: '200px', display: 'inline-block' }}>{feature.name}:</span>
                <StatusIndicator
                  type={
                    feature.status === 'enabled' ? 'success' :
                    feature.status === 'partial' ? 'warning' : 'error'
                  }
                  iconAriaLabel={getStatusLabel('policy', feature.status, feature.name)}
                >
                  {feature.status}
                </StatusIndicator>
                <Box color="text-body-secondary">{feature.description}</Box>
              </SpaceBetween>
            </Box>
          ))}
        </SpaceBetween>
      </Container>

      {/* Debug and Troubleshooting */}
      <Container
        header={
          <Header
            variant="h2"
            description="Debug information and troubleshooting tools"
          >
            Debug & Troubleshooting
          </Header>
        }
      >
        <SpaceBetween size="m">
          <Alert type="info">
            <Box variant="strong">Debug Mode</Box>
            <Box variant="p">
              Console logging is enabled. Check browser developer tools for detailed API interactions and error messages.
            </Box>
          </Alert>

          <ColumnLayout columns={2}>
            <SpaceBetween size="s">
              <Box variant="strong">Quick Actions</Box>
              <Button>Test API Connection</Button>
              <Button>Refresh All Data</Button>
              <Button>Clear Notifications</Button>
              <Button>Export Configuration</Button>
            </SpaceBetween>
            <SpaceBetween size="s">
              <Box variant="strong">Troubleshooting</Box>
              <Button variant="link" external>View Documentation</Button>
              <Button variant="link" external>GitHub Issues</Button>
              <Button variant="link" external>Troubleshooting Guide</Button>
            </SpaceBetween>
          </ColumnLayout>
        </SpaceBetween>
      </Container>
          </SpaceBetween>
        );

      case 'profiles':
        return <>{subViews.profiles}</>;

      case 'users':
        return <>{subViews.users}</>;

      case 'ami':
        return <>{subViews.ami}</>;

      case 'rightsizing':
        return <PlaceholderView title="Rightsizing Recommendations" description="Workspace rightsizing recommendations will help optimize your costs by suggesting better-sized workspaces based on actual usage patterns." />;

      case 'policy':
        return <PlaceholderView title="Policy Management" description="Policy management allows you to configure institutional policies, access controls, and governance rules for your Prism deployment." />;

      case 'marketplace':
        return <>{subViews.marketplace}</>;

      case 'idle':
        return <>{subViews.idle}</>;

      case 'logs':
        return <>{subViews.logs}</>;

      default:
        return (
          <Alert type="error">
            Unknown settings section: {settingsSection}
          </Alert>
        );
    }
  };

  const navButton = (section: SettingsSection, label: string) => (
    <button
      key={section}
      data-testid={`settings-nav-${section}`}
      onClick={() => onSectionChange(section)}
      className={[
        'block w-full text-left px-3 py-1.5 rounded text-sm cursor-pointer border-0',
        settingsSection === section
          ? 'bg-sidebar-accent text-sidebar-primary font-semibold'
          : 'bg-transparent text-sidebar-foreground font-normal hover:bg-sidebar-accent/50',
      ].join(' ')}
    >
      {label}
    </button>
  );

  return (
    <div className="flex h-full lg:flex-row flex-col">
      {/* Vertical sidebar — visible at lg+ */}
      <div className="hidden lg:flex lg:w-[220px] shrink-0 border-r border-border px-2 py-4 flex-col">
        <div className="px-3 pb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wide">
          Settings
        </div>
        {navItems.map(item => navButton(item.section, item.label))}
        <hr className="my-2 border-border" />
        <div className="px-3 pb-1 text-xs font-semibold text-muted-foreground/60 uppercase tracking-wide">
          Advanced
        </div>
        {advancedItems.map(item => navButton(item.section, item.label))}
      </div>
      {/* Horizontal tab bar — visible below lg */}
      <div className="flex lg:hidden overflow-x-auto border-b border-border px-2 gap-1 shrink-0 py-1">
        {[...navItems, ...advancedItems].map(({ section, label }) => (
          <button
            key={section}
            onClick={() => onSectionChange(section)}
            className={[
              'shrink-0 px-3 py-2 text-sm whitespace-nowrap border-b-2 -mb-[1px] border-0 bg-transparent cursor-pointer',
              settingsSection === section
                ? 'border-primary text-primary font-semibold'
                : 'border-transparent text-muted-foreground hover:text-foreground',
            ].join(' ')}
          >
            {label}
          </button>
        ))}
      </div>
      <div className="flex-1 p-5 overflow-auto">
        {renderSettingsContent()}
      </div>
    </div>
  );
};
