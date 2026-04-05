import React from 'react'
import {
  SpaceBetween,
  Header,
  Container,
  Select,
  FormField,
  Alert,
} from '../lib/cloudscape-shim'
import WebView from '../WebView'
import type { Instance, WebService } from '../lib/types'

interface WebViewViewProps {
  instances: Instance[]
}

export function WebViewView({ instances }: WebViewViewProps) {
  const [selectedService, setSelectedService] = React.useState<{instance: string, service: WebService} | null>(null)

  const instancesWithServices = instances.filter(i =>
    i.state === 'running' && i.web_services && i.web_services.length > 0
  )

  if (instancesWithServices.length === 0) {
    return (
      <Container header={<Header variant="h1">Web Services</Header>}>
        <Alert type="info">
          No running instances with web services available. Launch a workspace with Jupyter or RStudio to access web services.
        </Alert>
      </Container>
    )
  }

  const serviceOptions = instancesWithServices.flatMap(instance =>
    (instance.web_services || []).map(service => ({
      label: `${instance.name} - ${service.name} (${service.type})`,
      value: JSON.stringify({ instance: instance.name, service }),
      instanceName: instance.name,
      service: service
    }))
  )

  return (
    <SpaceBetween size="l">
      <Container header={<Header variant="h1">Web Services</Header>}>
        <SpaceBetween size="m">
          <FormField label="Select Web Service">
            <Select
              selectedOption={selectedService ?
                { label: `${selectedService.instance} - ${selectedService.service.name} (${selectedService.service.type})`,
                  value: JSON.stringify(selectedService) } : null}
              onChange={({ detail }) => {
                if (detail.selectedOption.value) {
                  const parsed = JSON.parse(detail.selectedOption.value)
                  setSelectedService(parsed)
                }
              }}
              options={serviceOptions.map(opt => ({ label: opt.label, value: opt.value }))}
              placeholder="Choose a web service"
            />
          </FormField>
          {selectedService && (
            <WebView
              url={selectedService.service.url || ''}
              serviceName={selectedService.service.name}
              instanceName={selectedService.instance}
            />
          )}
        </SpaceBetween>
      </Container>
    </SpaceBetween>
  )
}
