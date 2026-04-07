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
        <main id="main-content" role="main" tabIndex={-1} className="flex-1 overflow-auto p-4">
          {children}
        </main>
      </SidebarInset>
      <Toaster richColors position="bottom-right" expand={false} closeButton />
    </SidebarProvider>
  )
}
