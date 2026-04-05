import { useState } from 'react'
import { toast } from 'sonner'
import {
  SpaceBetween,
  Header,
  Container,
  Button,
  Box,
  Table,
  ColumnLayout,
  Link,
  ButtonDropdown,
  Spinner,
  Modal,
  Alert,
  Badge,
  Tabs,
} from '../lib/cloudscape-shim'
import { StatusIndicator } from '../components/status-indicator'
import { useApi } from '../hooks/use-api'
import type { AMI, AMIBuild, AMIRegion } from '../lib/types'

interface AMIManagementViewProps {
  amis: AMI[]
  amiRegions: AMIRegion[]
  amiBuilds: AMIBuild[]
  loading: boolean
  onRefresh: () => Promise<void>
}

export function AMIManagementView({
  amis,
  amiRegions,
  amiBuilds,
  loading,
  onRefresh,
}: AMIManagementViewProps) {
  const api = useApi()
  const [selectedTab, setSelectedTab] = useState<'amis' | 'builds' | 'regions'>('amis')
  const [selectedAMI, setSelectedAMI] = useState<AMI | null>(null)
  const [deleteModalVisible, setDeleteModalVisible] = useState(false)

  const totalSize = amis.reduce((sum, ami) => sum + ami.size_gb, 0)
  const monthlyCost = totalSize * 0.05

  const handleDeleteAMI = async () => {
    if (!selectedAMI) return
    try {
      await api.deleteAMI(selectedAMI.id)
      toast.success(`AMI ${selectedAMI.id} deleted successfully`)
      setDeleteModalVisible(false)
      setSelectedAMI(null)
      await onRefresh()
    } catch (error) {
      toast.error(`Failed to delete AMI: ${error}`)
    }
  }

  return (
    <SpaceBetween size="l">
      <Header
        variant="h1"
        description="Manage AMIs for fast workspace launching (30 seconds vs 5-8 minutes)"
        counter={`(${amis.length} AMIs)`}
        actions={
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={onRefresh} disabled={loading}>
              {loading ? <Spinner /> : 'Refresh'}
            </Button>
          </SpaceBetween>
        }
      >
        AMI Management
      </Header>

      <ColumnLayout columns={4} variant="text-grid">
        <Container header={<Header variant="h3">Total AMIs</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-info">
            {amis.length}
          </Box>
        </Container>
        <Container header={<Header variant="h3">Total Size</Header>}>
          <Box fontSize="display-l" fontWeight="bold">
            {totalSize.toFixed(1)} GB
          </Box>
        </Container>
        <Container header={<Header variant="h3">Monthly Cost</Header>}>
          <Box fontSize="display-l" fontWeight="bold" color="text-status-warning">
            ${monthlyCost.toFixed(2)}
          </Box>
          <Box variant="small" color="text-body-secondary">
            Snapshot storage
          </Box>
        </Container>
        <Container header={<Header variant="h3">Regions</Header>}>
          <Box fontSize="display-l" fontWeight="bold">
            {amiRegions.length}
          </Box>
        </Container>
      </ColumnLayout>

      <Tabs
        activeTabId={selectedTab}
        onChange={({ detail }) => setSelectedTab(detail.activeTabId as 'amis' | 'builds' | 'regions')}
        tabs={[
          {
            id: 'amis',
            label: 'AMIs',
            content: (
              <Container>
                <Table
                  columnDefinitions={[
                    {
                      id: 'id',
                      header: 'AMI ID',
                      cell: (item: AMI) => <Link fontSize="body-m" onFollow={() => setSelectedAMI(item)}>{item.id}</Link>,
                      sortingField: 'id'
                    },
                    {
                      id: 'template',
                      header: 'Template',
                      cell: (item: AMI) => item.template_name,
                      sortingField: 'template_name'
                    },
                    {
                      id: 'region',
                      header: 'Region',
                      cell: (item: AMI) => <Badge>{item.region}</Badge>,
                      sortingField: 'region'
                    },
                    {
                      id: 'state',
                      header: 'State',
                      cell: (item: AMI) => (
                        <StatusIndicator type={item.state === 'available' ? 'success' : 'pending'}>
                          {item.state}
                        </StatusIndicator>
                      )
                    },
                    {
                      id: 'architecture',
                      header: 'Architecture',
                      cell: (item: AMI) => item.architecture
                    },
                    {
                      id: 'size',
                      header: 'Size',
                      cell: (item: AMI) => `${item.size_gb.toFixed(1)} GB`,
                      sortingField: 'size_gb'
                    },
                    {
                      id: 'created',
                      header: 'Created',
                      cell: (item: AMI) => new Date(item.created_at).toLocaleDateString()
                    },
                    {
                      id: 'actions',
                      header: 'Actions',
                      cell: (item: AMI) => (
                        <ButtonDropdown
                          expandToViewport
                          items={[
                            { text: 'View Details', id: 'details' },
                            { text: 'Copy to Region', id: 'copy', disabled: true },
                            { text: 'Delete AMI', id: 'delete' }
                          ]}
                          onItemClick={({ detail }) => {
                            setSelectedAMI(item)
                            if (detail.id === 'delete') {
                              setDeleteModalVisible(true)
                            }
                          }}
                        >
                          Actions
                        </ButtonDropdown>
                      )
                    }
                  ]}
                  items={amis}
                  loadingText="Loading AMIs..."
                  loading={loading}
                  trackBy="id"
                  empty={
                    <Box textAlign="center" color="text-body-secondary">
                      <Box variant="strong" textAlign="center" color="text-body-secondary">
                        No AMIs available
                      </Box>
                      <Box variant="p" padding={{ bottom: 's' }} color="text-body-secondary">
                        Build an AMI to enable fast workspace launching (30 seconds vs 5-8 minutes).
                      </Box>
                    </Box>
                  }
                  sortingDisabled={false}
                />
              </Container>
            )
          },
          {
            id: 'builds',
            label: 'Build Status',
            content: (
              <Container>
                {amiBuilds.length === 0 ? (
                  <Box textAlign="center" padding="xl">
                    <Box variant="strong">No active builds</Box>
                    <Box variant="p" color="text-body-secondary">
                      AMI builds typically take 10-15 minutes to complete.
                    </Box>
                  </Box>
                ) : (
                  <Table
                    columnDefinitions={[
                      { id: 'id', header: 'Build ID', cell: (item: AMIBuild) => item.id },
                      { id: 'template', header: 'Template', cell: (item: AMIBuild) => item.template_name },
                      {
                        id: 'status',
                        header: 'Status',
                        cell: (item: AMIBuild) => (
                          <StatusIndicator
                            type={
                              item.status === 'completed' ? 'success' :
                              item.status === 'failed' ? 'error' : 'in-progress'
                            }
                          >
                            {item.status}
                          </StatusIndicator>
                        )
                      },
                      { id: 'progress', header: 'Progress', cell: (item: AMIBuild) => `${item.progress}%` },
                      { id: 'step', header: 'Current Step', cell: (item: AMIBuild) => item.current_step || '-' }
                    ]}
                    items={amiBuilds}
                    trackBy="id"
                  />
                )}
              </Container>
            )
          },
          {
            id: 'regions',
            label: 'Regional Coverage',
            content: (
              <Container>
                <Table
                  columnDefinitions={[
                    {
                      id: 'region',
                      header: 'Region',
                      cell: (item: AMIRegion) => <Badge color={item.ami_count > 0 ? 'green' : 'grey'}>{item.name}</Badge>,
                      sortingField: 'name'
                    },
                    {
                      id: 'count',
                      header: 'AMI Count',
                      cell: (item: AMIRegion) => item.ami_count,
                      sortingField: 'ami_count'
                    },
                    {
                      id: 'size',
                      header: 'Total Size',
                      cell: (item: AMIRegion) => `${item.total_size_gb.toFixed(1)} GB`,
                      sortingField: 'total_size_gb'
                    },
                    {
                      id: 'cost',
                      header: 'Monthly Cost',
                      cell: (item: AMIRegion) => `$${item.monthly_cost.toFixed(2)}`,
                      sortingField: 'monthly_cost'
                    }
                  ]}
                  items={amiRegions}
                  trackBy="name"
                  sortingDisabled={false}
                  empty={
                    <Box textAlign="center" padding="xl">
                      <Box variant="strong">No regional data available</Box>
                    </Box>
                  }
                />
              </Container>
            )
          }
        ]}
      />

      <Modal
        visible={deleteModalVisible}
        onDismiss={() => setDeleteModalVisible(false)}
        header="Delete AMI"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setDeleteModalVisible(false)}>Cancel</Button>
              <Button variant="primary" onClick={handleDeleteAMI}>Delete</Button>
            </SpaceBetween>
          </Box>
        }
      >
        <SpaceBetween size="m">
          <Alert type="warning">
            This will permanently delete the AMI and associated snapshots. This action cannot be undone.
          </Alert>
          {selectedAMI && (
            <Box>
              <Box variant="strong">AMI ID:</Box> {selectedAMI.id}
              <br />
              <Box variant="strong">Template:</Box> {selectedAMI.template_name}
              <br />
              <Box variant="strong">Size:</Box> {selectedAMI.size_gb.toFixed(1)} GB
            </Box>
          )}
        </SpaceBetween>
      </Modal>
    </SpaceBetween>
  )
}
