import { useState, useEffect } from 'react'
import { toast } from 'sonner'
import {
  SpaceBetween,
  Header,
  Button,
  Box,
  Table,
  Modal,
  Alert,
  Badge,
  Spinner,
  FormField,
  Input,
} from '../lib/cloudscape-shim'
import { useApi } from '../hooks/use-api'

interface ProfileFormData {
  name: string
  aws_profile: string
  region: string
}

interface Profile {
  id: string
  name: string
  aws_profile: string
  region: string
  type: string
  default?: boolean
  Default?: boolean
}

export function ProfileSelectorView() {
  const api = useApi()
  const [profiles, setProfiles] = useState<Profile[]>([])
  const [currentProfileId, setCurrentProfileId] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showEditDialog, setShowEditDialog] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [selectedProfile, setSelectedProfile] = useState<Profile | null>(null)
  const [formData, setFormData] = useState<ProfileFormData>({ name: '', aws_profile: '', region: '' })
  const [validationError, setValidationError] = useState('')

  const loadProfiles = async () => {
    setLoading(true)
    try {
      const response = await api.getProfiles()
      setProfiles(response)
      const current = response.find((p: Profile) => p.default || p.Default)
      if (current) {
        setCurrentProfileId(current.id)
      }
    } catch (error) {
      console.error('Failed to load profiles:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadProfiles()
  }, [])

  const validateRegion = (region: string): boolean => {
    if (!region) return true
    return /^[a-z]{2}(-[a-z]+)+-\d$/.test(region)
  }

  const handleCreateProfile = async () => {
    setValidationError('')
    if (!formData.name) { setValidationError('Profile name is required'); return }
    if (!formData.aws_profile) { setValidationError('AWS profile is required'); return }
    if (!validateRegion(formData.region)) {
      setValidationError('Region must be a valid AWS region format (e.g., us-east-1, eu-west-2)')
      return
    }
    setLoading(true)
    try {
      await api.createProfile(formData)
      setShowCreateDialog(false)
      setFormData({ name: '', aws_profile: '', region: '' })
      await loadProfiles()
    } catch (error: any) {
      setValidationError(error.message || 'Failed to create profile')
    } finally {
      setLoading(false)
    }
  }

  const handleUpdateProfile = async () => {
    if (!selectedProfile) return
    setValidationError('')
    if (!formData.name) { setValidationError('Profile name is required'); return }
    if (!validateRegion(formData.region)) {
      setValidationError('Region must be a valid AWS region format (e.g., us-east-1, eu-west-2)')
      return
    }
    setLoading(true)
    try {
      await api.updateProfile(selectedProfile.id, formData)
      setShowEditDialog(false)
      setSelectedProfile(null)
      setFormData({ name: '', aws_profile: '', region: '' })
      await loadProfiles()
    } catch (error: any) {
      setValidationError(error.message || 'Failed to update profile')
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteProfile = async () => {
    if (!selectedProfile) return
    setLoading(true)
    try {
      await api.deleteProfile(selectedProfile.id)
      setShowDeleteDialog(false)
      setSelectedProfile(null)
      await loadProfiles()
    } catch (error: any) {
      setValidationError(error.message || 'Failed to delete profile')
    } finally {
      setLoading(false)
    }
  }

  const handleSwitchProfile = async (profileId: string) => {
    try {
      const activatedProfile = await api.switchProfile(profileId)
      await loadProfiles()
      toast.success(`Switched to profile: ${activatedProfile.name}`)
    } catch (error) {
      toast.error(`Failed to switch profile: ${error}`)
    }
  }

  const openCreateDialog = () => {
    setFormData({ name: '', aws_profile: '', region: '' })
    setValidationError('')
    setShowCreateDialog(true)
  }

  const openEditDialog = (profile: Profile) => {
    setSelectedProfile(profile)
    setFormData({ name: profile.name, aws_profile: profile.aws_profile, region: profile.region || '' })
    setValidationError('')
    setShowEditDialog(true)
  }

  const openDeleteDialog = (profile: Profile) => {
    setSelectedProfile(profile)
    setValidationError('')
    setShowDeleteDialog(true)
  }

  return (
    <SpaceBetween size="l">
      <Header
        variant="h1"
        description="Manage AWS profiles for different accounts and regions"
        counter={`(${profiles.length} profiles)`}
        actions={
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={loadProfiles} disabled={loading}>
              {loading ? <Spinner /> : 'Refresh'}
            </Button>
            <Button
              variant="primary"
              onClick={openCreateDialog}
              data-testid="create-profile-button"
            >
              Create Profile
            </Button>
          </SpaceBetween>
        }
      >
        Profile Management
      </Header>

      <Table
        data-testid="profiles-table"
        columnDefinitions={[
          {
            id: 'name',
            header: 'Profile Name',
            cell: (item: Profile) => (
              <Box>
                {item.id === currentProfileId && (
                  <Badge color="blue" data-testid="current-profile-badge">Current</Badge>
                )}{' '}
                {item.name}
              </Box>
            ),
            sortingField: 'name'
          },
          {
            id: 'aws_profile',
            header: 'AWS Profile',
            cell: (item: Profile) => item.aws_profile,
            sortingField: 'aws_profile'
          },
          {
            id: 'region',
            header: 'Region',
            cell: (item: Profile) => item.region || '-',
            sortingField: 'region'
          },
          {
            id: 'type',
            header: 'Type',
            cell: (item: Profile) => item.type,
            sortingField: 'type'
          },
          {
            id: 'actions',
            header: 'Actions',
            cell: (item: Profile) => (
              <SpaceBetween direction="horizontal" size="xs">
                {item.id !== currentProfileId && (
                  <Button
                    onClick={() => handleSwitchProfile(item.id)}
                    data-testid={`switch-profile-${item.name}`}
                  >
                    Switch
                  </Button>
                )}
                <Button
                  onClick={() => openEditDialog(item)}
                  data-testid={`edit-profile-${item.name}`}
                >
                  Edit
                </Button>
                {item.id !== currentProfileId && (
                  <Button
                    onClick={() => openDeleteDialog(item)}
                    data-testid={`delete-profile-${item.name}`}
                  >
                    Delete
                  </Button>
                )}
              </SpaceBetween>
            )
          }
        ]}
        items={profiles}
        loading={loading}
        loadingText="Loading profiles..."
        empty={
          <Box textAlign="center" color="inherit" padding={{ vertical: 'xl' }}>
            <SpaceBetween size="m">
              <b>No profiles</b>
              <Button onClick={openCreateDialog}>Create Profile</Button>
            </SpaceBetween>
          </Box>
        }
      />

      <Modal
        visible={showCreateDialog}
        onDismiss={() => setShowCreateDialog(false)}
        header="Create Profile"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setShowCreateDialog(false)}>Cancel</Button>
              <Button variant="primary" onClick={handleCreateProfile} disabled={loading}>Create</Button>
            </SpaceBetween>
          </Box>
        }
      >
        <SpaceBetween size="m">
          {validationError && (
            <Alert type="error" data-testid="validation-error">{validationError}</Alert>
          )}
          <FormField label="Profile Name" description="A descriptive name for this profile">
            <Input
              value={formData.name}
              onChange={({ detail }) => setFormData({ ...formData, name: detail.value })}
              placeholder="e.g., production, development"
              data-testid="profile-name-input"
            />
          </FormField>
          <FormField label="AWS Profile" description="AWS CLI profile name from ~/.aws/credentials">
            <Input
              value={formData.aws_profile}
              onChange={({ detail }) => setFormData({ ...formData, aws_profile: detail.value })}
              placeholder="default"
              data-testid="aws-profile-input"
            />
          </FormField>
          <FormField label="Region" description="Default AWS region (optional)">
            <Input
              value={formData.region}
              onChange={({ detail }) => setFormData({ ...formData, region: detail.value })}
              placeholder="us-west-2"
              data-testid="region-input"
            />
          </FormField>
        </SpaceBetween>
      </Modal>

      <Modal
        visible={showEditDialog}
        onDismiss={() => setShowEditDialog(false)}
        header="Edit Profile"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setShowEditDialog(false)}>Cancel</Button>
              <Button variant="primary" onClick={handleUpdateProfile} disabled={loading}>Save</Button>
            </SpaceBetween>
          </Box>
        }
      >
        <SpaceBetween size="m">
          {validationError && (
            <Alert type="error" data-testid="validation-error">{validationError}</Alert>
          )}
          <FormField label="Profile Name">
            <Input
              value={formData.name}
              onChange={({ detail }) => setFormData({ ...formData, name: detail.value })}
              data-testid="edit-profile-name-input"
            />
          </FormField>
          <FormField label="AWS Profile">
            <Input
              value={formData.aws_profile}
              onChange={({ detail }) => setFormData({ ...formData, aws_profile: detail.value })}
              data-testid="edit-aws-profile-input"
            />
          </FormField>
          <FormField label="Region">
            <Input
              value={formData.region}
              onChange={({ detail }) => setFormData({ ...formData, region: detail.value })}
              data-testid="edit-region-input"
            />
          </FormField>
        </SpaceBetween>
      </Modal>

      <Modal
        visible={showDeleteDialog}
        onDismiss={() => setShowDeleteDialog(false)}
        header="Delete Profile"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setShowDeleteDialog(false)}>Cancel</Button>
              <Button variant="primary" onClick={handleDeleteProfile} disabled={loading}>Delete</Button>
            </SpaceBetween>
          </Box>
        }
      >
        <SpaceBetween size="m">
          {validationError && (
            <Alert type="error" data-testid="validation-error">{validationError}</Alert>
          )}
          {selectedProfile?.id === currentProfileId ? (
            <Alert type="warning">
              Cannot delete the currently active profile. Switch to a different profile first.
            </Alert>
          ) : (
            <Box>
              Are you sure you want to delete the profile <strong>{selectedProfile?.name}</strong>?
              This action cannot be undone.
            </Box>
          )}
        </SpaceBetween>
      </Modal>
    </SpaceBetween>
  )
}
