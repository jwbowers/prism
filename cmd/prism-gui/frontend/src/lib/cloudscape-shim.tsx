/**
 * Cloudscape compatibility shim.
 * Provides stub components mapping Cloudscape API to shadcn/Tailwind equivalents.
 * These are intentionally minimal — Phase 2-7 replace each view with native shadcn.
 *
 * Accepts `any` props to avoid TypeScript churn during migration.
 */
/* eslint-disable @typescript-eslint/no-explicit-any */
import React from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { Button as ShadButton } from '../components/ui/button'
import { Badge as ShadBadge } from '../components/ui/badge'
import { Input as ShadInput } from '../components/ui/input'
import { Textarea as ShadTextarea } from '../components/ui/textarea'
import { Checkbox as ShadCheckbox } from '../components/ui/checkbox'
import { Switch as ShadSwitch } from '../components/ui/switch'
import { Progress as ShadProgress } from '../components/ui/progress'
import { Alert as ShadAlert, AlertDescription } from '../components/ui/alert'
import { Spinner as SpinnerBase } from '../components/ui/spinner'
import { cn } from './utils'

// ── Layout ─────────────────────────────────────────────────────────────────

export function AppLayout({ content, navigation, navigationOpen: _no, onNavigationChange: _onNavChange, navigationHide: _nh, toolsHide: _th, maxContentWidth: _mw, contentType: _ct, breadcrumbs: _bc, notifications: _notifs, tools: _tools }: {
  content?: React.ReactNode
  navigation?: React.ReactNode
  navigationOpen?: boolean
  onNavigationChange?: (e: CE<{ open: boolean }>) => void
  navigationHide?: boolean
  toolsHide?: boolean
  maxContentWidth?: number
  contentType?: string
  breadcrumbs?: React.ReactNode
  notifications?: React.ReactNode
  tools?: React.ReactNode
  [k: string]: any
}) {
  return (
    <div className="flex h-screen overflow-hidden">
      {navigation && <aside className="w-64 border-r bg-sidebar flex-shrink-0 overflow-y-auto">{navigation}</aside>}
      <main id="main-content" className="flex-1 overflow-y-auto p-4" tabIndex={-1}>{content}</main>
    </div>
  )
}

export function SideNavigation({ header, items, onFollow, activeHref }: {
  header?: { text: string; href: string }
  items?: Array<{ type: string; text?: string; href?: string; items?: any[]; info?: React.ReactNode; external?: boolean; defaultExpanded?: boolean }>
  onFollow?: (e: { detail: { href: string; text?: string; external?: boolean }; preventDefault: () => void; stopPropagation: () => void }) => void
  activeHref?: string
  [k: string]: any
}) {
  const renderItems = (items: any[]) => items?.map((item: any, i: number) => {
    if (item.type === 'section') {
      return (
        <div key={i} className="mb-2">
          <div className="px-3 py-1 text-xs font-semibold text-muted-foreground uppercase tracking-wider">{item.text}</div>
          {renderItems(item.items)}
        </div>
      )
    }
    if (item.type === 'divider') return <hr key={i} className="my-2 border-sidebar-border" />
    const isActive = activeHref === item.href
    return (
      <a
        key={i}
        href={item.href}
        onClick={(e) => { e.preventDefault(); onFollow?.({ detail: { href: item.href, text: item.text, external: item.external }, preventDefault: () => e.preventDefault(), stopPropagation: () => e.stopPropagation() }) }}
        className={cn(
          'flex items-center gap-2 px-3 py-1.5 text-sm rounded-md mx-1 cursor-pointer',
          isActive ? 'bg-sidebar-accent text-sidebar-primary font-medium' : 'text-sidebar-foreground hover:bg-sidebar-accent/50'
        )}
      >
        {item.text}
        {item.info && <span className="ml-auto">{item.info}</span>}
      </a>
    )
  })
  return (
    <div className="py-2">
      {header && (
        <a href={header.href} onClick={(e) => { e.preventDefault(); onFollow?.({ detail: { href: header.href }, preventDefault: () => e.preventDefault(), stopPropagation: () => e.stopPropagation() }) }}
          className="flex items-center gap-2 px-3 py-2 font-semibold text-sm mb-2">
          {header.text}
        </a>
      )}
      {renderItems(items || [])}
    </div>
  )
}

export function Container({ header, children, className, ...rest }: any) {
  return (
    <div className={cn('rounded-lg border bg-card text-card-foreground shadow-sm', className)} {...rest}>
      {header && <div className="flex items-center justify-between border-b px-4 py-3">{header}</div>}
      <div className="p-4">{children}</div>
    </div>
  )
}

export function Header({ children, variant, actions, counter, description }: any) {
  const Tag = variant === 'h1' ? 'h1' : variant === 'h3' ? 'h3' : 'h2'
  return (
    <div className="flex items-center justify-between w-full gap-2">
      <div>
        <Tag className={cn('font-semibold', variant === 'h1' ? 'text-2xl' : variant === 'h3' ? 'text-base' : 'text-lg')}>
          {children}{counter && <span className="ml-1 text-muted-foreground text-sm font-normal">{counter}</span>}
        </Tag>
        {description && <p className="text-sm text-muted-foreground mt-0.5">{description}</p>}
      </div>
      {actions && <div className="flex items-center gap-2 flex-shrink-0">{actions}</div>}
    </div>
  )
}

export function SpaceBetween({ direction = 'vertical', size = 'm', children, ...rest }: any) {
  const gapMap: Record<string, string> = { xs: '1', s: '2', m: '4', l: '6', xl: '8', xxl: '12' }
  const gap = gapMap[size] || '4'
  if (direction === 'horizontal') {
    return <div className={`flex flex-wrap items-center gap-${gap}`} {...rest}>{children}</div>
  }
  return <div className={`flex flex-col gap-${gap}`} {...rest}>{children}</div>
}

export function ColumnLayout({ columns = 2, children, variant: _v, ...rest }: any) {
  return <div className={`grid grid-cols-${columns} gap-4`} {...rest}>{children}</div>
}

export function Box({ variant, fontSize, color, textAlign, float, padding, children, as: As = 'div', ...rest }: any) {
  const fontClasses: Record<string, string> = {
    'display-l': 'text-4xl font-bold',
    'heading-xl': 'text-2xl font-bold',
    'heading-l': 'text-xl font-semibold',
    'heading-m': 'text-lg font-semibold',
    'heading-s': 'text-base font-semibold',
    'heading-xs': 'text-sm font-semibold',
    'body-m': 'text-sm',
    'body-s': 'text-xs',
    'label': 'text-sm font-medium',
    'code': 'font-mono text-sm',
  }
  const colorClasses: Record<string, string> = {
    'text-status-info': 'text-primary dark:text-primary',
    'text-status-success': 'text-green-600 dark:text-green-400',
    'text-status-error': 'text-red-600 dark:text-red-400',
    'text-status-warning': 'text-yellow-600 dark:text-yellow-400',
    'text-label': 'text-foreground',
    'text-body-secondary': 'text-muted-foreground',
  }
  if (variant === 'p') As = 'p'
  if (variant === 'strong') As = 'strong'
  if (variant === 'small') As = 'small'
  if (variant === 'span') As = 'span'
  if (variant === 'code') As = 'code'
  const classes = cn(
    fontSize && fontClasses[fontSize],
    color && colorClasses[color],
    textAlign && `text-${textAlign}`,
    float && `float-${float}`,
    padding && `p-${padding}`,
  )
  return <As className={classes || undefined} {...rest}>{children}</As>
}

export function TextContent({ children }: any) {
  return <div className="prose prose-sm dark:prose-invert max-w-none">{children}</div>
}

export function Link({ href, onFollow, external, children, variant: _variant, fontSize, ...rest }: any) {
  const cls = cn('text-primary hover:underline cursor-pointer', fontSize === 'body-s' && 'text-xs')
  if (href) {
    return <a href={href} target={external ? '_blank' : undefined} rel={external ? 'noopener noreferrer' : undefined} className={cls} {...rest}>{children}</a>
  }
  return <span onClick={() => onFollow?.()} className={cls} role="link" tabIndex={0} {...rest}>{children}</span>
}

// ── Cloudscape event types (give contextual typing so callers can destructure { detail }) ──
// Note: Cloudscape events have a optional preventDefault() method on the event object itself.
type CE<T = any> = { detail: T; preventDefault?: () => void; stopPropagation?: () => void }

// ── Forms ──────────────────────────────────────────────────────────────────

// Button: accepts Cloudscape variant/size strings and maps them to shadcn equivalents
export function Button({ children, variant: cv, size: cs, iconName: _icon, loading, disabled, onClick, ...rest }: {
  children?: React.ReactNode
  variant?: string
  size?: string
  iconName?: string
  loading?: boolean
  disabled?: boolean
  onClick?: () => void
  [k: string]: any
}) {
  const variantMap: Record<string, 'default' | 'outline' | 'link' | 'ghost' | 'destructive' | 'secondary'> = {
    primary: 'default',
    normal: 'outline',
    link: 'link',
    'inline-link': 'link',
    icon: 'ghost',
    destructive: 'destructive',
  }
  const sizeMap: Record<string, 'default' | 'sm' | 'lg' | 'icon'> = {
    large: 'lg',
    small: 'sm',
    icon: 'icon',
  }
  return (
    <ShadButton
      variant={variantMap[cv ?? ''] ?? 'default'}
      size={sizeMap[cs ?? ''] ?? 'default'}
      disabled={disabled || loading}
      onClick={onClick}
      {...rest}
    >
      {loading && <Spinner size="sm" className="mr-2" />}
      {children}
    </ShadButton>
  )
}

export function Input({ value, onChange, placeholder, disabled, type = 'text', invalid, ...rest }: {
  value?: string
  onChange?: (e: CE<{ value: string }>) => void
  placeholder?: string
  disabled?: boolean
  type?: string
  invalid?: boolean
  [k: string]: any
}) {
  // Wrap in a div so data-testid consumers can do wrapper.querySelector('input')
  const hasTestId = 'data-testid' in rest
  const { 'data-testid': testId, ...inputRest } = rest as { 'data-testid'?: string; [k: string]: any }
  const input = (
    <ShadInput
      value={value ?? ''}
      onChange={(e: React.ChangeEvent<HTMLInputElement>) => onChange?.({ detail: { value: e.target.value } })}
      placeholder={placeholder}
      disabled={disabled}
      type={type}
      className={cn(invalid && 'border-destructive')}
      data-invalid={invalid || undefined}
      {...inputRest}
    />
  )
  if (hasTestId) return <div data-testid={testId}>{input}</div>
  return input
}

export function Textarea({ value, onChange, placeholder, disabled, rows, ...rest }: {
  value?: string
  onChange?: (e: CE<{ value: string }>) => void
  placeholder?: string
  disabled?: boolean
  rows?: number
  [k: string]: any
}) {
  const { 'data-testid': testId, ...taRest } = rest as { 'data-testid'?: string; [k: string]: any }
  const el = (
    <ShadTextarea
      value={value ?? ''}
      onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => onChange?.({ detail: { value: e.target.value } })}
      placeholder={placeholder}
      disabled={disabled}
      rows={rows}
      {...taRest}
    />
  )
  if (testId) return <div data-testid={testId}>{el}</div>
  return el
}

export function Select({ selectedOption, onChange, options, disabled, placeholder, invalid, ...rest }: {
  selectedOption?: { value: string; label: string; description?: string } | null
  onChange?: (e: CE<{ selectedOption: { value: string; label: string; description?: string } }>) => void
  options?: Array<{ value: string; label: string; disabled?: boolean; description?: string }>
  disabled?: boolean
  placeholder?: string
  invalid?: boolean
  [k: string]: any
}) {
  const { 'data-testid': testId, ...selectRest } = rest as { 'data-testid'?: string; [k: string]: any }
  const el = (
    <select
      value={selectedOption?.value ?? ''}
      onChange={(e) => {
        const opt = options?.find((o) => o.value === e.target.value) ?? { value: e.target.value, label: e.target.value }
        onChange?.({ detail: { selectedOption: opt } })
      }}
      disabled={disabled}
      className={cn(
        'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
        invalid && 'border-destructive'
      )}
      {...selectRest}
    >
      {placeholder && <option value="">{placeholder}</option>}
      {options?.map((o) => (
        <option key={o.value} value={o.value} disabled={o.disabled}>{o.label}</option>
      ))}
    </select>
  )
  if (testId) return <div data-testid={testId}>{el}</div>
  return el
}

export function Checkbox({ checked, onChange, disabled, children, description }: {
  checked?: boolean
  onChange?: (e: CE<{ checked: boolean }>) => void
  disabled?: boolean
  children?: React.ReactNode
  description?: string
  [k: string]: any
}) {
  return (
    <label className="flex items-start gap-2 cursor-pointer">
      <ShadCheckbox
        checked={checked ?? false}
        onCheckedChange={(v) => onChange?.({ detail: { checked: !!v } })}
        disabled={disabled}
        className="mt-0.5"
      />
      {(children || description) && (
        <span className="flex flex-col">
          {children && <span className="text-sm">{children}</span>}
          {description && <span className="text-xs text-muted-foreground">{description}</span>}
        </span>
      )}
    </label>
  )
}

export function Toggle({ checked, onChange, disabled, children, description }: {
  checked?: boolean
  onChange?: (e: CE<{ checked: boolean }>) => void
  disabled?: boolean
  children?: React.ReactNode
  description?: string
  [k: string]: any
}) {
  return (
    <label className="flex items-start gap-2 cursor-pointer">
      <ShadSwitch
        checked={checked ?? false}
        onCheckedChange={(v) => onChange?.({ detail: { checked: !!v } })}
        disabled={disabled}
      />
      {(children || description) && (
        <span className="flex flex-col">
          {children && <span className="text-sm">{children}</span>}
          {description && <span className="text-xs text-muted-foreground">{description}</span>}
        </span>
      )}
    </label>
  )
}

export function Form({ children, actions, header, errorText }: any) {
  return (
    <div className="space-y-4">
      {header && <div className="border-b pb-2 font-semibold">{header}</div>}
      {errorText && <div className="text-sm text-destructive">{errorText}</div>}
      {children}
      {actions && <div className="flex gap-2 pt-2 border-t">{actions}</div>}
    </div>
  )
}

export function FormField({ label, children, errorText, description, constraintText }: any) {
  // Generate a stable id from label text for proper htmlFor/id association (enables getByLabel in tests)
  const fieldId = typeof label === 'string'
    ? `field-${label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`
    : undefined
  // Clone child element to inject id so the label's htmlFor links to the input
  const childWithId = fieldId
    ? React.Children.map(children, (child) =>
        React.isValidElement(child) ? React.cloneElement(child as React.ReactElement<any>, { id: fieldId }) : child
      )
    : children
  return (
    <div className="space-y-1.5">
      {label && <label htmlFor={fieldId} className="text-sm font-medium leading-none">{label}</label>}
      {description && <p className="text-xs text-muted-foreground">{description}</p>}
      {childWithId}
      {constraintText && <p className="text-xs text-muted-foreground">{constraintText}</p>}
      {errorText && <p className="text-xs text-destructive">{errorText}</p>}
    </div>
  )
}

// ── Feedback ───────────────────────────────────────────────────────────────

export function Alert({ type = 'info', header, children, dismissible, onDismiss, action, ...rest }: any) {
  const variantMap: Record<string, string> = {
    error: 'border-destructive/50 text-destructive',
    warning: 'border-yellow-500/50 text-yellow-700 dark:text-yellow-400',
    success: 'border-green-500/50 text-green-700 dark:text-green-400',
    info: '',
  }
  return (
    <ShadAlert className={variantMap[type]} {...rest}>
      {header && <div className="font-semibold mb-1 text-sm">{header}</div>}
      <AlertDescription>{children}</AlertDescription>
      {(dismissible || action) && (
        <div className="flex gap-2 mt-2">
          {action}
          {dismissible && <ShadButton variant="ghost" size="sm" onClick={onDismiss}>Dismiss</ShadButton>}
        </div>
      )}
    </ShadAlert>
  )
}

export function Flashbar({ items }: any) {
  if (!items?.length) return null
  return (
    <div className="space-y-2 mb-4">
      {items.map((item: any, i: number) => (
        <Alert key={item.id ?? i} type={item.type} header={item.header} dismissible={item.dismissible} onDismiss={item.onDismiss}>
          {item.content}
        </Alert>
      ))}
    </div>
  )
}

// Spinner: wraps our spinner to accept Cloudscape size strings
export function Spinner({ size: cs, className }: { size?: string; className?: string }) {
  const sizeMap: Record<string, 'sm' | 'md' | 'lg'> = { small: 'sm', normal: 'md', large: 'lg', big: 'lg' }
  return <SpinnerBase size={sizeMap[cs ?? ''] ?? 'md'} className={className} />
}

export function StatusIndicator({ type = 'info', children }: any) {
  const dotClasses: Record<string, string> = {
    success: 'bg-green-500',
    error: 'bg-red-500',
    warning: 'bg-yellow-500',
    info: 'bg-primary',
    stopped: 'bg-muted-foreground',
    'in-progress': 'bg-primary animate-pulse',
    pending: 'bg-yellow-400 animate-pulse',
    loading: 'bg-primary animate-pulse',
  }
  return (
    <span className="inline-flex items-center gap-1.5">
      <span aria-hidden="true" className={cn('inline-block h-2 w-2 rounded-full flex-shrink-0', dotClasses[type] ?? 'bg-gray-400')} />
      {children && <span className="text-sm">{children}</span>}
    </span>
  )
}

export { ShadBadge as Badge }

export function ProgressBar({ value = 0, label, description, status }: any) {
  return (
    <div className="space-y-1">
      {(label || description) && (
        <div className="flex justify-between text-sm">
          {label && <span>{label}</span>}
          {description && <span className="text-muted-foreground">{description}</span>}
        </div>
      )}
      <ShadProgress value={value} className={cn(status === 'error' && '[&>div]:bg-destructive')} />
    </div>
  )
}

// ── Table ──────────────────────────────────────────────────────────────────

export function Table({ columnDefinitions, items, loading, loadingText, empty, header, filter, pagination, preferences, selectionType, selectedItems, onSelectionChange, trackBy, ariaLabels, ...rest }: {
  columnDefinitions?: Array<{ id?: string; header?: React.ReactNode; cell?: (item: any) => React.ReactNode; sortingField?: string; width?: string | number }>
  items?: any[]
  loading?: boolean
  loadingText?: string
  empty?: React.ReactNode
  header?: React.ReactNode
  filter?: React.ReactNode
  pagination?: React.ReactNode
  preferences?: React.ReactNode
  selectionType?: 'multi' | 'single'
  selectedItems?: any[]
  onSelectionChange?: (e: CE<{ selectedItems: any[] }>) => void
  onSortingChange?: (e: CE<{ sortingColumn: any; isDescending?: boolean }>) => void
  onRowClick?: (e: CE<{ item: any; rowIndex?: number }>) => void
  sortingColumn?: any
  sortingDescending?: boolean
  sortingDisabled?: boolean
  trackBy?: string
  ariaLabels?: Record<string, any>
  [k: string]: any
}) {
  const columns = columnDefinitions ?? []
  const rows = items ?? []
  return (
    <div className="space-y-2" {...rest}>
      {header && <div>{header}</div>}
      {filter && <div>{filter}</div>}
      {loading ? (
        <div className="flex items-center justify-center h-24 gap-2 text-muted-foreground">
          <Spinner size="sm" />
          <span>{loadingText ?? 'Loading...'}</span>
        </div>
      ) : (
        <div className="rounded-md border overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                {selectionType === 'multi' && (
                  <th className="w-10 p-2">
                    <ShadCheckbox
                      checked={selectedItems?.length === rows.length && rows.length > 0}
                      onCheckedChange={(v) => onSelectionChange?.({ detail: { selectedItems: v ? rows : [] } })}
                    />
                  </th>
                )}
                {columns.map((col: any, i: number) => (
                  <th key={i} className="px-3 py-2 text-left font-medium text-muted-foreground">{col.header}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {rows.length === 0 ? (
                <tr><td colSpan={columns.length + (selectionType ? 1 : 0)} className="text-center p-8 text-muted-foreground">{empty ?? 'No items'}</td></tr>
              ) : (
                rows.map((item: any, ri: number) => {
                  const key = trackBy ? item[trackBy] : ri
                  const isSelected = selectedItems?.some((s: any) => trackBy ? s[trackBy] === item[trackBy] : s === item)
                  return (
                    <tr key={key} className={cn('border-b hover:bg-muted/30 transition-colors', isSelected && 'bg-muted/50')}>
                      {selectionType === 'multi' && (
                        <td className="w-10 p-2">
                          <ShadCheckbox
                            checked={isSelected ?? false}
                            onCheckedChange={(v) => {
                              const next = v
                                ? [...(selectedItems ?? []), item]
                                : (selectedItems ?? []).filter((s: any) => (trackBy ? s[trackBy] !== item[trackBy] : s !== item))
                              onSelectionChange?.({ detail: { selectedItems: next } })
                            }}
                          />
                        </td>
                      )}
                      {columns.map((col: any, ci: number) => (
                        <td key={ci} className="px-3 py-2">{col.cell ? col.cell(item) : item[col.id]}</td>
                      ))}
                    </tr>
                  )
                })
              )}
            </tbody>
          </table>
        </div>
      )}
      {pagination && <div className="flex justify-end">{pagination}</div>}
    </div>
  )
}

export function PropertyFilter({ query, onChange, filteringPlaceholder }: {
  query?: { tokens?: Array<{ value: string; operator: string }>; operation?: string }
  onChange?: (e: CE<{ tokens: Array<{ value: string; operator: string }>; operation: 'and' | 'or' }>) => void
  filteringPlaceholder?: string
  [k: string]: any
}) {
  return (
    <ShadInput
      placeholder={filteringPlaceholder ?? 'Filter...'}
      value={query?.tokens?.map((t) => t.value).join(' ') ?? ''}
      onChange={(e) => onChange?.({ detail: { tokens: e.target.value ? [{ value: e.target.value, operator: ':' }] : [], operation: 'and' } })}
      className="max-w-sm"
    />
  )
}

export function TextFilter({ filteringText, onChange, filteringPlaceholder }: {
  filteringText?: string
  onChange?: (e: CE<{ filteringText: string }>) => void
  filteringPlaceholder?: string
  [k: string]: any
}) {
  return (
    <ShadInput
      placeholder={filteringPlaceholder ?? 'Search...'}
      value={filteringText ?? ''}
      onChange={(e) => onChange?.({ detail: { filteringText: e.target.value } })}
      className="max-w-sm"
    />
  )
}

export function Pagination({ currentPageIndex, pagesCount, onChange, disabled, ariaLabels: _al }: {
  currentPageIndex?: number
  pagesCount?: number
  onChange?: (e: CE<{ currentPageIndex: number }>) => void
  disabled?: boolean
  ariaLabels?: { nextPageLabel?: string; previousPageLabel?: string; pageLabel?: (n: number) => string }
  [k: string]: any
}) {
  const cur = currentPageIndex ?? 1
  const max = pagesCount ?? 1
  return (
    <div className="flex items-center gap-1">
      <ShadButton variant="outline" size="sm" disabled={disabled || cur <= 1}
        onClick={() => onChange?.({ detail: { currentPageIndex: cur - 1 } })}>
        ‹
      </ShadButton>
      <span className="text-sm px-2">{cur} / {max}</span>
      <ShadButton variant="outline" size="sm" disabled={disabled || cur >= max}
        onClick={() => onChange?.({ detail: { currentPageIndex: cur + 1 } })}>
        ›
      </ShadButton>
    </div>
  )
}

// ── Overlays ───────────────────────────────────────────────────────────────

export function Modal({ visible, onDismiss, header, children, footer, size, ...rest }: any) {
  const sizeClass = size === 'large' ? 'max-w-4xl' : size === 'small' ? 'max-w-sm' : 'max-w-2xl'
  return (
    <AnimatePresence>
      {visible && (
        <div className="fixed inset-0 z-50 flex items-center justify-center" role="dialog" aria-modal="true" aria-label={typeof header === 'string' ? header : undefined} {...rest}>
          <motion.div
            className="absolute inset-0 bg-black/50"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, transition: { duration: 0.15 } }}
            exit={{ opacity: 0, transition: { duration: 0.1 } }}
            onClick={onDismiss}
          />
          <motion.div
            className={cn('relative bg-card rounded-lg shadow-xl w-full mx-4 flex flex-col max-h-[90vh]', sizeClass)}
            initial={{ opacity: 0, scale: 0.97, y: 10 }}
            animate={{ opacity: 1, scale: 1, y: 0, transition: { duration: 0.15, ease: [0.4, 0, 0.2, 1] } }}
            exit={{ opacity: 0, scale: 0.97, y: 4, transition: { duration: 0.1 } }}
          >
            {header && (
              <div className="flex items-center justify-between p-4 border-b">
                <div className="font-semibold text-lg">{header}</div>
                <ShadButton variant="ghost" size="sm" onClick={onDismiss} aria-label="Close">✕</ShadButton>
              </div>
            )}
            <div className="overflow-y-auto p-4 flex-1">{children}</div>
            {footer && <div className="p-4 border-t flex justify-end gap-2">{footer}</div>}
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}

export function ButtonDropdown({ items, onItemClick, disabled, children, variant, expandToViewport: _et }: {
  items?: Array<{ id?: string; text?: string; type?: string; disabled?: boolean; items?: any[]; description?: string; iconName?: string }>
  onItemClick?: (e: CE<{ id: string }>) => void
  disabled?: boolean
  children?: React.ReactNode
  variant?: string
  expandToViewport?: boolean
  [k: string]: any
}) {
  const [open, setOpen] = React.useState(false)
  const variantMap: Record<string, 'default' | 'outline' | 'ghost'> = { primary: 'default', normal: 'outline', icon: 'ghost' }
  return (
    <div className="relative inline-block">
      <ShadButton variant={variantMap[variant ?? ''] ?? 'outline'} disabled={disabled} onClick={() => setOpen(!open)}>
        {children} <span className="ml-1">▾</span>
      </ShadButton>
      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div role="menu" className="absolute right-0 z-50 mt-1 min-w-40 rounded-md border bg-popover shadow-md">
            {items?.map((item, i) => {
              if (item.type === 'divider') return <hr key={i} className="my-1" />
              return (
                <button
                  key={item.id ?? i}
                  role="menuitem"
                  disabled={item.disabled}
                  onClick={() => { setOpen(false); onItemClick?.({ detail: { id: item.id ?? '' } }) }}
                  className="flex w-full items-center px-3 py-1.5 text-sm hover:bg-accent disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {item.text}
                </button>
              )
            })}
          </div>
        </>
      )}
    </div>
  )
}

export function Tabs({ tabs, activeTabId, onChange, variant, ...rest }: {
  tabs?: Array<{ id: string; label: string; content?: React.ReactNode; disabled?: boolean }>
  activeTabId?: string
  onChange?: (e: CE<{ activeTabId: string }>) => void
  variant?: string
  [k: string]: any
}) {
  const [active, setActive] = React.useState(activeTabId ?? tabs?.[0]?.id)
  // Sync with controlled activeTabId prop
  React.useEffect(() => {
    if (activeTabId !== undefined) setActive(activeTabId)
  }, [activeTabId])
  const current = active ?? tabs?.[0]?.id
  return (
    <div {...rest}>
      <div role="tablist" className={cn('flex border-b', variant === 'container' && 'bg-muted/50 rounded-t-md px-2')}>
        {tabs?.map((tab) => (
          <button
            key={tab.id}
            role="tab"
            aria-selected={current === tab.id}
            disabled={tab.disabled}
            onClick={() => { setActive(tab.id); onChange?.({ detail: { activeTabId: tab.id } }) }}
            className={cn(
              'px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
              current === tab.id ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <div role="tabpanel" className="pt-4">
        {tabs?.find((t) => t.id === current)?.content}
      </div>
    </div>
  )
}

export function Wizard({ steps, activeStepIndex = 0, onNavigate, onSubmit, onCancel, isLoadingNextStep, submitButtonText = 'Submit', i18nStrings }: {
  steps?: Array<{ title: string; content: React.ReactNode; description?: string; isOptional?: boolean }>
  activeStepIndex?: number
  onNavigate?: (e: CE<{ requestedStepIndex: number; reason: string }>) => void
  onSubmit?: () => void
  onCancel?: () => void
  isLoadingNextStep?: boolean
  submitButtonText?: string
  i18nStrings?: {
    nextButton?: string
    previousButton?: string
    cancelButton?: string
    submitButton?: string
    optional?: string
    navigationAriaLabel?: string
    stepNumberLabel?: (stepNumber: number) => string
    collapsedStepsLabel?: (stepNumber: number, stepsCount: number) => string
    skipToButtonLabel?: (step: { title: string }) => string
    errorIconAriaLabel?: string
  }
  [k: string]: any
}) {
  return (
    <div className="space-y-4">
      <div className="flex gap-2 border-b pb-3">
        {steps?.map((step: any, i: number) => (
          <div key={i} className={cn('flex items-center gap-1 text-sm', i === activeStepIndex ? 'text-primary font-medium' : 'text-muted-foreground')}>
            <span className={cn('w-6 h-6 rounded-full flex items-center justify-center text-xs', i === activeStepIndex ? 'bg-primary text-primary-foreground' : 'bg-muted')}>{i + 1}</span>
            {step.title}
            {i < steps.length - 1 && <span className="mx-1">›</span>}
          </div>
        ))}
      </div>
      <div>{steps?.[activeStepIndex]?.content}</div>
      <div className="flex justify-between pt-2 border-t">
        <ShadButton variant="ghost" onClick={onCancel}>{i18nStrings?.cancelButton ?? 'Cancel'}</ShadButton>
        <div className="flex gap-2">
          {activeStepIndex > 0 && (
            <ShadButton variant="outline" onClick={() => onNavigate?.({ detail: { requestedStepIndex: activeStepIndex - 1, reason: 'previous' } })}>
              {i18nStrings?.previousButton ?? 'Previous'}
            </ShadButton>
          )}
          {activeStepIndex < (steps?.length ?? 1) - 1 ? (
            <ShadButton onClick={() => onNavigate?.({ detail: { requestedStepIndex: activeStepIndex + 1, reason: 'next' } })} disabled={isLoadingNextStep}>
              {isLoadingNextStep ? <><Spinner size="sm" className="mr-2" />{i18nStrings?.nextButton ?? 'Next'}</> : (i18nStrings?.nextButton ?? 'Next')}
            </ShadButton>
          ) : (
            <ShadButton onClick={onSubmit} disabled={isLoadingNextStep}>
              {isLoadingNextStep ? <><Spinner size="sm" className="mr-2" />{submitButtonText}</> : submitButtonText}
            </ShadButton>
          )}
        </div>
      </div>
    </div>
  )
}

// ── Content display ─────────────────────────────────────────────────────────

export function Cards({ cardDefinition, items, loading, loadingText, empty, header, filter, pagination, selectionType, selectedItems, onSelectionChange, trackBy, cardsPerRow }: {
  cardDefinition?: { header?: (item: any) => React.ReactNode; sections?: Array<{ id?: string; header?: React.ReactNode; content?: (item: any) => React.ReactNode }> }
  items?: any[]
  loading?: boolean
  loadingText?: string
  empty?: React.ReactNode
  header?: React.ReactNode
  filter?: React.ReactNode
  pagination?: React.ReactNode
  selectionType?: 'single' | 'multi'
  selectedItems?: any[]
  onSelectionChange?: (e: CE<{ selectedItems: any[] }>) => void
  trackBy?: string
  cardsPerRow?: Array<{ cards: number; minWidth?: number }>
  [k: string]: any
}) {
  const cols = cardsPerRow?.[0]?.cards ?? 3
  const rows = items ?? []
  return (
    <div className="space-y-2">
      {header && <div>{header}</div>}
      {filter && <div>{filter}</div>}
      {loading ? (
        <div className="flex items-center justify-center h-24 gap-2 text-muted-foreground">
          <Spinner size="sm" />
          <span>{loadingText ?? 'Loading...'}</span>
        </div>
      ) : rows.length === 0 ? (
        <div className="text-center p-8 text-muted-foreground">{empty ?? 'No items'}</div>
      ) : (
        <motion.div
          className={`grid grid-cols-${Math.min(cols, 4)} gap-4`}
          initial="hidden"
          animate="visible"
          variants={{ visible: { transition: { staggerChildren: 0.045 } } }}
        >
          {rows.map((item: any, i: number) => {
            const key = trackBy ? item[trackBy] : i
            const isSelected = selectedItems?.some((s: any) => trackBy ? s[trackBy] === item[trackBy] : s === item)
            return (
              <motion.div
                key={key}
                variants={{
                  hidden: { opacity: 0, y: 10 },
                  visible: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } }
                }}
                className={cn('rounded-lg border bg-card p-4 cursor-pointer', isSelected && 'ring-2 ring-primary')}
                onClick={() => {
                  if (!selectionType) return
                  const next = isSelected
                    ? (selectedItems ?? []).filter((s: any) => trackBy ? s[trackBy] !== item[trackBy] : s !== item)
                    : selectionType === 'single' ? [item] : [...(selectedItems ?? []), item]
                  onSelectionChange?.({ detail: { selectedItems: next } })
                }}>
                {cardDefinition?.header?.(item) && <div className="font-medium mb-2">{cardDefinition.header(item)}</div>}
                {cardDefinition?.sections?.map((sec: any, si: number) => (
                  <div key={si} className="text-sm">
                    {sec.header && <div className="text-xs font-medium text-muted-foreground mb-0.5">{sec.header}</div>}
                    {sec.content?.(item)}
                  </div>
                ))}
              </motion.div>
            )
          })}
        </motion.div>
      )}
      {pagination && <div className="flex justify-end">{pagination}</div>}
    </div>
  )
}

// ── Additional components ──────────────────────────────────────────────────

export function Grid({ gridDefinition, children }: any) {
  const cols = gridDefinition?.length ?? 2
  const childArray = React.Children.toArray(children)
  return (
    <div className={`grid grid-cols-${Math.min(cols, 4)} gap-4`}>
      {childArray.map((child, i) => (
        <div key={i} className={gridDefinition?.[i]?.colspan ? `col-span-${gridDefinition[i].colspan}` : undefined}>
          {child}
        </div>
      ))}
    </div>
  )
}

export function DatePicker({ value, onChange, placeholder, disabled, openCalendarAriaLabel: _oca }: {
  value?: string
  onChange?: (e: CE<{ value: string }>) => void
  placeholder?: string
  disabled?: boolean
  openCalendarAriaLabel?: (selectedDate: string | null) => string
  [k: string]: any
}) {
  return (
    <input
      type="date"
      value={value ?? ''}
      onChange={(e) => onChange?.({ detail: { value: e.target.value } })}
      placeholder={placeholder}
      disabled={disabled}
      className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
    />
  )
}
