import { useState, useMemo } from 'react'
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
  FormField,
  Select,
  TextFilter,
  Toggle,
} from '../lib/cloudscape-shim'
import { StatusIndicator } from '../components/status-indicator'
import { useApi } from '../hooks/use-api'
import type { EFSVolume, EBSVolume, Instance, S3Mount, FileEntry, StorageAnalyticsSummary } from '../lib/types'

interface StorageManagementViewProps {
  efsVolumes: EFSVolume[]
  ebsVolumes: EBSVolume[]
  instances: Instance[]
  loading: boolean
  onRefresh: () => Promise<void>
  onOpenCreateEFS: () => void
  onOpenCreateEBS: () => void
  onDeleteRequest: (volumeName: string, type: 'efs-volume' | 'ebs-volume', onConfirm: () => Promise<void>) => void
}

export function StorageManagementView({
  efsVolumes,
  ebsVolumes,
  instances,
  loading,
  onRefresh,
  onOpenCreateEFS,
  onOpenCreateEBS,
  onDeleteRequest,
}: StorageManagementViewProps) {
  const api = useApi()

  const [activeTabId, setActiveTabId] = useState('shared')
  const [efsFilterText, setEfsFilterText] = useState('')
  const [ebsFilterText, setEbsFilterText] = useState('')

  const [attachModalVisible, setAttachModalVisible] = useState(false)
  const [attachModalVolume, setAttachModalVolume] = useState<EBSVolume | null>(null)
  const [selectedAttachInstance, setSelectedAttachInstance] = useState('')

  const [mountModalVisible, setMountModalVisible] = useState(false)
  const [mountModalVolume, setMountModalVolume] = useState<EFSVolume | null>(null)
  const [selectedMountInstance, setSelectedMountInstance] = useState('')

  const [unmountModalVisible, setUnmountModalVisible] = useState(false)
  const [unmountModalVolume, setUnmountModalVolume] = useState<EFSVolume | null>(null)

  const [detachModalVisible, setDetachModalVisible] = useState(false)
  const [detachModalVolume, setDetachModalVolume] = useState<EBSVolume | null>(null)

  const [mountS3ModalVisible, setMountS3ModalVisible] = useState(false)
  const [mountS3Bucket, setMountS3Bucket] = useState('')
  const [mountS3Path, setMountS3Path] = useState('')
  const [mountS3Method, setMountS3Method] = useState('mountpoint')
  const [mountS3ReadOnly, setMountS3ReadOnly] = useState(false)

  const [s3MountsInstance, setS3MountsInstance] = useState('')
  const [s3Mounts, setS3Mounts] = useState<S3Mount[]>([])
  const [s3MountLoading, setS3MountLoading] = useState(false)

  const [pushFileModalVisible, setPushFileModalVisible] = useState(false)
  const [pushFileLocal, setPushFileLocal] = useState('')
  const [pushFileRemote, setPushFileRemote] = useState('')

  const [pullFileModalVisible, setPullFileModalVisible] = useState(false)
  const [pullFileRemote, setPullFileRemote] = useState('')
  const [pullFileLocal, setPullFileLocal] = useState('')

  const [filesInstance, setFilesInstance] = useState('')
  const [filesPath, setFilesPath] = useState('/home')
  const [fileEntries, setFileEntries] = useState<FileEntry[]>([])
  const [filesLoading, setFilesLoading] = useState(false)

  const [analyticsPeriod, setAnalyticsPeriod] = useState('daily')
  const [analyticsData, setAnalyticsData] = useState<StorageAnalyticsSummary[]>([])
  const [analyticsLoading, setAnalyticsLoading] = useState(false)

  const filteredEFSVolumes = useMemo(() => {
    if (!efsFilterText) return efsVolumes
    const q = efsFilterText.toLowerCase()
    return efsVolumes.filter(v => v.name.toLowerCase().includes(q) || v.filesystem_id.toLowerCase().includes(q))
  }, [efsVolumes, efsFilterText])

  const filteredEBSVolumes = useMemo(() => {
    if (!ebsFilterText) return ebsVolumes
    const q = ebsFilterText.toLowerCase()
    return ebsVolumes.filter(v => v.name.toLowerCase().includes(q) || v.volume_id.toLowerCase().includes(q))
  }, [ebsVolumes, ebsFilterText])

  const instanceOptions = instances.map(i => ({ label: i.name, value: i.name }))

  async function handleAttachEBS() {
    if (!attachModalVolume || !selectedAttachInstance) return
    toast.info(`Attaching volume ${attachModalVolume.name} to ${selectedAttachInstance}...`)
    setAttachModalVisible(false)
    try {
      await api.attachEBSVolume(attachModalVolume.name, selectedAttachInstance)
      toast.success(`Volume ${attachModalVolume.name} attached to ${selectedAttachInstance}`)
      await onRefresh()
    } catch (e: unknown) {
      toast.error(`Failed to attach volume: ${e instanceof Error ? e.message : String(e)}`)
    }
  }

  async function handleDetachEBS() {
    if (!detachModalVolume) return
    toast.info(`Detaching volume ${detachModalVolume.name}...`)
    setDetachModalVisible(false)
    try {
      await api.detachEBSVolume(detachModalVolume.name)
      toast.success(`Volume ${detachModalVolume.name} detached`)
      await onRefresh()
    } catch (e: unknown) {
      toast.error(`Failed to detach volume: ${e instanceof Error ? e.message : String(e)}`)
    }
  }

  async function handleMountEFS() {
    if (!mountModalVolume || !selectedMountInstance) return
    toast.info(`Mounting volume ${mountModalVolume.name} to ${selectedMountInstance}...`)
    setMountModalVisible(false)
    try {
      await api.mountEFSVolume(mountModalVolume.name, selectedMountInstance)
      toast.success(`Volume ${mountModalVolume.name} mounted to ${selectedMountInstance}`)
      await onRefresh()
    } catch (e: unknown) {
      toast.error(`Failed to mount volume: ${e instanceof Error ? e.message : String(e)}`)
    }
  }

  async function handleUnmountEFS() {
    if (!unmountModalVolume) return
    const volumeName = unmountModalVolume.name
    toast.info(`Unmounting volume ${volumeName}...`)
    setUnmountModalVisible(false)
    try {
      await api.unmountEFSVolume(volumeName, instances[0].name)
      toast.success(`Volume ${volumeName} unmounted`)
      await onRefresh()
    } catch (e: unknown) {
      toast.error(`Failed to unmount volume: ${e instanceof Error ? e.message : String(e)}`)
    }
  }

  async function handleMountS3() {
    if (!mountS3Bucket || !s3MountsInstance) return
    toast.info(`Mounting S3 bucket ${mountS3Bucket}...`)
    setMountS3ModalVisible(false)
    try {
      await api.mountS3Bucket(s3MountsInstance, mountS3Bucket, mountS3Path, mountS3Method, mountS3ReadOnly)
      toast.success(`S3 bucket ${mountS3Bucket} mounted`)
      await loadS3Mounts(s3MountsInstance)
    } catch (e: unknown) {
      toast.error(`Failed to mount S3 bucket: ${e instanceof Error ? e.message : String(e)}`)
    }
  }

  async function handleUnmountS3(bucket: string) {
    if (!s3MountsInstance) return
    toast.info(`Unmounting S3 bucket ${bucket}...`)
    try {
      await api.unmountS3Bucket(s3MountsInstance, bucket)
      toast.success(`S3 bucket ${bucket} unmounted`)
      await loadS3Mounts(s3MountsInstance)
    } catch (e: unknown) {
      toast.error(`Failed to unmount S3 bucket: ${e instanceof Error ? e.message : String(e)}`)
    }
  }

  async function loadS3Mounts(instanceName: string) {
    if (!instanceName) return
    setS3MountLoading(true)
    try {
      const mounts = await api.listInstanceS3Mounts(instanceName)
      setS3Mounts(mounts || [])
    } catch (e: unknown) {
      toast.error(`Failed to load S3 mounts: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setS3MountLoading(false)
    }
  }

  async function loadFiles(instanceName: string, path: string) {
    if (!instanceName) return
    setFilesLoading(true)
    try {
      const entries = await api.listInstanceFiles(instanceName, path)
      setFileEntries(entries || [])
    } catch (e: unknown) {
      toast.error(`Failed to list files: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setFilesLoading(false)
    }
  }

  async function handlePushFile() {
    if (!filesInstance || !pushFileLocal || !pushFileRemote) return
    toast.info(`Pushing file to ${filesInstance}...`)
    setPushFileModalVisible(false)
    try {
      await api.pushFileToInstance(filesInstance, pushFileLocal, pushFileRemote)
      toast.success('File pushed successfully')
      await loadFiles(filesInstance, filesPath)
    } catch (e: unknown) {
      toast.error(`Failed to push file: ${e instanceof Error ? e.message : String(e)}`)
    }
  }

  async function handlePullFile() {
    if (!filesInstance || !pullFileRemote || !pullFileLocal) return
    toast.info(`Pulling file from ${filesInstance}...`)
    setPullFileModalVisible(false)
    try {
      await api.pullFileFromInstance(filesInstance, pullFileRemote, pullFileLocal)
      toast.success('File pulled successfully')
    } catch (e: unknown) {
      toast.error(`Failed to pull file: ${e instanceof Error ? e.message : String(e)}`)
    }
  }

  async function loadAnalytics() {
    setAnalyticsLoading(true)
    try {
      const data = await api.getAllStorageAnalytics(analyticsPeriod)
      setAnalyticsData(data || [])
    } catch (e: unknown) {
      toast.error(`Failed to load analytics: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setAnalyticsLoading(false)
    }
  }

  const totalEFSSize = efsVolumes.reduce((acc, v) => acc + Math.round((v.size_bytes || 0) / 1024 / 1024 / 1024), 0)
  const totalEBSSize = ebsVolumes.reduce((acc, v) => acc + (v.size_gb || 0), 0)
  const mountedEFSCount = efsVolumes.filter(v => v.state === 'available').length
  const attachedEBSCount = ebsVolumes.filter(v => v.state === 'in-use').length

  return (
    <SpaceBetween size="l" data-testid="storage-page">
      <Header
        variant="h1"
        actions={
          <SpaceBetween direction="horizontal" size="xs">
            <Button
              onClick={onRefresh}
              loading={loading}
              iconName="refresh"
              data-testid="refresh-storage-button"
            >
              Refresh
            </Button>
            <Button onClick={onOpenCreateEFS} variant="normal" data-testid="create-efs-header-button">
              Create Shared Volume (EFS)
            </Button>
            <Button onClick={onOpenCreateEBS} variant="primary" data-testid="create-ebs-button">
              Create Private Volume (EBS)
            </Button>
          </SpaceBetween>
        }
      >
        Storage Management
      </Header>

      <Container
        header={<Header variant="h2">Storage Overview</Header>}
      >
        <SpaceBetween size="m">
          <Box>
            Prism supports three types of storage for your research workstations: shared EFS volumes (accessible from multiple instances), private EBS volumes (attached to a single instance), and S3 bucket mounts for object storage access.
          </Box>
          <Alert type="info" header="Storage Selection Guide">
            <SpaceBetween size="xs">
              <Box><strong>Shared Storage (EFS)</strong>: Use for datasets and results you need to access from multiple workstations simultaneously. EFS volumes persist independently of instances.</Box>
              <Box><strong>Private Storage (EBS)</strong>: Use for high-performance local storage attached to a single instance. EBS volumes can be detached and reattached to different instances.</Box>
              <Box><strong>S3 Mounts</strong>: Use for accessing large object storage datasets without copying files. Mount S3 buckets directly as a filesystem.</Box>
            </SpaceBetween>
          </Alert>
        </SpaceBetween>
      </Container>

      <Tabs
        activeTabId={activeTabId}
        onChange={({ detail }: { detail: { activeTabId: string } }) => setActiveTabId(detail.activeTabId)}
        tabs={[
          {
            id: 'shared',
            label: `Shared Volumes (EFS) (${efsVolumes.length})`,
            content: (
              <SpaceBetween size="m">
                <Header variant="h3">Shared Storage Volumes</Header>
                <TextFilter
                  filteringText={efsFilterText}
                  filteringPlaceholder="Filter EFS volumes by name or ID"
                  onChange={({ detail }: { detail: { filteringText: string } }) => setEfsFilterText(detail.filteringText)}
                  data-testid="efs-filter"
                />
                <Table
                  data-testid="efs-table"
                  loading={loading}
                  loadingText="Loading EFS volumes..."
                  items={filteredEFSVolumes}
                  empty={
                    <Box textAlign="center" color="inherit">
                      <SpaceBetween size="m">
                        <b>No shared storage volumes found</b>
                        <Box>Create a shared EFS volume to get started.</Box>
                        <Button onClick={onOpenCreateEFS}>Create Shared Volume</Button>
                      </SpaceBetween>
                    </Box>
                  }
                  columnDefinitions={[
                    {
                      id: 'name',
                      header: 'Name',
                      cell: (item: EFSVolume) => <Box fontWeight="bold">{item.name}</Box>,
                    },
                    {
                      id: 'filesystem_id',
                      header: 'Filesystem ID',
                      cell: (item: EFSVolume) => <Box variant="code">{item.filesystem_id}</Box>,
                    },
                    {
                      id: 'status',
                      header: 'Status',
                      cell: (item: EFSVolume) => (
                        <StatusIndicator type={item.state === 'available' ? 'success' : item.state === 'creating' ? 'pending' : 'error'}>
                          {item.state}
                        </StatusIndicator>
                      ),
                    },
                    {
                      id: 'size',
                      header: 'Size',
                      cell: (item: EFSVolume) => item.size_bytes ? `${Math.round(item.size_bytes / 1024 / 1024 / 1024)} GB` : 'Auto',
                    },
                    {
                      id: 'attached_to',
                      header: 'Attached To',
                      cell: (item: EFSVolume) => item.attached_to ? <Box>{item.attached_to}</Box> : '-',
                    },
                    {
                      id: 'actions',
                      header: 'Actions',
                      cell: (item: EFSVolume) => (
                        <ButtonDropdown
                          items={[
                            { id: 'mount', text: 'Mount to Instance' },
                            { id: 'unmount', text: 'Unmount' },
                            { id: 'delete', text: 'Delete' },
                          ]}
                          onItemClick={({ detail }: { detail: { id: string } }) => {
                            if (detail.id === 'mount') {
                              setMountModalVolume(item)
                              setSelectedMountInstance('')
                              setMountModalVisible(true)
                            } else if (detail.id === 'unmount') {
                              setUnmountModalVolume(item)
                              setUnmountModalVisible(true)
                            } else if (detail.id === 'delete') {
                              onDeleteRequest(item.name, 'efs-volume', async () => {
                                await api.deleteEFSVolume(item.name)
                                await onRefresh()
                              })
                            }
                          }}
                          data-testid={`efs-actions-${item.name}`}
                        >
                          Actions
                        </ButtonDropdown>
                      ),
                    },
                  ]}
                />
              </SpaceBetween>
            ),
          },
          {
            id: 'private',
            label: `Private Volumes (EBS) (${ebsVolumes.length})`,
            content: (
              <SpaceBetween size="m">
                <TextFilter
                  filteringText={ebsFilterText}
                  filteringPlaceholder="Filter EBS volumes by name or ID"
                  onChange={({ detail }: { detail: { filteringText: string } }) => setEbsFilterText(detail.filteringText)}
                  data-testid="ebs-filter"
                />
                <Table
                  data-testid="ebs-table"
                  loading={loading}
                  loadingText="Loading EBS volumes..."
                  items={filteredEBSVolumes}
                  empty={
                    <Box textAlign="center" color="inherit">
                      <SpaceBetween size="m">
                        <b>No private volumes</b>
                        <Box>Create a private EBS volume to get started.</Box>
                        <Button onClick={onOpenCreateEBS}>Create Private Volume</Button>
                      </SpaceBetween>
                    </Box>
                  }
                  columnDefinitions={[
                    {
                      id: 'name',
                      header: 'Name',
                      cell: (item: EBSVolume) => <Box fontWeight="bold">{item.name}</Box>,
                    },
                    {
                      id: 'volume_id',
                      header: 'Volume ID',
                      cell: (item: EBSVolume) => <Box variant="code">{item.volume_id}</Box>,
                    },
                    {
                      id: 'status',
                      header: 'Status',
                      cell: (item: EBSVolume) => (
                        <StatusIndicator type={item.state === 'available' ? 'success' : item.state === 'in-use' ? 'success' : item.state === 'creating' ? 'pending' : 'error'}>
                          {item.state}
                        </StatusIndicator>
                      ),
                    },
                    {
                      id: 'size',
                      header: 'Size',
                      cell: (item: EBSVolume) => `${item.size_gb} GB`,
                    },
                    {
                      id: 'volume_type',
                      header: 'Type',
                      cell: (item: EBSVolume) => item.volume_type || 'gp3',
                    },
                    {
                      id: 'attached_to',
                      header: 'Attached To',
                      cell: (item: EBSVolume) => item.attached_to || '-',
                    },
                    {
                      id: 'actions',
                      header: 'Actions',
                      cell: (item: EBSVolume) => (
                        <ButtonDropdown
                          items={[
                            { id: 'attach', text: 'Attach to Instance' },
                            { id: 'detach', text: 'Detach' },
                            { id: 'delete', text: 'Delete' },
                          ]}
                          onItemClick={({ detail }: { detail: { id: string } }) => {
                            if (detail.id === 'attach') {
                              setAttachModalVolume(item)
                              setSelectedAttachInstance('')
                              setAttachModalVisible(true)
                            } else if (detail.id === 'detach') {
                              setDetachModalVolume(item)
                              setDetachModalVisible(true)
                            } else if (detail.id === 'delete') {
                              onDeleteRequest(item.name, 'ebs-volume', async () => {
                                await api.deleteEBSVolume(item.name)
                                await onRefresh()
                              })
                            }
                          }}
                          data-testid={`ebs-actions-${item.name}`}
                        >
                          Actions
                        </ButtonDropdown>
                      ),
                    },
                  ]}
                />
              </SpaceBetween>
            ),
          },
          {
            id: 's3-mounts',
            label: 'S3 Mounts',
            content: (
              <SpaceBetween size="m">
                <Container header={<Header variant="h3" actions={<Button onClick={() => { setMountS3Bucket(''); setMountS3Path(''); setMountS3Method('mountpoint'); setMountS3ReadOnly(false); setMountS3ModalVisible(true) }} data-testid="mount-s3-button">Mount S3 Bucket</Button>}>S3 Bucket Mounts</Header>}>
                  <SpaceBetween size="m">
                    <FormField label="Select Instance">
                      <Select
                        selectedOption={s3MountsInstance ? { label: s3MountsInstance, value: s3MountsInstance } : null}
                        onChange={({ detail }: { detail: { selectedOption: { value?: string } } }) => {
                          const name = detail.selectedOption.value || ''
                          setS3MountsInstance(name)
                          if (name) loadS3Mounts(name)
                        }}
                        options={instanceOptions}
                        placeholder="Select an instance to view S3 mounts"
                        data-testid="s3-instance-select"
                      />
                    </FormField>
                    {s3MountLoading ? (
                      <Spinner />
                    ) : s3MountsInstance ? (
                      <Table
                        data-testid="s3-mounts-table"
                        items={s3Mounts}
                        empty={
                          <Box textAlign="center">
                            <SpaceBetween size="m">
                              <b>No S3 mounts</b>
                              <Box>Mount an S3 bucket to access object storage as a filesystem.</Box>
                            </SpaceBetween>
                          </Box>
                        }
                        columnDefinitions={[
                          {
                            id: 'bucket',
                            header: 'Bucket',
                            cell: (item: S3Mount) => item.bucket_name,
                          },
                          {
                            id: 'mount_path',
                            header: 'Mount Path',
                            cell: (item: S3Mount) => <Box variant="code">{item.mount_path}</Box>,
                          },
                          {
                            id: 'method',
                            header: 'Method',
                            cell: (item: S3Mount) => item.method || 'mountpoint',
                          },
                          {
                            id: 'read_only',
                            header: 'Read Only',
                            cell: (item: S3Mount) => item.read_only ? <Badge color="blue">Read Only</Badge> : <Badge color="green">Read/Write</Badge>,
                          },
                          {
                            id: 'actions',
                            header: 'Actions',
                            cell: (item: S3Mount) => (
                              <Button
                                variant="normal"
                                onClick={() => handleUnmountS3(item.bucket_name)}
                                data-testid={`unmount-s3-${item.bucket_name}`}
                              >
                                Unmount
                              </Button>
                            ),
                          },
                        ]}
                      />
                    ) : null}
                  </SpaceBetween>
                </Container>
              </SpaceBetween>
            ),
          },
          {
            id: 'files',
            label: 'File Transfer',
            content: (
              <SpaceBetween size="m">
                <Container
                  header={
                    <Header
                      variant="h3"
                      actions={
                        <SpaceBetween direction="horizontal" size="xs">
                          <Button
                            onClick={() => { setPushFileLocal(''); setPushFileRemote(''); setPushFileModalVisible(true) }}
                            data-testid="push-file-button"
                          >
                            Push File
                          </Button>
                          <Button
                            onClick={() => { setPullFileRemote(''); setPullFileLocal(''); setPullFileModalVisible(true) }}
                            data-testid="pull-file-button"
                          >
                            Pull File
                          </Button>
                          <Button
                            onClick={() => filesInstance && loadFiles(filesInstance, filesPath)}
                            iconName="refresh"
                            data-testid="refresh-files-button"
                          >
                            Refresh
                          </Button>
                        </SpaceBetween>
                      }
                    >
                      File Browser
                    </Header>
                  }
                >
                  <SpaceBetween size="m">
                    <ColumnLayout columns={2}>
                      <FormField label="Select Instance">
                        <Select
                          selectedOption={filesInstance ? { label: filesInstance, value: filesInstance } : null}
                          onChange={({ detail }: { detail: { selectedOption: { value?: string } } }) => {
                            const name = detail.selectedOption.value || ''
                            setFilesInstance(name)
                            if (name) loadFiles(name, filesPath)
                          }}
                          options={instanceOptions}
                          placeholder="Select an instance"
                          data-testid="files-instance-select"
                        />
                      </FormField>
                      <FormField label="Path">
                        <SpaceBetween direction="horizontal" size="xs">
                          <input
                            type="text"
                            value={filesPath}
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFilesPath(e.target.value)}
                            onKeyDown={(e: React.KeyboardEvent<HTMLInputElement>) => {
                              if (e.key === 'Enter' && filesInstance) loadFiles(filesInstance, filesPath)
                            }}
                            placeholder="/home"
                            style={{ padding: '6px 8px', border: '1px solid #aab7b8', borderRadius: '2px', fontSize: '14px', width: '200px' }}
                            data-testid="files-path-input"
                          />
                          <Button onClick={() => filesInstance && loadFiles(filesInstance, filesPath)}>Browse</Button>
                        </SpaceBetween>
                      </FormField>
                    </ColumnLayout>
                    {filesLoading ? (
                      <Spinner />
                    ) : filesInstance ? (
                      <Table
                        data-testid="files-table"
                        items={fileEntries}
                        empty={
                          <Box textAlign="center">
                            <b>No files found</b>
                          </Box>
                        }
                        columnDefinitions={[
                          {
                            id: 'name',
                            header: 'Name',
                            cell: (item: FileEntry) => {
                              const baseName = item.path.split('/').pop() || item.path
                              return item.is_dir ? (
                                <Link
                                  onFollow={() => {
                                    const newPath = item.path
                                    setFilesPath(newPath)
                                    loadFiles(filesInstance, newPath)
                                  }}
                                  data-testid={`file-entry-${baseName}`}
                                >
                                  📁 {baseName}
                                </Link>
                              ) : (
                                <Box>{baseName}</Box>
                              )
                            },
                          },
                          {
                            id: 'size',
                            header: 'Size',
                            cell: (item: FileEntry) => item.is_dir ? '-' : item.size_bytes ? `${item.size_bytes} bytes` : '-',
                          },
                          {
                            id: 'modified',
                            header: 'Modified',
                            cell: (item: FileEntry) => item.modified_at || '-',
                          },
                          {
                            id: 'permissions',
                            header: 'Permissions',
                            cell: (item: FileEntry) => item.permissions ? <Box variant="code">{item.permissions}</Box> : '-',
                          },
                        ]}
                      />
                    ) : null}
                  </SpaceBetween>
                </Container>
              </SpaceBetween>
            ),
          },
          {
            id: 'analytics',
            label: 'Analytics',
            content: (
              <SpaceBetween size="m">
                <Container
                  header={
                    <Header
                      variant="h3"
                      actions={
                        <SpaceBetween direction="horizontal" size="xs">
                          <Select
                            selectedOption={{ label: analyticsPeriod === 'daily' ? 'Daily' : analyticsPeriod === 'weekly' ? 'Weekly' : 'Monthly', value: analyticsPeriod }}
                            onChange={({ detail }: { detail: { selectedOption: { value?: string } } }) => {
                              setAnalyticsPeriod(detail.selectedOption.value || 'daily')
                            }}
                            options={[
                              { label: 'Daily', value: 'daily' },
                              { label: 'Weekly', value: 'weekly' },
                              { label: 'Monthly', value: 'monthly' },
                            ]}
                            data-testid="analytics-period-select"
                          />
                          <Button onClick={loadAnalytics} loading={analyticsLoading} data-testid="load-analytics-button">
                            Load Analytics
                          </Button>
                        </SpaceBetween>
                      }
                    >
                      Storage Analytics
                    </Header>
                  }
                >
                  {analyticsLoading ? (
                    <Spinner />
                  ) : analyticsData.length > 0 ? (
                    <Table
                      data-testid="analytics-table"
                      items={analyticsData}
                      columnDefinitions={[
                        {
                          id: 'storage_name',
                          header: 'Volume',
                          cell: (item: StorageAnalyticsSummary) => item.storage_name,
                        },
                        {
                          id: 'type',
                          header: 'Type',
                          cell: (item: StorageAnalyticsSummary) => item.type,
                        },
                        {
                          id: 'usage_percent',
                          header: 'Usage',
                          cell: (item: StorageAnalyticsSummary) => `${item.usage_percent.toFixed(1)}%`,
                        },
                        {
                          id: 'daily_cost',
                          header: 'Daily Cost',
                          cell: (item: StorageAnalyticsSummary) => item.daily_cost ? `$${item.daily_cost.toFixed(4)}` : '-',
                        },
                        {
                          id: 'total_cost',
                          header: 'Total Cost',
                          cell: (item: StorageAnalyticsSummary) => item.total_cost ? `$${item.total_cost.toFixed(2)}` : '-',
                        },
                        {
                          id: 'period',
                          header: 'Period',
                          cell: (item: StorageAnalyticsSummary) => item.period || '-',
                        },
                      ]}
                    />
                  ) : (
                    <Box textAlign="center" color="inherit">
                      <SpaceBetween size="m">
                        <b>No analytics data</b>
                        <Box>Click &quot;Load Analytics&quot; to fetch storage usage data.</Box>
                      </SpaceBetween>
                    </Box>
                  )}
                </Container>
              </SpaceBetween>
            ),
          },
        ]}
      />

      {/* Mount S3 Modal */}
      <Modal
        visible={mountS3ModalVisible}
        onDismiss={() => setMountS3ModalVisible(false)}
        header="Mount S3 Bucket"
        footer={
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={() => setMountS3ModalVisible(false)}>Cancel</Button>
            <Button variant="primary" onClick={handleMountS3} disabled={!mountS3Bucket || !s3MountsInstance} data-testid="confirm-mount-s3-button">Mount</Button>
          </SpaceBetween>
        }
        data-testid="mount-s3-modal"
      >
        <SpaceBetween size="m">
          <FormField label="S3 Bucket Name" description="Enter the name of the S3 bucket to mount">
            <input
              type="text"
              value={mountS3Bucket}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setMountS3Bucket(e.target.value)}
              placeholder="my-bucket"
              style={{ padding: '6px 8px', border: '1px solid #aab7b8', borderRadius: '2px', fontSize: '14px', width: '100%', boxSizing: 'border-box' }}
              data-testid="s3-bucket-input"
            />
          </FormField>
          <FormField label="Mount Path" description="Path on the instance where the bucket will be mounted">
            <input
              type="text"
              value={mountS3Path}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setMountS3Path(e.target.value)}
              placeholder="/mnt/s3/my-bucket"
              style={{ padding: '6px 8px', border: '1px solid #aab7b8', borderRadius: '2px', fontSize: '14px', width: '100%', boxSizing: 'border-box' }}
              data-testid="s3-path-input"
            />
          </FormField>
          <FormField label="Mount Method">
            <Select
              selectedOption={{ label: mountS3Method === 'mountpoint' ? 'Mountpoint (AWS)' : 'S3FS', value: mountS3Method }}
              onChange={({ detail }: { detail: { selectedOption: { value?: string } } }) => setMountS3Method(detail.selectedOption.value || 'mountpoint')}
              options={[
                { label: 'Mountpoint (AWS)', value: 'mountpoint' },
                { label: 'S3FS', value: 's3fs' },
              ]}
              data-testid="s3-method-select"
            />
          </FormField>
          <FormField label="Read Only">
            <Toggle
              checked={mountS3ReadOnly}
              onChange={({ detail }: { detail: { checked: boolean } }) => setMountS3ReadOnly(detail.checked)}
              data-testid="s3-readonly-toggle"
            >
              Mount as read-only
            </Toggle>
          </FormField>
        </SpaceBetween>
      </Modal>

      {/* Push File Modal */}
      <Modal
        visible={pushFileModalVisible}
        onDismiss={() => setPushFileModalVisible(false)}
        header="Push File to Instance"
        footer={
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={() => setPushFileModalVisible(false)}>Cancel</Button>
            <Button variant="primary" onClick={handlePushFile} disabled={!filesInstance || !pushFileLocal || !pushFileRemote} data-testid="confirm-push-file-button">Push</Button>
          </SpaceBetween>
        }
        data-testid="push-file-modal"
      >
        <SpaceBetween size="m">
          <FormField label="Local File Path" description="Path to the file on your local machine">
            <input
              type="text"
              value={pushFileLocal}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPushFileLocal(e.target.value)}
              placeholder="/path/to/local/file"
              style={{ padding: '6px 8px', border: '1px solid #aab7b8', borderRadius: '2px', fontSize: '14px', width: '100%', boxSizing: 'border-box' }}
              data-testid="push-local-input"
            />
          </FormField>
          <FormField label="Remote Path" description="Destination path on the instance">
            <input
              type="text"
              value={pushFileRemote}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPushFileRemote(e.target.value)}
              placeholder="/home/user/file"
              style={{ padding: '6px 8px', border: '1px solid #aab7b8', borderRadius: '2px', fontSize: '14px', width: '100%', boxSizing: 'border-box' }}
              data-testid="push-remote-input"
            />
          </FormField>
        </SpaceBetween>
      </Modal>

      {/* Pull File Modal */}
      <Modal
        visible={pullFileModalVisible}
        onDismiss={() => setPullFileModalVisible(false)}
        header="Pull File from Instance"
        footer={
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={() => setPullFileModalVisible(false)}>Cancel</Button>
            <Button variant="primary" onClick={handlePullFile} disabled={!filesInstance || !pullFileRemote || !pullFileLocal} data-testid="confirm-pull-file-button">Pull</Button>
          </SpaceBetween>
        }
        data-testid="pull-file-modal"
      >
        <SpaceBetween size="m">
          <FormField label="Remote File Path" description="Path to the file on the instance">
            <input
              type="text"
              value={pullFileRemote}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPullFileRemote(e.target.value)}
              placeholder="/home/user/file"
              style={{ padding: '6px 8px', border: '1px solid #aab7b8', borderRadius: '2px', fontSize: '14px', width: '100%', boxSizing: 'border-box' }}
              data-testid="pull-remote-input"
            />
          </FormField>
          <FormField label="Local Path" description="Destination path on your local machine">
            <input
              type="text"
              value={pullFileLocal}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPullFileLocal(e.target.value)}
              placeholder="/path/to/local/destination"
              style={{ padding: '6px 8px', border: '1px solid #aab7b8', borderRadius: '2px', fontSize: '14px', width: '100%', boxSizing: 'border-box' }}
              data-testid="pull-local-input"
            />
          </FormField>
        </SpaceBetween>
      </Modal>

      {/* Attach EBS Modal */}
      <Modal
        visible={attachModalVisible}
        onDismiss={() => setAttachModalVisible(false)}
        header={`Attach Volume: ${attachModalVolume?.name}`}
        footer={
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={() => setAttachModalVisible(false)}>Cancel</Button>
            <Button variant="primary" onClick={handleAttachEBS} disabled={!selectedAttachInstance} data-testid="confirm-attach-button">Attach</Button>
          </SpaceBetween>
        }
        data-testid="attach-ebs-modal"
      >
        <SpaceBetween size="m">
          <Box>Select an instance to attach the volume <strong>{attachModalVolume?.name}</strong> to.</Box>
          <FormField label="Target Instance">
            <Select
              selectedOption={selectedAttachInstance ? { label: selectedAttachInstance, value: selectedAttachInstance } : null}
              onChange={({ detail }: { detail: { selectedOption: { value?: string } } }) => setSelectedAttachInstance(detail.selectedOption.value || '')}
              options={instanceOptions}
              placeholder="Select an instance"
              data-testid="attach-instance-select"
            />
          </FormField>
        </SpaceBetween>
      </Modal>

      {/* Mount EFS Modal */}
      <Modal
        visible={mountModalVisible}
        onDismiss={() => setMountModalVisible(false)}
        header={`Mount Volume: ${mountModalVolume?.name}`}
        footer={
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={() => setMountModalVisible(false)}>Cancel</Button>
            <Button variant="primary" onClick={handleMountEFS} disabled={!selectedMountInstance} data-testid="confirm-mount-button">Mount</Button>
          </SpaceBetween>
        }
        data-testid="mount-efs-modal"
      >
        <SpaceBetween size="m">
          <Box>Select an instance to mount the EFS volume <strong>{mountModalVolume?.name}</strong> to.</Box>
          <FormField label="Target Instance">
            <Select
              selectedOption={selectedMountInstance ? { label: selectedMountInstance, value: selectedMountInstance } : null}
              onChange={({ detail }: { detail: { selectedOption: { value?: string } } }) => setSelectedMountInstance(detail.selectedOption.value || '')}
              options={instanceOptions}
              placeholder="Select an instance"
              data-testid="mount-instance-select"
            />
          </FormField>
        </SpaceBetween>
      </Modal>

      {/* Unmount EFS Modal */}
      <Modal
        visible={unmountModalVisible}
        onDismiss={() => setUnmountModalVisible(false)}
        header={`Unmount Volume: ${unmountModalVolume?.name}`}
        footer={
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={() => setUnmountModalVisible(false)}>Cancel</Button>
            <Button variant="primary" onClick={handleUnmountEFS} data-testid="confirm-unmount-button">Unmount</Button>
          </SpaceBetween>
        }
        data-testid="unmount-efs-modal"
      >
        <SpaceBetween size="m">
          <Alert type="warning">
            Unmounting the volume will make it unavailable to the currently running instance. The data will not be lost.
          </Alert>
          <Box>Are you sure you want to unmount <strong>{unmountModalVolume?.name}</strong>?</Box>
        </SpaceBetween>
      </Modal>

      {/* Detach EBS Modal */}
      <Modal
        visible={detachModalVisible}
        onDismiss={() => setDetachModalVisible(false)}
        header={`Detach Volume: ${detachModalVolume?.name}`}
        footer={
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={() => setDetachModalVisible(false)}>Cancel</Button>
            <Button variant="primary" onClick={handleDetachEBS} data-testid="confirm-detach-button">Detach</Button>
          </SpaceBetween>
        }
        data-testid="detach-ebs-modal"
      >
        <SpaceBetween size="m">
          <Alert type="warning">
            Detaching the volume will make it unavailable to the currently running instance. The data will not be lost.
          </Alert>
          <Box>Are you sure you want to detach <strong>{detachModalVolume?.name}</strong>?</Box>
        </SpaceBetween>
      </Modal>

      {/* Storage Statistics */}
      <Container header={<Header variant="h2">Storage Statistics</Header>}>
        <ColumnLayout columns={4} variant="text-grid">
          <SpaceBetween size="xs">
            <Box variant="awsui-key-label">EFS Volumes</Box>
            <Box variant="awsui-value-large">{efsVolumes.length}</Box>
            <Box color="text-body-secondary">{mountedEFSCount} available/mounted</Box>
          </SpaceBetween>
          <SpaceBetween size="xs">
            <Box variant="awsui-key-label">EBS Volumes</Box>
            <Box variant="awsui-value-large">{ebsVolumes.length}</Box>
            <Box color="text-body-secondary">{attachedEBSCount} in use</Box>
          </SpaceBetween>
          <SpaceBetween size="xs">
            <Box variant="awsui-key-label">Total EFS Storage</Box>
            <Box variant="awsui-value-large">{totalEFSSize} GB</Box>
            <Box color="text-body-secondary">Shared storage</Box>
          </SpaceBetween>
          <SpaceBetween size="xs">
            <Box variant="awsui-key-label">Total EBS Storage</Box>
            <Box variant="awsui-value-large">{totalEBSSize} GB</Box>
            <Box color="text-body-secondary">Private storage</Box>
          </SpaceBetween>
        </ColumnLayout>
      </Container>
    </SpaceBetween>
  )
}
