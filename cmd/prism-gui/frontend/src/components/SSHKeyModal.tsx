import React from 'react';
import {
  Modal,
  Box,
  SpaceBetween,
  Button,
  Alert,
  Spinner,
  Container,
  Header,
  ColumnLayout,
} from '../lib/cloudscape-shim';

interface SSHKeyModalProps {
  visible: boolean;
  username: string;
  onDismiss: () => void;
  onGenerate: (username: string) => Promise<SSHKeyResponse>;
}

interface SSHKeyResponse {
  public_key: string;
  private_key: string;
  fingerprint: string;
  generated_at: string;
}

export const SSHKeyModal: React.FC<SSHKeyModalProps> = ({
  visible,
  username,
  onDismiss,
  onGenerate
}) => {
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [keyData, setKeyData] = React.useState<SSHKeyResponse | null>(null);
  const [copied, setCopied] = React.useState(false);

  React.useEffect(() => {
    if (!visible) {
      // Reset state when modal closes
      setKeyData(null);
      setError(null);
      setCopied(false);
    }
  }, [visible]);

  const handleGenerate = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await onGenerate(username);
      setKeyData(response);
    } catch (err: any) {
      setError(err.message || 'Failed to generate SSH key');
    } finally {
      setLoading(false);
    }
  };

  const handleCopyPublicKey = () => {
    if (keyData?.public_key) {
      navigator.clipboard.writeText(keyData.public_key);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleDownloadPrivateKey = () => {
    if (keyData?.private_key) {
      const blob = new Blob([keyData.private_key], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${username}_id_rsa`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    }
  };

  return (
    <Modal
      visible={visible}
      onDismiss={onDismiss}
      header={`SSH Key Management - ${username}`}
      size="large"
      data-testid="ssh-key-modal"
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            <Button onClick={onDismiss}>Close</Button>
            {!keyData && (
              <Button
                variant="primary"
                onClick={handleGenerate}
                loading={loading}
                data-testid="generate-ssh-key-button"
              >
                Generate SSH Key
              </Button>
            )}
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="l">
        {error && (
          <Alert type="error" dismissible onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {loading && (
          <Container>
            <Box textAlign="center" padding={{ vertical: 'xxl' }}>
              <Spinner size="large" />
              <Box variant="p" padding={{ top: 'm' }}>
                Generating SSH key pair...
              </Box>
            </Box>
          </Container>
        )}

        {!keyData && !loading && (
          <Alert type="info">
            <SpaceBetween size="s">
              <Box variant="p">
                Generate an SSH key pair for user <strong>{username}</strong>.
              </Box>
              <Box variant="p">
                The private key will be available for download immediately after generation.
                For security reasons, the private key cannot be retrieved later.
              </Box>
              <Box variant="p">
                <strong>Important:</strong> Store the private key securely. You will need it to connect to workspaces.
              </Box>
            </SpaceBetween>
          </Alert>
        )}

        {keyData && (
          <SpaceBetween size="l">
            <Alert type="success">
              SSH key pair generated successfully!
            </Alert>

            {/* Key Metadata */}
            <Container header={<Header variant="h3">Key Information</Header>}>
              <ColumnLayout columns={2} variant="text-grid">
                <div>
                  <Box variant="awsui-key-label">Fingerprint</Box>
                  <div data-testid="ssh-key-fingerprint">
                    <code>{keyData.fingerprint}</code>
                  </div>
                </div>
                <div>
                  <Box variant="awsui-key-label">Generated</Box>
                  <div data-testid="ssh-key-generated-at">
                    {new Date(keyData.generated_at).toLocaleString()}
                  </div>
                </div>
              </ColumnLayout>
            </Container>

            {/* Public Key */}
            <Container
              header={
                <Header
                  variant="h3"
                  actions={
                    <Button
                      iconName="copy"
                      onClick={handleCopyPublicKey}
                      data-testid="copy-public-key-button"
                    >
                      {copied ? 'Copied!' : 'Copy'}
                    </Button>
                  }
                >
                  Public Key
                </Header>
              }
            >
              <Box variant="p" color="text-status-info" fontSize="body-s">
                This public key has been automatically added to the user's profile.
              </Box>
              <Box padding={{ top: 's' }}>
                <textarea
                  readOnly
                  value={keyData.public_key}
                  data-testid="ssh-public-key-display"
                  aria-label="SSH public key — copy to authorized_keys on your server"
                  aria-readonly="true"
                  style={{
                    width: '100%',
                    fontFamily: 'monospace',
                    fontSize: '12px',
                    padding: '8px',
                    border: '1px solid #ccc',
                    borderRadius: '4px',
                    minHeight: '80px',
                    resize: 'vertical'
                  }}
                />
              </Box>
            </Container>

            {/* Private Key */}
            <Container
              header={
                <Header
                  variant="h3"
                  actions={
                    <Button
                      iconName="download"
                      onClick={handleDownloadPrivateKey}
                      variant="primary"
                      data-testid="download-private-key-button"
                    >
                      Download Private Key
                    </Button>
                  }
                >
                  Private Key
                </Header>
              }
            >
              <Alert type="warning">
                <SpaceBetween size="s">
                  <Box variant="p">
                    <strong>Download and store this private key immediately.</strong>
                  </Box>
                  <Box variant="p">
                    For security reasons, the private key will not be shown again after closing this dialog.
                  </Box>
                  <Box variant="p">
                    Save it to <code>~/.ssh/{username}_id_rsa</code> and set permissions:
                  </Box>
                  <Box variant="code">
                    chmod 600 ~/.ssh/{username}_id_rsa
                  </Box>
                </SpaceBetween>
              </Alert>

              <Box padding={{ top: 's' }}>
                <textarea
                  readOnly
                  value={keyData.private_key}
                  data-testid="ssh-private-key-display"
                  aria-label="SSH private key — save as ~/.ssh/id_rsa with chmod 600"
                  aria-readonly="true"
                  style={{
                    width: '100%',
                    fontFamily: 'monospace',
                    fontSize: '12px',
                    padding: '8px',
                    border: '1px solid #ccc',
                    borderRadius: '4px',
                    minHeight: '200px',
                    resize: 'vertical'
                  }}
                />
              </Box>
            </Container>

            {/* Usage Instructions */}
            <Container header={<Header variant="h3">Usage Instructions</Header>}>
              <SpaceBetween size="s">
                <Box variant="p">
                  To connect to a workspace using this key:
                </Box>
                <Box variant="code">
                  ssh -i ~/.ssh/{username}_id_rsa {username}@&lt;workspace-ip&gt;
                </Box>
                <Box variant="p">
                  Or add to your SSH config (<code>~/.ssh/config</code>):
                </Box>
                <Box variant="code">
                  Host prism-workspace
                  <br />
                  &nbsp;&nbsp;HostName &lt;workspace-ip&gt;
                  <br />
                  &nbsp;&nbsp;User {username}
                  <br />
                  &nbsp;&nbsp;IdentityFile ~/.ssh/{username}_id_rsa
                </Box>
              </SpaceBetween>
            </Container>
          </SpaceBetween>
        )}
      </SpaceBetween>
    </Modal>
  );
};
