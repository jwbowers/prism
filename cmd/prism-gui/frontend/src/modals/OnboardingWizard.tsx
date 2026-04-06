import { useState } from 'react'
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  Alert,
  Container,
  ColumnLayout,
  Header,
} from '../lib/cloudscape-shim'

export interface OnboardingWizardProps {
  visible: boolean
  onComplete: () => void
}

export function OnboardingWizard({ visible, onComplete }: OnboardingWizardProps) {
  const [onboardingStep, setOnboardingStep] = useState(0)
  const totalSteps = 3

  const handleNext = () => {
    if (onboardingStep < totalSteps - 1) {
      setOnboardingStep(onboardingStep + 1)
    } else {
      // Complete onboarding
      localStorage.setItem('prism_onboarding_complete', 'true')
      setOnboardingStep(0)
      onComplete()
    }
  }

  const handleBack = () => {
    if (onboardingStep > 0) {
      setOnboardingStep(onboardingStep - 1)
    }
  }

  const handleSkip = () => {
    localStorage.setItem('prism_onboarding_complete', 'true')
    setOnboardingStep(0)
    onComplete()
  }

  return (
    <Modal
      visible={visible}
      onDismiss={handleSkip}
      header={`Welcome to Prism - Step ${onboardingStep + 1} of ${totalSteps}`}
      size="large"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            {onboardingStep > 0 && (
              <Button onClick={handleBack}>
                Back
              </Button>
            )}
            <Button variant="link" onClick={handleSkip}>
              Skip Tour
            </Button>
            <Button variant="primary" onClick={handleNext}>
              {onboardingStep < totalSteps - 1 ? 'Next' : 'Get Started'}
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="l">
        {/* Step 1: AWS Profile Setup */}
        {onboardingStep === 0 && (
          <SpaceBetween size="m">
            <Alert type="info" header="AWS Credentials Configured">
              Prism is already connected to your AWS account using the configured profile.
            </Alert>
            <Box variant="h2">Step 1: AWS Configuration</Box>
            <Box>
              Prism manages cloud workstations in your AWS account. Your current AWS configuration:
            </Box>
            <Container>
              <ColumnLayout columns={2} variant="text-grid">
                <div>
                  <Box variant="awsui-key-label">AWS Profile</Box>
                  <Box fontWeight="bold">aws</Box>
                  <Box variant="small" color="text-body-secondary">
                    Your AWS credentials profile
                  </Box>
                </div>
                <div>
                  <Box variant="awsui-key-label">Region</Box>
                  <Box fontWeight="bold">us-west-2</Box>
                  <Box variant="small" color="text-body-secondary">
                    Resources will be created here
                  </Box>
                </div>
              </ColumnLayout>
            </Container>
            <Box variant="p" color="text-body-secondary">
              Prism uses your AWS credentials to create and manage cloud workstations.
              You maintain full control over your resources and costs.
            </Box>
          </SpaceBetween>
        )}

        {/* Step 2: Template Discovery Tour */}
        {onboardingStep === 1 && (
          <SpaceBetween size="m">
            <Box variant="h2">Step 2: Choose Your Research Environment</Box>
            <Box>
              Prism provides pre-configured templates for different research workflows.
              Each template includes specialized software, libraries, and tools.
            </Box>
            <ColumnLayout columns={2}>
              <Container header={<Header variant="h3">Popular Templates</Header>}>
                <SpaceBetween size="s">
                  <Box>
                    <Box variant="strong">Python Machine Learning</Box>
                    <Box variant="small" color="text-body-secondary">
                      Python 3, Jupyter, TensorFlow, PyTorch, scikit-learn
                    </Box>
                  </Box>
                  <Box>
                    <Box variant="strong">R Research Environment</Box>
                    <Box variant="small" color="text-body-secondary">
                      R, RStudio Server, tidyverse, statistical packages
                    </Box>
                  </Box>
                  <Box>
                    <Box variant="strong">Collaborative Workspace</Box>
                    <Box variant="small" color="text-body-secondary">
                      Multi-language support with Python, R, Julia
                    </Box>
                  </Box>
                </SpaceBetween>
              </Container>
              <Container header={<Header variant="h3">What's Included</Header>}>
                <SpaceBetween size="s">
                  <Box>✓ Pre-installed software and dependencies</Box>
                  <Box>✓ Optimized workspace sizing for your workload</Box>
                  <Box>✓ Persistent storage for your data</Box>
                  <Box>✓ SSH and remote access configured</Box>
                  <Box>✓ Security best practices applied</Box>
                </SpaceBetween>
              </Container>
            </ColumnLayout>
            <Alert type="info">
              You can browse all available templates in the <strong>Templates</strong> section after completing this tour.
            </Alert>
          </SpaceBetween>
        )}

        {/* Step 3: Launch Your First Workspace */}
        {onboardingStep === 2 && (
          <SpaceBetween size="m">
            <Box variant="h2">Step 3: Launch Your First Workstation</Box>
            <Box>
              Ready to get started? Here's how to launch your first cloud workstation:
            </Box>
            <Container>
              <SpaceBetween size="m">
                <div>
                  <Box variant="h4">1. Select a Template</Box>
                  <Box>Choose a template that matches your research needs from the Templates page.</Box>
                </div>
                <div>
                  <Box variant="h4">2. Configure Workspace</Box>
                  <Box>Give your workstation a name and select the appropriate size (Small, Medium, Large).</Box>
                </div>
                <div>
                  <Box variant="h4">3. Launch & Connect</Box>
                  <Box>Prism creates your workspace in minutes. Connect via SSH or web interface when ready.</Box>
                </div>
              </SpaceBetween>
            </Container>
            <Alert type="success" header="You're All Set!">
              After clicking "Get Started", explore the dashboard to see your system status,
              browse templates, and launch your first cloud workstation.
            </Alert>
            <Box variant="p" color="text-body-secondary">
              💡 <strong>Tip:</strong> Start with a Medium (M) sized workspace for most workloads.
              You can always stop, resize, or terminate workspaces to manage costs.
            </Box>
          </SpaceBetween>
        )}
      </SpaceBetween>
    </Modal>
  )
}
