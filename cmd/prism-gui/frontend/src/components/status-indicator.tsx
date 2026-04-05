import { cn } from '../lib/utils'

type StatusType = "success" | "error" | "warning" | "info" | "stopped" | "in-progress" | "pending"

interface StatusIndicatorProps {
  type?: StatusType
  children?: React.ReactNode
  className?: string
}

const dotClasses: Record<StatusType, string> = {
  success: "bg-green-500",
  error: "bg-red-500",
  warning: "bg-yellow-500",
  info: "bg-blue-500",
  stopped: "bg-gray-400",
  "in-progress": "bg-blue-500 animate-pulse",
  pending: "bg-yellow-400 animate-pulse",
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
