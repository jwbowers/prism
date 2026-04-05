import { useState, useEffect } from 'react'
import { toast } from 'sonner'
import {
  SpaceBetween,
  Header,
  Container,
  Button,
  Box,
  ColumnLayout,
  Link,
  Spinner,
  Modal,
  Alert,
  Badge,
  Cards,
  FormField,
  Input,
  Select,
} from '../lib/cloudscape-shim'
import { StatusIndicator } from '../components/status-indicator'
import { useApi } from '../hooks/use-api'
import type { MarketplaceTemplate, MarketplaceCategory } from '../lib/types'

interface MarketplaceViewProps {
  marketplaceTemplates: MarketplaceTemplate[]
  marketplaceCategories: MarketplaceCategory[]
  loading: boolean
  onRefresh: () => Promise<void>
}

export function MarketplaceView({
  marketplaceTemplates,
  marketplaceCategories,
  loading,
  onRefresh,
}: MarketplaceViewProps) {
  const api = useApi()
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string>('')
  const [selectedTemplate, setSelectedTemplate] = useState<MarketplaceTemplate | null>(null)
  const [installModalVisible, setInstallModalVisible] = useState(false)
  const [filteredTemplates, setFilteredTemplates] = useState<MarketplaceTemplate[]>(marketplaceTemplates)

  useEffect(() => {
    let filtered = marketplaceTemplates
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      filtered = filtered.filter(t =>
        t.name.toLowerCase().includes(query) ||
        t.display_name.toLowerCase().includes(query) ||
        t.description.toLowerCase().includes(query) ||
        (t.tags && t.tags.some(tag => tag.toLowerCase().includes(query)))
      )
    }
    if (selectedCategory) {
      filtered = filtered.filter(t => t.category === selectedCategory)
    }
    setFilteredTemplates(filtered)
  }, [searchQuery, selectedCategory, marketplaceTemplates])

  const handleInstallTemplate = async () => {
    if (!selectedTemplate) return
    try {
      await api.installMarketplaceTemplate(selectedTemplate.id, selectedTemplate.id)
      toast.success(`Installing template: ${selectedTemplate.display_name}`)
      setInstallModalVisible(false)
      setSelectedTemplate(null)
      await onRefresh()
    } catch (error) {
      toast.error(`Failed to install template: ${error}`)
    }
  }

  const renderRatingStars = (rating: number) => {
    const stars = []
    for (let i = 1; i <= 5; i++) {
      stars.push(i <= rating ? '★' : '☆')
    }
    return stars.join('')
  }

  return (
    <SpaceBetween size="l">
      <Header
        variant="h1"
        description="Discover and install community-contributed research templates"
        counter={`(${filteredTemplates.length} templates)`}
        actions={
          <Button onClick={onRefresh} disabled={loading}>
            {loading ? <Spinner /> : 'Refresh'}
          </Button>
        }
      >
        Template Marketplace
      </Header>

      <Container>
        <SpaceBetween size="m">
          <FormField label="Search templates" description="Search by name, description, or tags">
            <Input
              value={searchQuery}
              onChange={({ detail }) => setSearchQuery(detail.value)}
              placeholder="Search templates..."
              clearAriaLabel="Clear search"
              type="search"
            />
          </FormField>
          <FormField label="Category" description="Filter by template category">
            <Select
              selectedOption={selectedCategory ? { label: selectedCategory, value: selectedCategory } : null}
              onChange={({ detail }) => setSelectedCategory(detail.selectedOption?.value || '')}
              options={[
                { label: 'All Categories', value: '' },
                ...marketplaceCategories.map(c => ({ label: `${c.name} (${c.count})`, value: c.id }))
              ]}
              placeholder="All Categories"
              selectedAriaLabel="Selected"
            />
          </FormField>
        </SpaceBetween>
      </Container>

      <Cards
        cardDefinition={{
          header: (item: MarketplaceTemplate) => (
            <SpaceBetween direction="horizontal" size="xs">
              <Link fontSize="heading-m" onFollow={() => setSelectedTemplate(item)}>
                {item.display_name || item.name}
              </Link>
              {item.verified && <Badge color="blue">Verified</Badge>}
              {item.featured && <Badge color="green">Featured</Badge>}
            </SpaceBetween>
          ),
          sections: [
            {
              id: 'description',
              content: (item: MarketplaceTemplate) => (
                <Box>
                  <Box variant="p" color="text-body-secondary">
                    {item.description}
                  </Box>
                </Box>
              )
            },
            {
              id: 'metadata',
              content: (item: MarketplaceTemplate) => (
                <ColumnLayout columns={2} variant="text-grid">
                  <div>
                    <Box variant="awsui-key-label">Publisher</Box>
                    <Box>{item.publisher || item.author}</Box>
                  </div>
                  <div>
                    <Box variant="awsui-key-label">Category</Box>
                    <Badge>{item.category}</Badge>
                  </div>
                  <div>
                    <Box variant="awsui-key-label">Rating</Box>
                    <Box color={item.rating >= 4 ? 'text-status-success' : 'inherit'}>
                      {renderRatingStars(item.rating)} ({item.rating.toFixed(1)})
                    </Box>
                  </div>
                  <div>
                    <Box variant="awsui-key-label">Downloads</Box>
                    <Box>{item.downloads.toLocaleString()}</Box>
                  </div>
                </ColumnLayout>
              )
            },
            {
              id: 'tags',
              content: (item: MarketplaceTemplate) =>
                item.tags && item.tags.length > 0 ? (
                  <SpaceBetween direction="horizontal" size="xs">
                    {item.tags.slice(0, 5).map(tag => (
                      <Badge key={tag} color="grey">{tag}</Badge>
                    ))}
                  </SpaceBetween>
                ) : null
            },
            {
              id: 'actions',
              content: (item: MarketplaceTemplate) => (
                <SpaceBetween direction="horizontal" size="xs">
                  <Button
                    onClick={() => {
                      setSelectedTemplate(item)
                      setInstallModalVisible(true)
                    }}
                  >
                    Install
                  </Button>
                  <Button onClick={() => setSelectedTemplate(item)}>
                    View Details
                  </Button>
                </SpaceBetween>
              )
            }
          ]
        }}
        items={filteredTemplates}
        cardsPerRow={[{ cards: 1 }, { minWidth: 500, cards: 2 }]}
        loading={loading}
        loadingText="Loading marketplace templates..."
        empty={
          <Box textAlign="center" padding="xl">
            <Box variant="strong">No templates found</Box>
            <Box variant="p" color="text-body-secondary">
              {searchQuery || selectedCategory
                ? 'Try adjusting your search or filter criteria.'
                : 'No marketplace templates available.'}
            </Box>
          </Box>
        }
      />

      {selectedTemplate && !installModalVisible && (
        <Container
          header={
            <Header
              variant="h2"
              actions={<Button onClick={() => setSelectedTemplate(null)}>Close</Button>}
            >
              {selectedTemplate.display_name || selectedTemplate.name}
            </Header>
          }
        >
          <SpaceBetween size="l">
            <ColumnLayout columns={2}>
              <SpaceBetween size="m">
                <div>
                  <Box variant="awsui-key-label">Publisher</Box>
                  <Box>{selectedTemplate.publisher || selectedTemplate.author}</Box>
                </div>
                <div>
                  <Box variant="awsui-key-label">Category</Box>
                  <Badge>{selectedTemplate.category}</Badge>
                </div>
                <div>
                  <Box variant="awsui-key-label">Version</Box>
                  <Box>{selectedTemplate.version}</Box>
                </div>
                <div>
                  <Box variant="awsui-key-label">Verified</Box>
                  {selectedTemplate.verified ? (
                    <StatusIndicator type="success">Verified Publisher</StatusIndicator>
                  ) : (
                    <StatusIndicator type="pending">Community</StatusIndicator>
                  )}
                </div>
              </SpaceBetween>
              <SpaceBetween size="m">
                <div>
                  <Box variant="awsui-key-label">Rating</Box>
                  <Box color={selectedTemplate.rating >= 4 ? 'text-status-success' : 'inherit'}>
                    {renderRatingStars(selectedTemplate.rating)} ({selectedTemplate.rating.toFixed(1)})
                  </Box>
                </div>
                <div>
                  <Box variant="awsui-key-label">Downloads</Box>
                  <Box>{selectedTemplate.downloads.toLocaleString()}</Box>
                </div>
                <div>
                  <Box variant="awsui-key-label">Created</Box>
                  <Box>{new Date(selectedTemplate.created_at).toLocaleDateString()}</Box>
                </div>
                <div>
                  <Box variant="awsui-key-label">Last Updated</Box>
                  <Box>{new Date(selectedTemplate.updated_at).toLocaleDateString()}</Box>
                </div>
              </SpaceBetween>
            </ColumnLayout>

            <div>
              <Box variant="awsui-key-label">Description</Box>
              <Box variant="p">{selectedTemplate.description}</Box>
            </div>

            {selectedTemplate.tags && selectedTemplate.tags.length > 0 && (
              <div>
                <Box variant="awsui-key-label">Tags</Box>
                <SpaceBetween direction="horizontal" size="xs">
                  {selectedTemplate.tags.map(tag => (
                    <Badge key={tag} color="grey">{tag}</Badge>
                  ))}
                </SpaceBetween>
              </div>
            )}

            {selectedTemplate.badges && selectedTemplate.badges.length > 0 && (
              <div>
                <Box variant="awsui-key-label">Badges</Box>
                <SpaceBetween direction="horizontal" size="xs">
                  {selectedTemplate.badges.map(badge => (
                    <Badge key={badge} color="blue">{badge}</Badge>
                  ))}
                </SpaceBetween>
              </div>
            )}

            {selectedTemplate.ami_available && (
              <Alert type="info">
                This template has pre-built AMIs available for faster launches (30 seconds vs 5-8 minutes).
              </Alert>
            )}

            <Button variant="primary" onClick={() => setInstallModalVisible(true)}>
              Install Template
            </Button>
          </SpaceBetween>
        </Container>
      )}

      <Modal
        visible={installModalVisible}
        onDismiss={() => { setInstallModalVisible(false); setSelectedTemplate(null) }}
        header="Install Marketplace Template"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => { setInstallModalVisible(false); setSelectedTemplate(null) }}>Cancel</Button>
              <Button variant="primary" onClick={handleInstallTemplate}>Install</Button>
            </SpaceBetween>
          </Box>
        }
      >
        {selectedTemplate && (
          <SpaceBetween size="m">
            <Alert type="info">
              This will download and install the template to your local templates directory.
            </Alert>
            <div>
              <Box variant="strong">Template:</Box> {selectedTemplate.display_name || selectedTemplate.name}
              <br />
              <Box variant="strong">Publisher:</Box> {selectedTemplate.publisher || selectedTemplate.author}
              <br />
              <Box variant="strong">Version:</Box> {selectedTemplate.version}
              <br />
              {selectedTemplate.verified && (
                <>
                  <Box variant="strong">Status:</Box> <StatusIndicator type="success">Verified Publisher</StatusIndicator>
                </>
              )}
            </div>
          </SpaceBetween>
        )}
      </Modal>
    </SpaceBetween>
  )
}
