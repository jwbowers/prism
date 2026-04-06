import { useState, useEffect } from 'react'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  Form,
  FormField,
  Input,
  Select,
  Alert,
  Checkbox,
} from '../lib/cloudscape-shim'
import { Template } from '../lib/types'
import { getTemplateName, getTemplateDescription } from '../lib/template-utils'

export interface LaunchConfig {
  name: string
  size: string
  spot: boolean
  hibernation: boolean
  dryRun: boolean
}

export interface LaunchModalProps {
  visible: boolean
  selectedTemplate: Template | null
  onDismiss: () => void
  onLaunch: (config: LaunchConfig) => void
}

export function LaunchModal({ visible, selectedTemplate, onDismiss, onLaunch }: LaunchModalProps) {
  const [launchConfig, setLaunchConfig] = useState<LaunchConfig>({
    name: '',
    size: 'M',
    spot: false,
    hibernation: false,
    dryRun: false,
  })

  // Reset config when modal opens
  useEffect(() => {
    if (visible) {
      setLaunchConfig({
        name: '',
        size: 'M',
        spot: false,
        hibernation: false,
        dryRun: false,
      })
    }
  }, [visible])

  return (
    <Modal
      onDismiss={onDismiss}
      visible={visible}
      header={`Launch ${selectedTemplate ? getTemplateName(selectedTemplate) : 'Research Environment'}`}
      size="medium"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button variant="link" onClick={onDismiss}>
              Cancel
            </Button>
            <Button
              variant="primary"
              disabled={!launchConfig.name.trim()}
              onClick={() => onLaunch(launchConfig)}
            >
              Launch Workspace
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <Form>
        <SpaceBetween size="m">
          <FormField
            label="Workspace name"
            description="Choose a descriptive name for your research project"
            errorText={!launchConfig.name.trim() ? "Workspace name is required" : ""}
          >
            <Input
              value={launchConfig.name}
              onChange={({ detail }) => setLaunchConfig(prev => ({ ...prev, name: detail.value }))}
              placeholder="my-research-project"
            />
          </FormField>

          <FormField label="Workspace size" description="Choose the right size for your workload">
            <Select
              selectedOption={{ label: "Medium (M) - Recommended", value: "M" }}
              onChange={({ detail }) => setLaunchConfig(prev => ({ ...prev, size: detail.selectedOption.value || 'M' }))}
              options={[
                { label: "Small (S) - Light workloads", value: "S" },
                { label: "Medium (M) - Recommended", value: "M" },
                { label: "Large (L) - Heavy compute", value: "L" },
                { label: "Extra Large (XL) - Maximum performance", value: "XL" }
              ]}
              data-testid="instance-size-select"
            />
          </FormField>

          {selectedTemplate && (
            <Alert type="info">
              <Box>
                <Box variant="strong">Template: {getTemplateName(selectedTemplate)}</Box>
                <Box>Description: {getTemplateDescription(selectedTemplate)}</Box>
                {selectedTemplate.package_manager && (
                  <Box>Package Manager: {selectedTemplate.package_manager}</Box>
                )}
                {selectedTemplate.complexity && (
                  <Box>Complexity: {selectedTemplate.complexity}</Box>
                )}
              </Box>
            </Alert>
          )}

          <FormField
            label="Instance Options"
            description="Configure advanced instance settings"
          >
            <SpaceBetween size="s">
              <Checkbox
                checked={launchConfig.spot || false}
                onChange={({ detail }) => setLaunchConfig(prev => ({ ...prev, spot: detail.checked }))}
              >
                Spot instance - use lower-cost spot pricing
              </Checkbox>
              <Checkbox
                checked={launchConfig.hibernation || false}
                onChange={({ detail }) => setLaunchConfig(prev => ({ ...prev, hibernation: detail.checked }))}
              >
                Hibernation - enable instance hibernation support
              </Checkbox>
            </SpaceBetween>
          </FormField>

          <FormField
            label="Validation"
            description="Test your configuration without actually launching resources"
          >
            <Checkbox
              checked={launchConfig.dryRun || false}
              onChange={({ detail }) => setLaunchConfig(prev => ({ ...prev, dryRun: detail.checked }))}
            >
              Dry run mode - validate without creating resources
            </Checkbox>
          </FormField>
        </SpaceBetween>
      </Form>
    </Modal>
  )
}
