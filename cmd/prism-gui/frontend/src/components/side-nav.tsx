import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarSeparator,
} from './ui/sidebar'
import { Badge } from './ui/badge'
import type { AppState } from '../lib/types'

type ActiveView = AppState['activeView']

interface SideNavProps {
  activeView: ActiveView
  onNavigate: (view: ActiveView) => void
  instanceCount: number
  hasRunningInstances: boolean
  templateCount: number
  pendingInvitations: number
  activeCourses: number
  activeWorkshops: number
  budgetPoolCount: number
  pendingApprovalsCount: number
}

interface NavItem {
  view: ActiveView
  label: string
  badge?: React.ReactNode
}

interface NavSection {
  items: NavItem[]
  separator?: boolean
}

export function SideNav({
  activeView,
  onNavigate,
  instanceCount,
  hasRunningInstances,
  templateCount,
  pendingInvitations,
  activeCourses,
  activeWorkshops,
  budgetPoolCount,
  pendingApprovalsCount,
}: SideNavProps) {
  const sections: NavSection[] = [
    {
      items: [
        { view: 'dashboard', label: 'Dashboard' },
      ],
    },
    {
      separator: true,
      items: [
        {
          view: 'templates',
          label: 'Templates',
          badge: templateCount > 0 ? <Badge variant="secondary">{templateCount}</Badge> : undefined,
        },
        {
          view: 'workspaces',
          label: 'My Workspaces',
          badge: instanceCount > 0 ? (
            <Badge variant={hasRunningInstances ? 'default' : 'secondary'}>{instanceCount}</Badge>
          ) : undefined,
        },
      ],
    },
    {
      separator: true,
      items: [
        { view: 'storage', label: 'Storage' },
        { view: 'backups', label: 'Backups' },
        { view: 'projects', label: 'Projects' },
        { view: 'users', label: 'Users' },
        {
          view: 'invitations',
          label: 'Invitations',
          badge: pendingInvitations > 0 ? <Badge variant="secondary">{pendingInvitations}</Badge> : undefined,
        },
        {
          view: 'courses',
          label: 'Courses',
          badge: activeCourses > 0 ? <Badge variant="secondary">{activeCourses}</Badge> : undefined,
        },
        {
          view: 'workshops',
          label: 'Workshops',
          badge: activeWorkshops > 0 ? <Badge variant="default">{activeWorkshops}</Badge> : undefined,
        },
        { view: 'capacity-blocks', label: 'Capacity Blocks' },
        {
          view: 'budgets',
          label: 'Budgets',
          badge: budgetPoolCount > 0 ? <Badge variant="secondary">{budgetPoolCount}</Badge> : undefined,
        },
        {
          view: 'approvals',
          label: 'Approvals',
          badge: pendingApprovalsCount > 0 ? (
            <Badge variant="destructive">{pendingApprovalsCount}</Badge>
          ) : undefined,
        },
      ],
    },
    {
      separator: true,
      items: [
        { view: 'ami', label: 'AMI Management' },
        { view: 'marketplace', label: 'Marketplace' },
        { view: 'rightsizing', label: 'Rightsizing' },
        { view: 'policy', label: 'Policy' },
        { view: 'idle', label: 'Idle Detection' },
        { view: 'logs', label: 'Logs' },
        { view: 'settings', label: 'Settings' },
      ],
    },
  ]

  return (
    <Sidebar>
      <SidebarHeader className="border-b px-4 py-3">
        <a href="/" className="text-lg font-semibold tracking-tight no-underline text-foreground">Prism</a>
      </SidebarHeader>
      <SidebarContent>
        {sections.map((section, si) => (
          <SidebarGroup key={si}>
            {section.separator && <SidebarSeparator />}
            <SidebarGroupContent>
              <SidebarMenu>
                {section.items.map((item) => (
                  <SidebarMenuItem key={item.view}>
                    <SidebarMenuButton
                      isActive={activeView === item.view}
                      onClick={() => onNavigate(item.view)}
                      className="justify-between"
                    >
                      <span>{item.label}</span>
                      {item.badge}
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>
    </Sidebar>
  )
}
