import { useState, useMemo } from 'react'
import { toast } from 'sonner'
import {
  SpaceBetween,
  Header,
  Button,
  Box,
  ColumnLayout,
  Container,
  Select,
  Table,
  Link,
  Badge,
  StatusIndicator,
  ButtonDropdown,
  Spinner,
} from '../lib/cloudscape-shim'
import { useApi } from '../hooks/use-api'
import { CreateBudgetModal } from '../modals/CreateBudgetModal'
import type { Budget } from '../lib/types'

interface BudgetPoolManagementViewProps {
  budgetPools: Budget[]
  loading: boolean
  onRefresh: () => void
}

// Top-level component (not inside PrismApp) to prevent re-mount on state change (#13).
export function BudgetPoolManagementView({ budgetPools, loading, onRefresh }: BudgetPoolManagementViewProps) {
  const api = useApi()
  const [budgetFilter, setBudgetFilter] = useState<string>('all')
  const [createModalVisible, setCreateModalVisible] = useState(false)

  const enrichedBudgets = useMemo(() => {
    return budgetPools.map(budget => {
      const spentPercent = budget.allocated_amount > 0
        ? (budget.spent_amount / budget.allocated_amount) * 100
        : 0
      const status: 'ok' | 'warning' | 'critical' =
        spentPercent >= 95 ? 'critical' :
        spentPercent >= 80 ? 'warning' : 'ok'
      return { ...budget, spentPercent, status }
    })
  }, [budgetPools])

  const filteredBudgets = useMemo(() => {
    if (budgetFilter === 'all') return enrichedBudgets
    if (budgetFilter === 'warning') return enrichedBudgets.filter(b => b.status === 'warning')
    if (budgetFilter === 'critical') return enrichedBudgets.filter(b => b.status === 'critical')
    return enrichedBudgets
  }, [enrichedBudgets, budgetFilter])

  const totalBudgetAmount = enrichedBudgets.reduce((sum, b) => sum + b.total_amount, 0)
  const totalAllocated = enrichedBudgets.reduce((sum, b) => sum + b.allocated_amount, 0)
  const totalSpent = enrichedBudgets.reduce((sum, b) => sum + b.spent_amount, 0)
  const criticalCount = enrichedBudgets.filter(b => b.status === 'critical').length

  const handleDeleteBudget = async (budgetId: string, budgetName: string) => {
    try {
      await api.deleteBudgetPool(budgetId)
      toast.success('Budget Deleted', { description: `Budget pool "${budgetName}" deleted successfully` })
      onRefresh()
    } catch (error) {
      toast.error('Failed to Delete Budget', {
        description: error instanceof Error ? error.message : String(error)
      })
    }
  }

  const handleCreateBudget = async (data: {
    name: string
    description: string
    totalAmount: string
    period: string
    alertThreshold: string
  }) => {
    const budget = await api.createBudgetPool({
      name: data.name,
      description: data.description,
      total_amount: parseFloat(data.totalAmount),
      period: data.period,
      start_date: new Date().toISOString(),
      alert_threshold: parseFloat(data.alertThreshold) / 100,
      created_by: 'current-user',
    })
    toast.success('Budget Created', {
      description: `Budget pool "${budget.name}" created successfully`
    })
    onRefresh()
  }

  return (
    <>
      <SpaceBetween size="l">
        <Header
          variant="h1"
          description="Manage budget pools, project allocations, and spending forecasts"
          counter={`(${budgetPools.length} budget pools)`}
          actions={
            <SpaceBetween direction="horizontal" size="xs">
              <Button onClick={onRefresh} disabled={loading}>
                {loading ? <Spinner /> : 'Refresh'}
              </Button>
              <Button
                variant="primary"
                data-testid="create-budget-button"
                onClick={() => setCreateModalVisible(true)}
              >
                Create Budget Pool
              </Button>
            </SpaceBetween>
          }
        >
          Budget Overview
        </Header>

        <ColumnLayout columns={4} variant="text-grid">
          <Container header={<Header variant="h3">Total Budgets</Header>}>
            <Box fontSize="display-l" fontWeight="bold" color="text-status-info">
              {budgetPools.length}
            </Box>
          </Container>
          <Container header={<Header variant="h3">Total Allocated</Header>}>
            <Box fontSize="display-l" fontWeight="bold" color="text-body-secondary">
              ${totalAllocated.toFixed(2)}
            </Box>
            <Box variant="small" color="text-body-secondary">
              of ${totalBudgetAmount.toFixed(2)} total
            </Box>
          </Container>
          <Container header={<Header variant="h3">Total Spent</Header>}>
            <Box fontSize="display-l" fontWeight="bold" color={totalSpent / totalAllocated > 0.8 ? 'text-status-error' : 'text-status-success'}>
              ${totalSpent.toFixed(2)}
            </Box>
            <Box variant="small" color="text-body-secondary">
              {totalAllocated > 0 ? ((totalSpent / totalAllocated) * 100).toFixed(1) : 0}% of allocated
            </Box>
          </Container>
          <Container header={<Header variant="h3">Active Alerts</Header>}>
            <Box fontSize="display-l" fontWeight="bold" color={criticalCount > 0 ? 'text-status-error' : 'text-status-success'}>
              {criticalCount}
            </Box>
            <Box variant="small" color="text-body-secondary">
              Critical budget alerts
            </Box>
          </Container>
        </ColumnLayout>

        <Container
          header={
            <Header
              variant="h2"
              description="Budget pools with project allocations and spending status"
              counter={`(${filteredBudgets.length})`}
              actions={
                <SpaceBetween direction="horizontal" size="xs">
                  <Select
                    selectedOption={{
                      label: budgetFilter === 'all' ? 'All Budgets' :
                             budgetFilter === 'warning' ? 'Warning (80-95%)' : 'Critical (≥95%)',
                      value: budgetFilter
                    }}
                    onChange={({ detail }) => setBudgetFilter(detail.selectedOption.value!)}
                    options={[
                      { label: 'All Budgets', value: 'all' },
                      { label: 'Warning (80-95%)', value: 'warning' },
                      { label: 'Critical (≥95%)', value: 'critical' }
                    ]}
                    data-testid="budget-filter-select"
                  />
                </SpaceBetween>
              }
            >
              Budget Pools
            </Header>
          }
        >
          <Table
            data-testid="budgets-table"
            columnDefinitions={[
              {
                id: 'name',
                header: 'Budget Name',
                cell: (item: Budget & { spentPercent: number; status: string }) => (
                  <Link fontSize="body-m">
                    {item.name}
                  </Link>
                ),
                sortingField: 'name'
              },
              {
                id: 'total',
                header: 'Total Amount',
                cell: (item: Budget) => `$${item.total_amount.toFixed(2)}`,
                sortingField: 'total_amount'
              },
              {
                id: 'allocated',
                header: 'Allocated',
                cell: (item: Budget) => {
                  const utilization = item.total_amount > 0 ? (item.allocated_amount / item.total_amount) * 100 : 0
                  return (
                    <SpaceBetween direction="horizontal" size="xs">
                      <span>${item.allocated_amount.toFixed(2)}</span>
                      <Badge color={utilization > 90 ? 'red' : utilization > 70 ? 'blue' : 'green'}>
                        {utilization.toFixed(1)}%
                      </Badge>
                    </SpaceBetween>
                  )
                },
                sortingField: 'allocated_amount'
              },
              {
                id: 'spent',
                header: 'Spent',
                cell: (item: Budget & { spentPercent: number; status: string }) => {
                  const colorType = item.status === 'critical' ? 'error' :
                                   item.status === 'warning' ? 'warning' : 'success'
                  return (
                    <SpaceBetween direction="horizontal" size="xs">
                      <StatusIndicator type={colorType}>
                        ${item.spent_amount.toFixed(2)}
                      </StatusIndicator>
                      <Badge color={item.status === 'critical' ? 'red' :
                                    item.status === 'warning' ? 'blue' : 'green'}>
                        {item.spentPercent.toFixed(1)}%
                      </Badge>
                    </SpaceBetween>
                  )
                }
              },
              {
                id: 'remaining',
                header: 'Remaining',
                cell: (item: Budget) => {
                  const remaining = item.allocated_amount - item.spent_amount
                  return `$${remaining.toFixed(2)}`
                }
              },
              {
                id: 'period',
                header: 'Period',
                cell: (item: Budget) => item.period,
                sortingField: 'period'
              },
              {
                id: 'actions',
                header: 'Actions',
                cell: (item: Budget) => (
                  <ButtonDropdown
                    data-testid={`budget-actions-${item.id}`}
                    expandToViewport
                    items={[
                      { text: 'View Summary', id: 'view' },
                      { text: 'Manage Allocations', id: 'allocations' },
                      { text: 'Spending Report', id: 'report' },
                      { text: 'Edit Budget', id: 'edit' },
                      { text: 'Delete', id: 'delete' }
                    ]}
                    onItemClick={({ detail }) => {
                      if (detail.id === 'delete') {
                        handleDeleteBudget(item.id, item.name)
                      } else {
                        const actionNames: Record<string, string> = {
                          view: 'View Summary', allocations: 'Manage Allocations',
                          report: 'Spending Report', edit: 'Edit Budget'
                        }
                        toast.info('Budget Action', {
                          description: `${actionNames[detail.id] ?? detail.id} for budget "${item.name}" - Feature coming soon!`
                        })
                      }
                    }}
                  >
                    Actions
                  </ButtonDropdown>
                )
              }
            ]}
            items={filteredBudgets}
            loadingText="Loading budget pools..."
            empty={
              <Box textAlign="center" color="text-body-secondary">
                <Box variant="strong" textAlign="center" color="text-body-secondary">
                  No budget pools found
                </Box>
                <Box variant="p" padding={{ bottom: 's' }} color="text-body-secondary">
                  Create your first budget pool to track spending across projects.
                </Box>
                <Button variant="primary" onClick={() => setCreateModalVisible(true)}>
                  Create Budget Pool
                </Button>
              </Box>
            }
          />
        </Container>
      </SpaceBetween>

      <CreateBudgetModal
        visible={createModalVisible}
        onDismiss={() => setCreateModalVisible(false)}
        onSubmit={handleCreateBudget}
      />
    </>
  )
}
