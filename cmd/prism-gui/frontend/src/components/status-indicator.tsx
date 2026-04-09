import { cn } from '../lib/utils'

type StatusType = "success" | "error" | "warning" | "info" | "stopped" | "in-progress" | "pending"

interface StatusIndicatorProps {
  type?: StatusType
  children?: React.ReactNode
  className?: string
}

const dotClasses: Record<StatusType, string> = {
  success: "bg-success",
  error: "bg-destructive",
  warning: "bg-warning",
  info: "bg-primary",
  stopped: "bg-muted-foreground",
  "in-progress": "bg-primary animate-pulse",
  pending: "bg-warning animate-pulse",
}

export function StatusIndicator({ type = "info", children, className }: StatusIndicatorProps) {
  return (
    <span className={cn("inline-flex items-center gap-1.5", className)}>
      <span
        aria-hidden="true"
        className={cn("inline-block h-2 w-2 rounded-full flex-shrink-0", dotClasses[type])}
      />
      {children && <span>{children}</span>}
    </span>
  )
}
