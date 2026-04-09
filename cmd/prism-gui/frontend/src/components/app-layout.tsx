import { useRef, useEffect } from 'react'
import { Toaster } from 'sonner'
import { SidebarInset, SidebarProvider } from './ui/sidebar'

interface AppLayoutProps {
  sidebar: React.ReactNode
  children: React.ReactNode
  viewKey?: string
}

export function AppLayout({ sidebar, children, viewKey }: AppLayoutProps) {
  const contentRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (typeof contentRef.current?.scrollTo === 'function') {
      contentRef.current.scrollTo(0, 0)
    }
  }, [viewKey])

  return (
    <SidebarProvider>
      {sidebar}
      <SidebarInset>
        <div ref={contentRef} className="flex-1 overflow-auto p-4">
          {children}
        </div>
      </SidebarInset>
      <Toaster richColors position="bottom-right" expand={false} closeButton />
    </SidebarProvider>
  )
}
