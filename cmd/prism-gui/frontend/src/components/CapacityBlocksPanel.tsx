/**
 * CapacityBlocksPanel — EC2 Capacity Blocks management (v0.20.0, Issue #63)
 *
 * Lists active/pending capacity reservations and allows users to:
 *   - Reserve a new capacity block (instance type, count, AZ, start time, duration)
 *   - View block details (state, start/end, cost estimate)
 *   - Cancel an active block
 */

import { useEffect, useState, useCallback } from 'react';
import {
  Container,
  Header,
  Table,
  Button,
  SpaceBetween,
  Modal,
  FormField,
  Input,
  Select,
  DatePicker,
  Box,
  Badge,
  Alert,
  TextContent,
} from '../lib/cloudscape-shim';

interface CapacityBlock {
  id: string;
  instance_type: string;
  instance_count: number;
  availability_zone: string;
  start_time: string;
  end_time: string;
  duration_hours: number;
  state: string;
  total_cost: number;
}

const BASE_URL = 'http://localhost:8947';

async function apiFetch<T>(path: string, method = 'GET', body?: unknown): Promise<T> {
  const opts: RequestInit = { method, headers: { 'Content-Type': 'application/json' } };
  if (body !== undefined) opts.body = JSON.stringify(body);
  const resp = await fetch(`${BASE_URL}${path}`, opts);
  if (!resp.ok) {
    const text = await resp.text();
    throw new Error(`HTTP ${resp.status}: ${text}`);
  }
  if (resp.status === 204) return undefined as T;
  return resp.json() as T;
}

const DURATION_OPTIONS = [
  { label: '1 hour', value: '1' },
  { label: '2 hours', value: '2' },
  { label: '4 hours', value: '4' },
  { label: '8 hours', value: '8' },
  { label: '12 hours', value: '12' },
  { label: '24 hours', value: '24' },
];

const INSTANCE_TYPE_OPTIONS = [
  { label: 'p3.2xlarge (1× V100)', value: 'p3.2xlarge' },
  { label: 'p3.8xlarge (4× V100)', value: 'p3.8xlarge' },
  { label: 'p3.16xlarge (8× V100)', value: 'p3.16xlarge' },
  { label: 'p4d.24xlarge (8× A100)', value: 'p4d.24xlarge' },
  { label: 'g5.xlarge (1× A10G)', value: 'g5.xlarge' },
  { label: 'g5.12xlarge (4× A10G)', value: 'g5.12xlarge' },
];

function stateColor(state: string): 'blue' | 'green' | 'grey' | 'red' {
  switch (state) {
    case 'active': return 'green';
    case 'payment-pending': return 'blue';
    case 'expired': return 'grey';
    case 'cancelled': return 'red';
    default: return 'grey';
  }
}

export default function CapacityBlocksPanel() {
  const [blocks, setBlocks] = useState<CapacityBlock[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [reserveOpen, setReserveOpen] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<CapacityBlock | null>(null);

  // Reserve form state
  const [rInstanceType, setRInstanceType] = useState<{ label: string; value: string } | null>(null);
  const [rCount, setRCount] = useState('1');
  const [rAZ, setRAZ] = useState('');
  const [rStartDate, setRStartDate] = useState('');
  const [rStartTime, setRStartTime] = useState('09:00');
  const [rDuration, setRDuration] = useState<{ label: string; value: string } | null>(DURATION_OPTIONS[3]);
  const [reserving, setReserving] = useState(false);
  const [cancelling, setCancelling] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await apiFetch<CapacityBlock[]>('/api/v1/capacity-blocks');
      setBlocks(data || []);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  async function handleReserve() {
    if (!rInstanceType || !rStartDate) return;
    setReserving(true);
    try {
      const startISO = `${rStartDate}T${rStartTime}:00Z`;
      await apiFetch('/api/v1/capacity-blocks', 'POST', {
        instance_type: rInstanceType.value,
        instance_count: parseInt(rCount, 10) || 1,
        availability_zone: rAZ || undefined,
        start_time: startISO,
        duration_hours: parseInt(rDuration?.value || '1', 10),
      });
      setReserveOpen(false);
      await load();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setReserving(false);
    }
  }

  async function handleCancel() {
    if (!cancelTarget) return;
    setCancelling(true);
    try {
      await apiFetch(`/api/v1/capacity-blocks/${cancelTarget.id}`, 'DELETE');
      setCancelTarget(null);
      await load();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setCancelling(false);
    }
  }

  return (
    <Container
      header={
        <Header
          variant="h2"
          actions={
            <SpaceBetween direction="horizontal" size="xs">
              <Button onClick={load} iconName="refresh" disabled={loading}>Refresh</Button>
              <Button variant="primary" onClick={() => setReserveOpen(true)} data-testid="reserve-capacity-block-button">
                Reserve Capacity Block
              </Button>
            </SpaceBetween>
          }
        >
          EC2 Capacity Blocks
        </Header>
      }
    >
      <SpaceBetween size="m">
        {error && <Alert type="error" onDismiss={() => setError('')}>{error}</Alert>}

        <TextContent>
          <p>
            Reserve EC2 Capacity Blocks to guarantee GPU instance availability for scheduled ML training runs.
            Blocks use <em>targeted</em> matching — instances launched with the block ID will be guaranteed capacity.
          </p>
        </TextContent>

        <Table
          data-testid="capacity-blocks-table"
          loading={loading}
          loadingText="Loading capacity blocks..."
          items={blocks}
          empty={
            <Box textAlign="center" color="inherit">
              <b>No capacity blocks</b>
              <Box variant="p" color="inherit">Reserve a capacity block to guarantee GPU availability.</Box>
            </Box>
          }
          columnDefinitions={[
            { id: 'id', header: 'ID', cell: (b) => <code style={{ fontSize: '0.85em' }}>{b.id}</code> },
            { id: 'type', header: 'Instance Type', cell: (b) => b.instance_type },
            { id: 'count', header: 'Count', cell: (b) => b.instance_count },
            { id: 'az', header: 'AZ', cell: (b) => b.availability_zone || '—' },
            { id: 'state', header: 'State', cell: (b) => <Badge color={stateColor(b.state)}>{b.state}</Badge> },
            { id: 'start', header: 'Start', cell: (b) => b.start_time ? new Date(b.start_time).toLocaleString() : '—' },
            { id: 'end', header: 'End', cell: (b) => b.end_time ? new Date(b.end_time).toLocaleString() : '—' },
            { id: 'duration', header: 'Hours', cell: (b) => b.duration_hours || '—' },
            {
              id: 'actions', header: 'Actions',
              cell: (b) => (
                b.state === 'active' || b.state === 'payment-pending'
                  ? <Button
                      variant="link"
                      data-testid={`cancel-block-${b.id}`}
                      onClick={() => setCancelTarget(b)}
                    >Cancel</Button>
                  : null
              )
            },
          ]}
        />
      </SpaceBetween>

      {/* Reserve Modal */}
      <Modal
        data-testid="reserve-capacity-block-modal"
        visible={reserveOpen}
        onDismiss={() => setReserveOpen(false)}
        header="Reserve EC2 Capacity Block"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setReserveOpen(false)}>Cancel</Button>
              <Button
                variant="primary"
                loading={reserving}
                disabled={!rInstanceType || !rStartDate}
                onClick={handleReserve}
                data-testid="reserve-submit-button"
              >Reserve</Button>
            </SpaceBetween>
          </Box>
        }
      >
        <SpaceBetween size="m">
          <FormField label="Instance Type" description="GPU instance type for the reservation">
            <Select
              selectedOption={rInstanceType}
              onChange={({ detail }) => setRInstanceType(detail.selectedOption as { label: string; value: string })}
              options={INSTANCE_TYPE_OPTIONS}
              placeholder="Select instance type"
              data-testid="reserve-instance-type"
            />
          </FormField>
          <FormField label="Instance Count">
            <Input
              type="number"
              value={rCount}
              onChange={({ detail }) => setRCount(detail.value)}
              data-testid="reserve-count-input"
            />
          </FormField>
          <FormField label="Availability Zone" description="Optional — omit to let AWS choose">
            <Input
              value={rAZ}
              onChange={({ detail }) => setRAZ(detail.value)}
              placeholder="e.g. us-west-2a"
              data-testid="reserve-az-input"
            />
          </FormField>
          <FormField label="Start Date">
            <DatePicker
              value={rStartDate}
              onChange={({ detail }) => setRStartDate(detail.value)}
              placeholder="YYYY/MM/DD"
              data-testid="reserve-start-date"
            />
          </FormField>
          <FormField label="Start Time (UTC)" description="HH:MM format">
            <Input
              value={rStartTime}
              onChange={({ detail }) => setRStartTime(detail.value)}
              placeholder="09:00"
              data-testid="reserve-start-time"
            />
          </FormField>
          <FormField label="Duration">
            <Select
              selectedOption={rDuration}
              onChange={({ detail }) => setRDuration(detail.selectedOption as { label: string; value: string })}
              options={DURATION_OPTIONS}
              data-testid="reserve-duration"
            />
          </FormField>
        </SpaceBetween>
      </Modal>

      {/* Cancel Confirmation Modal */}
      <Modal
        data-testid="cancel-capacity-block-modal"
        visible={!!cancelTarget}
        onDismiss={() => setCancelTarget(null)}
        header="Cancel Capacity Block"
        footer={
          <Box float="right">
            <SpaceBetween direction="horizontal" size="xs">
              <Button variant="link" onClick={() => setCancelTarget(null)}>Keep</Button>
              <Button
                variant="primary"
                loading={cancelling}
                onClick={handleCancel}
                data-testid="cancel-confirm-button"
              >Cancel Block</Button>
            </SpaceBetween>
          </Box>
        }
      >
        <TextContent>
          <p>Cancel capacity block <strong>{cancelTarget?.id}</strong> ({cancelTarget?.instance_type} × {cancelTarget?.instance_count})?
          This action cannot be undone.</p>
        </TextContent>
      </Modal>
    </Container>
  );
}
