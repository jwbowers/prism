import { useState } from 'react'
import { toast } from 'sonner'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  Cards,
  Badge,
  FormField,
  Input,
  Select,
  Alert,
  Container,
  Header,
  ColumnLayout,
  Tabs,
  Wizard,
  ProgressBar,
} from '../lib/cloudscape-shim'
import { useApi } from '../hooks/use-api'
import type { Template } from '../lib/types'
import { getTemplateName, getTemplateSlug, getTemplateDescription, getTemplateTags } from '../lib/template-utils'

interface QuickStartWizardProps {
  visible: boolean
  templates: Record<string, Template>
  onDismiss: () => void
  onSuccess: () => void
  onNavigateToWorkspaces: () => void
}

export function QuickStartWizard({ visible, templates, onDismiss, onSuccess, onNavigateToWorkspaces }: QuickStartWizardProps) {
  const api = useApi()

  const [activeStepIndex, setActiveStepIndex] = useState(0)
  const [config, setConfig] = useState({
    selectedTemplate: null as Template | null,
    workspaceName: '',
    size: 'M',
    launchInProgress: false,
    launchedWorkspaceId: null as string | null
  })

  const handleWizardNavigate = (event: { detail: { requestedStepIndex: number; reason: string } }) => {
    setActiveStepIndex(event.detail.requestedStepIndex)
  }

  const handleWizardCancel = () => {
    onDismiss()
    setActiveStepIndex(0)
    setConfig({
      selectedTemplate: null,
      workspaceName: '',
      size: 'M',
      launchInProgress: false,
      launchedWorkspaceId: null
    })
  }

  const handleWizardSubmit = async () => {
    if (!config.selectedTemplate) return

    setConfig(prev => ({ ...prev, launchInProgress: true }))
    setActiveStepIndex(3) // Move to progress step

    try {
      const result = await api.launchInstance(
        getTemplateSlug(config.selectedTemplate),
        config.workspaceName,
        config.size
      )

      setConfig(prev => ({
        ...prev,
        launchInProgress: false,
        launchedWorkspaceId: result?.id || null
      }))

      toast.success(`Workspace "${config.workspaceName}" launched successfully!`)

      // Refresh workspace list
      await onSuccess()
    } catch (error) {
      setConfig(prev => ({ ...prev, launchInProgress: false }))
      toast.error(`Failed to launch workspace: ${error instanceof Error ? error.message : 'Unknown error'}`)
    }
  }

  const getSizeDescription = (size: string): string => {
    const descriptions: Record<string, string> = {
      'S': 'Small - 2 vCPU, 4GB RAM (~$0.08/hour)',
      'M': 'Medium - 4 vCPU, 8GB RAM (~$0.16/hour)',
      'L': 'Large - 8 vCPU, 16GB RAM (~$0.32/hour)',
      'XL': 'Extra Large - 16 vCPU, 32GB RAM (~$0.64/hour)'
    }
    return descriptions[size] || descriptions['M']
  }

  const getCategoryTemplates = (category: string): Template[] => {
    return Object.values(templates).filter(t => {
      const name = getTemplateName(t).toLowerCase()
      const desc = getTemplateDescription(t).toLowerCase()
      switch (category) {
        case 'ml':
          return name.includes('machine learning') || name.includes('ml') || name.includes('python') && desc.includes('tensorflow')
        case 'datascience':
          return name.includes('python') || name.includes('jupyter') || name.includes('data')
        case 'r':
          return name.includes('r ') || name.includes('rstudio')
        case 'bio':
          return name.includes('bio') || name.includes('genomics')
        default:
          return true
      }
    })
  }

  return (
    <Modal
      visible={visible}
      onDismiss={handleWizardCancel}
      size="large"
      header="Quick Start - Launch Workspace"
    >
      <Wizard
        i18nStrings={{
          stepNumberLabel: stepNumber => `Step ${stepNumber}`,
          collapsedStepsLabel: (stepNumber, stepsCount) => `Step ${stepNumber} of ${stepsCount}`,
          skipToButtonLabel: (step) => `Skip to ${step.title}`,
          navigationAriaLabel: "Steps",
          cancelButton: "Cancel",
          previousButton: "Previous",
          nextButton: "Next",
          submitButton: "Launch Workspace",
          optional: "optional"
        }}
        onNavigate={handleWizardNavigate}
        onCancel={handleWizardCancel}
        onSubmit={handleWizardSubmit}
        activeStepIndex={activeStepIndex}
        isLoadingNextStep={config.launchInProgress}
        steps={[
          {
            title: "Select Template",
            description: "Choose a pre-configured research environment",
            content: (
              <SpaceBetween size="l">
                <Alert type="info">
                  Select a template that matches your research needs. Each template includes specialized software and tools.
                </Alert>

                <Tabs
                  tabs={[
                    {
                      id: "all",
                      label: "All Templates",
                      content: (
                        <Cards
                          cardDefinition={{
                            header: item => (
                              <Box variant="h3">{getTemplateName(item)}</Box>
                            ),
                            sections: [
                              {
                                id: "description",
                                content: item => getTemplateDescription(item)
                              },
                              {
                                id: "tags",
                                content: item => (
                                  <SpaceBetween direction="horizontal" size="xs">
                                    {getTemplateTags(item).slice(0, 3).map((tag, idx) => (
                                      <Badge key={idx} color="blue">{tag}</Badge>
                                    ))}
                                  </SpaceBetween>
                                )
                              }
                            ]
                          }}
                          items={Object.values(templates)}
                          selectionType="single"
                          selectedItems={config.selectedTemplate ? [config.selectedTemplate] : []}
                          onSelectionChange={({ detail }) => {
                            if (detail.selectedItems.length > 0) {
                              setConfig(prev => ({
                                ...prev,
                                selectedTemplate: detail.selectedItems[0]
                              }))
                            }
                          }}
                          cardsPerRow={[{ cards: 1 }, { minWidth: 500, cards: 2 }]}
                          empty={
                            <Box textAlign="center" color="inherit">
                              <b>No templates available</b>
                              <Box padding={{ bottom: "s" }} variant="p" color="inherit">
                                No research templates found.
                              </Box>
                            </Box>
                          }
                        />
                      )
                    },
                    {
                      id: "ml",
                      label: "ML/AI",
                      content: (
                        <Cards
                          cardDefinition={{
                            header: item => <Box variant="h3">{getTemplateName(item)}</Box>,
                            sections: [{ id: "description", content: item => getTemplateDescription(item) }]
                          }}
                          items={getCategoryTemplates('ml')}
                          selectionType="single"
                          selectedItems={config.selectedTemplate ? [config.selectedTemplate] : []}
                          onSelectionChange={({ detail }) => {
                            if (detail.selectedItems.length > 0) {
                              setConfig(prev => ({ ...prev, selectedTemplate: detail.selectedItems[0] }))
                            }
                          }}
                          cardsPerRow={[{ cards: 1 }, { minWidth: 500, cards: 2 }]}
                        />
                      )
                    },
                    {
                      id: "datascience",
                      label: "Data Science",
                      content: (
                        <Cards
                          cardDefinition={{
                            header: item => <Box variant="h3">{getTemplateName(item)}</Box>,
                            sections: [{ id: "description", content: item => getTemplateDescription(item) }]
                          }}
                          items={getCategoryTemplates('datascience')}
                          selectionType="single"
                          selectedItems={config.selectedTemplate ? [config.selectedTemplate] : []}
                          onSelectionChange={({ detail }) => {
                            if (detail.selectedItems.length > 0) {
                              setConfig(prev => ({ ...prev, selectedTemplate: detail.selectedItems[0] }))
                            }
                          }}
                          cardsPerRow={[{ cards: 1 }, { minWidth: 500, cards: 2 }]}
                        />
                      )
                    }
                  ]}
                />
              </SpaceBetween>
            ),
            isOptional: false
          },
          {
            title: "Configure Workspace",
            description: "Set workspace name and size",
            content: (
              <SpaceBetween size="l">
                <FormField
                  label="Workspace Name"
                  description="Choose a unique name for your workspace"
                  constraintText="Use lowercase letters, numbers, and hyphens only"
                >
                  <Input
                    value={config.workspaceName}
                    onChange={({ detail }) => setConfig(prev => ({ ...prev, workspaceName: detail.value }))}
                    placeholder="my-research-workspace"
                  />
                </FormField>

                <FormField
                  label="Workspace Size"
                  description="Choose the compute resources for your workspace"
                >
                  <Select
                    selectedOption={{ label: getSizeDescription(config.size), value: config.size }}
                    onChange={({ detail }) => setConfig(prev => ({ ...prev, size: detail.selectedOption.value || 'M' }))}
                    options={[
                      { label: getSizeDescription('S'), value: 'S' },
                      { label: getSizeDescription('M'), value: 'M' },
                      { label: getSizeDescription('L'), value: 'L' },
                      { label: getSizeDescription('XL'), value: 'XL' }
                    ]}
                  />
                </FormField>

                <Alert type="info">
                  💡 <strong>Tip:</strong> Start with Medium size for most workloads. You can always stop and resize later.
                </Alert>
              </SpaceBetween>
            ),
            isOptional: false
          },
          {
            title: "Review & Launch",
            description: "Review your configuration",
            content: (
              <SpaceBetween size="l">
                <Container header={<Header variant="h3">Configuration Summary</Header>}>
                  <ColumnLayout columns={2} variant="text-grid">
                    <div>
                      <Box variant="awsui-key-label">Template</Box>
                      <Box>{config.selectedTemplate ? getTemplateName(config.selectedTemplate) : 'None'}</Box>
                    </div>
                    <div>
                      <Box variant="awsui-key-label">Workspace Name</Box>
                      <Box>{config.workspaceName || 'Not set'}</Box>
                    </div>
                    <div>
                      <Box variant="awsui-key-label">Size</Box>
                      <Box>{getSizeDescription(config.size)}</Box>
                    </div>
                    <div>
                      <Box variant="awsui-key-label">Estimated Cost</Box>
                      <Box data-testid="cost-estimate">
                        {config.size === 'S' && '~$0.08/hour (~$58/month)'}
                        {config.size === 'M' && '~$0.16/hour (~$115/month)'}
                        {config.size === 'L' && '~$0.32/hour (~$230/month)'}
                        {config.size === 'XL' && '~$0.64/hour (~$460/month)'}
                      </Box>
                    </div>
                  </ColumnLayout>
                </Container>

                <Alert type="warning">
                  <strong>Cost Reminder:</strong> Remember to stop or hibernate your workspace when not in use to save costs.
                </Alert>

                {config.selectedTemplate && config.workspaceName && (
                  <Alert type="success">
                    ✅ Ready to launch! Click "Launch Workspace" to proceed.
                  </Alert>
                )}
              </SpaceBetween>
            ),
            isOptional: false
          },
          {
            title: "Launch Progress",
            description: "Launching your workspace",
            content: (
              <SpaceBetween size="l">
                {config.launchInProgress && (
                  <Box>
                    <ProgressBar value={50} description="Launching workspace..." />
                    <Box margin={{ top: "m" }} color="text-body-secondary">
                      This typically takes 2-3 minutes. Your workspace is being provisioned with all required software and configurations.
                    </Box>
                  </Box>
                )}

                {!config.launchInProgress && config.launchedWorkspaceId && (
                  <Alert type="success" header="Workspace Launched Successfully!">
                    <SpaceBetween size="m">
                      <Box>
                        Your workspace <strong>{config.workspaceName}</strong> is now running and ready to use.
                      </Box>
                      <Box>
                        <strong>Next Steps:</strong>
                        <ul>
                          <li>Connect via SSH or web interface from the Workspaces page</li>
                          <li>Access pre-installed software and tools</li>
                          <li>Remember to stop or hibernate when done to save costs</li>
                        </ul>
                      </Box>
                      <SpaceBetween direction="horizontal" size="s">
                        <Button
                          variant="primary"
                          onClick={() => {
                            onNavigateToWorkspaces()
                            handleWizardCancel()
                          }}
                        >
                          View Workspace
                        </Button>
                        <Button onClick={handleWizardCancel}>
                          Close
                        </Button>
                      </SpaceBetween>
                    </SpaceBetween>
                  </Alert>
                )}

                {!config.launchInProgress && !config.launchedWorkspaceId && (
                  <Alert type="info">
                    Click "Launch Workspace" to start the deployment process.
                  </Alert>
                )}
              </SpaceBetween>
            ),
            isOptional: false
          }
        ]}
      />
    </Modal>
  )
}
