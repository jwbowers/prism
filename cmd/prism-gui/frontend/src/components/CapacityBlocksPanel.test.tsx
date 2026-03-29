/**
 * CapacityBlocksPanel unit tests — v0.20.0
 *
 * Tests for: empty state, block list rendering, reserve modal fields,
 * cancel confirmation modal.
 *
 * Uses MSW to mock the daemon API (http://localhost:8947).
 */

import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import CapacityBlocksPanel from './CapacityBlocksPanel';

const BASE = 'http://localhost:8947';

const mockBlocks = [
  {
    id: 'cr-0abc123def456789',
    instance_type: 'p3.8xlarge',
    instance_count: 2,
    availability_zone: 'us-west-2a',
    start_time: '2026-04-01T09:00:00Z',
    end_time: '2026-04-01T17:00:00Z',
    duration_hours: 8,
    state: 'active',
    total_cost: 0,
  },
  {
    id: 'cr-0dead000beef0001',
    instance_type: 'g5.xlarge',
    instance_count: 1,
    availability_zone: '',
    start_time: '2026-05-01T00:00:00Z',
    end_time: '2026-05-01T04:00:00Z',
    duration_hours: 4,
    state: 'payment-pending',
    total_cost: 0,
  },
];

const server = setupServer(
  http.get(`${BASE}/api/v1/capacity-blocks`, () => HttpResponse.json(mockBlocks)),
  http.post(`${BASE}/api/v1/capacity-blocks`, () => HttpResponse.json(mockBlocks[0], { status: 201 })),
  http.delete(`${BASE}/api/v1/capacity-blocks/:id`, () => new HttpResponse(null, { status: 204 })),
);

beforeAll(() => server.listen({ onUnhandledRequest: 'warn' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('CapacityBlocksPanel', () => {
  describe('empty state', () => {
    it('shows "No capacity blocks" when the list is empty', async () => {
      server.use(
        http.get(`${BASE}/api/v1/capacity-blocks`, () => HttpResponse.json([]))
      );
      render(<CapacityBlocksPanel />);
      await waitFor(() => {
        expect(screen.getByText(/no capacity blocks/i)).toBeDefined();
      });
    });
  });

  describe('list rendering', () => {
    it('renders the capacity blocks table', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => {
        expect(screen.getByTestId('capacity-blocks-table')).toBeDefined();
      });
    });

    it('shows block instance types', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => {
        expect(screen.getByText('p3.8xlarge')).toBeDefined();
        expect(screen.getByText('g5.xlarge')).toBeDefined();
      });
    });

    it('shows block states as badges', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => {
        expect(screen.getByText('active')).toBeDefined();
        expect(screen.getByText('payment-pending')).toBeDefined();
      });
    });

    it('shows Cancel button for active blocks', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => {
        expect(screen.getAllByText('Cancel').length).toBeGreaterThanOrEqual(1);
      });
    });
  });

  describe('Reserve modal', () => {
    it('opens Reserve Capacity Block modal on button click', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => screen.getByTestId('reserve-capacity-block-button'));

      fireEvent.click(screen.getByTestId('reserve-capacity-block-button'));
      await waitFor(() => {
        expect(screen.getByTestId('reserve-capacity-block-modal')).toBeDefined();
      });
    });

    it('shows required form fields in reserve modal', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => screen.getByTestId('reserve-capacity-block-button'));
      fireEvent.click(screen.getByTestId('reserve-capacity-block-button'));

      await waitFor(() => {
        expect(screen.getByTestId('reserve-instance-type')).toBeDefined();
        expect(screen.getByTestId('reserve-count-input')).toBeDefined();
        expect(screen.getByTestId('reserve-duration')).toBeDefined();
      });
    });

    it('Reserve button is disabled when no instance type selected', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => screen.getByTestId('reserve-capacity-block-button'));
      fireEvent.click(screen.getByTestId('reserve-capacity-block-button'));

      await waitFor(() => screen.getByTestId('reserve-submit-button'));
      const btn = screen.getByTestId('reserve-submit-button');
      expect((btn as HTMLButtonElement).disabled).toBe(true);
    });
  });

  describe('Cancel confirmation modal', () => {
    it('opens cancel confirmation when Cancel button clicked', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => screen.getAllByText('Cancel'));

      const cancelBtns = screen.getAllByText('Cancel');
      fireEvent.click(cancelBtns[0]);

      await waitFor(() => {
        expect(screen.getByTestId('cancel-capacity-block-modal')).toBeDefined();
      });
    });

    it('shows Cancel Block confirm button', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => screen.getAllByText('Cancel'));

      fireEvent.click(screen.getAllByText('Cancel')[0]);
      await waitFor(() => {
        expect(screen.getByTestId('cancel-confirm-button')).toBeDefined();
      });
    });

    it('shows Keep button in cancel modal', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => screen.getAllByText('Cancel'));

      fireEvent.click(screen.getAllByText('Cancel')[0]);
      await waitFor(() => {
        expect(screen.getByText('Keep')).toBeDefined();
      });
    });
  });

  describe('refresh', () => {
    it('shows Refresh button', async () => {
      render(<CapacityBlocksPanel />);
      await waitFor(() => {
        expect(screen.getByText('Refresh')).toBeDefined();
      });
    });
  });
});
