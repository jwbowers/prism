import {
  SpaceBetween,
  Header,
  Container,
  Button,
  Box,
  Spinner,
} from '../lib/cloudscape-shim'
import { getTemplateName, getTemplateSlug, getTemplateDescription } from '../lib/template-utils'
import type { Template } from '../lib/types'

interface TemplateSelectionViewProps {
  templates: Record<string, Template>
  loading: boolean
  selectedTemplate: Template | null
  onRefresh: () => void
  onLaunch: () => void
  onSelectTemplate: (template: Template) => void
}

export function TemplateSelectionView({
  templates,
  loading,
  selectedTemplate,
  onRefresh,
  onLaunch,
  onSelectTemplate,
}: TemplateSelectionViewProps) {
  // Deduplicate templates by name (keep first occurrence)
  const templateList = Object.values(templates).reduce((acc, template) => {
    const name = getTemplateName(template)
    if (!acc.some(t => getTemplateName(t) === name)) {
      acc.push(template)
    }
    return acc
  }, [] as Template[])

  if (loading) {
    return (
      <Container>
        <Box data-testid="loading" textAlign="center" padding="xl">
          <Spinner size="large" />
          <Box variant="p" color="text-body-secondary">
            Loading templates from AWS...
          </Box>
        </Box>
      </Container>
    )
  }

  if (templateList.length === 0) {
    return (
      <Container>
        <Box textAlign="center" padding="xl">
          <Box variant="strong">No templates available</Box>
          <Box variant="p">Unable to load templates. Check your connection.</Box>
          <Button onClick={onRefresh}>Retry</Button>
        </Box>
      </Container>
    )
  }

  return (
    <SpaceBetween size="l">
      <Container
        header={
          <Header
            variant="h1"
            description={`${templateList.length} pre-configured research environments ready to launch`}
            counter={`(${templateList.length} templates)`}
            actions={
              <SpaceBetween direction="horizontal" size="xs">
                <Button onClick={onRefresh} disabled={loading}>
                  {loading ? <Spinner /> : 'Refresh'}
                </Button>
                <Button
                  variant="primary"
                  disabled={!selectedTemplate}
                  onClick={onLaunch}
                >
                  Launch Selected
                </Button>
              </SpaceBetween>
            }
          >
            Research Templates
          </Header>
        }
      >
        <SpaceBetween size="m" data-testid="cards">
          {templateList.map((template, index) => (
            <Container
              key={getTemplateSlug(template) || `${getTemplateName(template)}-${index}`}
              data-testid="template-card"
            >
              <SpaceBetween size="s">
                <Box>
                  <Box variant="h3">{getTemplateName(template)}</Box>
                  <Box variant="small" color="text-body-secondary">
                    {getTemplateDescription(template)}
                  </Box>
                </Box>
                <Box>
                  <Button variant="primary" onClick={() => onSelectTemplate(template)}>
                    Launch Template
                  </Button>
                </Box>
              </SpaceBetween>
            </Container>
          ))}
        </SpaceBetween>
      </Container>
    </SpaceBetween>
  )
}
