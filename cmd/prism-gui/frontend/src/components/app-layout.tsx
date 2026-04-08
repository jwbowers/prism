import { Toaster } from 'sonner'
import { SidebarInset, SidebarProvider } from './ui/sidebar'

interface AppLayoutProps {
  sidebar: React.ReactNode
  children: React.ReactNode
}

export function AppLayout({ sidebar, children }: AppLayoutProps) {
  return (
    <SidebarProvider>
      {sidebar}
      <SidebarInset>
        <div className="flex-1 overflow-auto p-4">
          {children}
        </div>
      </SidebarInset>
      <Toaster richColors position="bottom-right" expand={false} closeButton />
    </SidebarProvider>
  )
}
