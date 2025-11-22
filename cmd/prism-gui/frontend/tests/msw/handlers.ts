/**
 * MSW (Mock Service Worker) Handlers for Prism Daemon API
 *
 * These handlers intercept HTTP requests to the daemon API and return mock responses.
 * Use these in component tests to avoid starting a real daemon.
 */

import { http, HttpResponse } from 'msw';
import {
  createMockTemplates,
  createMockInstances,
  createMockProfiles,
  createMockVolumes,
  createMockEBSStorages,
  createMockHealthResponse,
  createMockLaunchResponse,
  createMockConnectionInfo,
  createMockHibernationStatus,
  createMockIdlePolicies,
  createMockError,
} from '../utils/mock-data-factories';

const DAEMON_URL = 'http://localhost:8947';

/**
 * Default MSW handlers for all daemon API endpoints
 */
export const handlers = [
  // Health check
  http.get(`${DAEMON_URL}/api/v1/health`, () => {
    return HttpResponse.json(createMockHealthResponse());
  }),

  // Templates
  http.get(`${DAEMON_URL}/api/v1/templates`, () => {
    return HttpResponse.json(createMockTemplates());
  }),

  http.get(`${DAEMON_URL}/api/v1/templates/:name`, ({ params }) => {
    const templates = createMockTemplates();
    const template = templates.find((t) => t.Name === params.name);
    if (template) {
      return HttpResponse.json(template);
    }
    return HttpResponse.json(createMockError('Template not found', 404), { status: 404 });
  }),

  // Instances
  http.get(`${DAEMON_URL}/api/v1/instances`, () => {
    return HttpResponse.json(createMockInstances());
  }),

  http.get(`${DAEMON_URL}/api/v1/instances/:id`, ({ params }) => {
    const instances = createMockInstances();
    const instance = instances.find((i) => i.id === params.id);
    if (instance) {
      return HttpResponse.json(instance);
    }
    return HttpResponse.json(createMockError('Instance not found', 404), { status: 404 });
  }),

  http.post(`${DAEMON_URL}/api/v1/instances/launch`, async ({ request }) => {
    const body = await request.json() as any;
    return HttpResponse.json(createMockLaunchResponse());
  }),

  http.post(`${DAEMON_URL}/api/v1/instances/:id/stop`, () => {
    return HttpResponse.json({ status: 'stopping' });
  }),

  http.post(`${DAEMON_URL}/api/v1/instances/:id/start`, () => {
    return HttpResponse.json({ status: 'starting' });
  }),

  http.post(`${DAEMON_URL}/api/v1/instances/:id/terminate`, () => {
    return HttpResponse.json({ status: 'terminating' });
  }),

  http.post(`${DAEMON_URL}/api/v1/instances/:id/hibernate`, () => {
    return HttpResponse.json({ status: 'hibernating' });
  }),

  http.post(`${DAEMON_URL}/api/v1/instances/:id/resume`, () => {
    return HttpResponse.json({ status: 'resuming' });
  }),

  http.get(`${DAEMON_URL}/api/v1/instances/:id/hibernation-status`, () => {
    return HttpResponse.json(createMockHibernationStatus());
  }),

  http.get(`${DAEMON_URL}/api/v1/instances/:id/connection`, () => {
    return HttpResponse.json(createMockConnectionInfo());
  }),

  // Profiles
  http.get(`${DAEMON_URL}/api/v1/profiles`, () => {
    return HttpResponse.json(createMockProfiles());
  }),

  http.get(`${DAEMON_URL}/api/v1/profiles/current`, () => {
    const profiles = createMockProfiles();
    const defaultProfile = profiles.find((p) => p.is_default);
    return HttpResponse.json(defaultProfile || profiles[0]);
  }),

  http.post(`${DAEMON_URL}/api/v1/profiles`, async ({ request }) => {
    const body = await request.json() as any;
    return HttpResponse.json({ name: body.name, status: 'created' });
  }),

  http.put(`${DAEMON_URL}/api/v1/profiles/:name`, async ({ request, params }) => {
    const body = await request.json() as any;
    return HttpResponse.json({ name: params.name, status: 'updated' });
  }),

  http.delete(`${DAEMON_URL}/api/v1/profiles/:name`, () => {
    return HttpResponse.json({ status: 'deleted' });
  }),

  http.post(`${DAEMON_URL}/api/v1/profiles/:name/switch`, ({ params }) => {
    return HttpResponse.json({ current_profile: params.name });
  }),

  // Volumes (EFS)
  http.get(`${DAEMON_URL}/api/v1/volumes`, () => {
    return HttpResponse.json(createMockVolumes());
  }),

  http.get(`${DAEMON_URL}/api/v1/volumes/:id`, ({ params }) => {
    const volumes = createMockVolumes();
    const volume = volumes.find((v) => v.filesystem_id === params.id);
    if (volume) {
      return HttpResponse.json(volume);
    }
    return HttpResponse.json(createMockError('Volume not found', 404), { status: 404 });
  }),

  http.post(`${DAEMON_URL}/api/v1/volumes`, async ({ request }) => {
    const body = await request.json() as any;
    return HttpResponse.json({
      filesystem_id: 'fs-new123456789',
      name: body.name,
      status: 'creating',
    });
  }),

  http.delete(`${DAEMON_URL}/api/v1/volumes/:id`, () => {
    return HttpResponse.json({ status: 'deleting' });
  }),

  http.post(`${DAEMON_URL}/api/v1/volumes/:id/mount`, () => {
    return HttpResponse.json({ status: 'mounting' });
  }),

  http.post(`${DAEMON_URL}/api/v1/volumes/:id/unmount`, () => {
    return HttpResponse.json({ status: 'unmounting' });
  }),

  // EBS Storage
  http.get(`${DAEMON_URL}/api/v1/storage`, () => {
    return HttpResponse.json(createMockEBSStorages());
  }),

  http.get(`${DAEMON_URL}/api/v1/storage/:id`, ({ params }) => {
    const storages = createMockEBSStorages();
    const storage = storages.find((s) => s.volume_id === params.id);
    if (storage) {
      return HttpResponse.json(storage);
    }
    return HttpResponse.json(createMockError('Storage not found', 404), { status: 404 });
  }),

  http.post(`${DAEMON_URL}/api/v1/storage`, async ({ request }) => {
    const body = await request.json() as any;
    return HttpResponse.json({
      volume_id: 'vol-new123456789',
      name: body.name,
      status: 'creating',
    });
  }),

  http.delete(`${DAEMON_URL}/api/v1/storage/:id`, () => {
    return HttpResponse.json({ status: 'deleting' });
  }),

  http.post(`${DAEMON_URL}/api/v1/storage/:id/attach`, () => {
    return HttpResponse.json({ status: 'attaching' });
  }),

  http.post(`${DAEMON_URL}/api/v1/storage/:id/detach`, () => {
    return HttpResponse.json({ status: 'detaching' });
  }),

  // Idle Policies
  http.get(`${DAEMON_URL}/api/v1/idle/policies`, () => {
    return HttpResponse.json(createMockIdlePolicies());
  }),

  http.get(`${DAEMON_URL}/api/v1/idle/policies/:id`, ({ params }) => {
    const policies = createMockIdlePolicies();
    const policy = policies.find((p) => p.id === params.id);
    if (policy) {
      return HttpResponse.json(policy);
    }
    return HttpResponse.json(createMockError('Policy not found', 404), { status: 404 });
  }),

  http.post(`${DAEMON_URL}/api/v1/idle/policies`, async ({ request }) => {
    const body = await request.json() as any;
    return HttpResponse.json({
      id: `policy-${Date.now()}`,
      ...body,
      status: 'created',
    });
  }),

  http.put(`${DAEMON_URL}/api/v1/idle/policies/:id`, async ({ request, params }) => {
    const body = await request.json() as any;
    return HttpResponse.json({
      id: params.id,
      ...body,
      status: 'updated',
    });
  }),

  http.delete(`${DAEMON_URL}/api/v1/idle/policies/:id`, () => {
    return HttpResponse.json({ status: 'deleted' });
  }),

  http.post(`${DAEMON_URL}/api/v1/idle/instances/:id/apply-policy`, () => {
    return HttpResponse.json({ status: 'policy applied' });
  }),

  http.get(`${DAEMON_URL}/api/v1/idle/history`, () => {
    return HttpResponse.json([
      {
        instance_id: 'i-1234567890abcdef0',
        instance_name: 'my-ml-research',
        action: 'hibernate',
        reason: 'idle_timeout',
        timestamp: '2025-11-16T14:30:00Z',
        cost_saved: 2.40,
      },
    ]);
  }),

  // Backup & Snapshot
  http.get(`${DAEMON_URL}/api/v1/backups`, () => {
    return HttpResponse.json([
      {
        id: 'backup-123',
        instance_id: 'i-1234567890abcdef0',
        name: 'my-ml-research-backup',
        created_at: '2025-11-15T10:00:00Z',
        size_gb: 50,
        status: 'available',
      },
    ]);
  }),

  http.post(`${DAEMON_URL}/api/v1/backups`, async ({ request }) => {
    const body = await request.json() as any;
    return HttpResponse.json({
      id: `backup-${Date.now()}`,
      status: 'creating',
    });
  }),

  http.delete(`${DAEMON_URL}/api/v1/backups/:id`, () => {
    return HttpResponse.json({ status: 'deleting' });
  }),

  http.post(`${DAEMON_URL}/api/v1/backups/:id/restore`, () => {
    return HttpResponse.json({ status: 'restoring', instance_id: 'i-restored123' });
  }),
];

/**
 * Error handlers for testing error scenarios
 */
export const errorHandlers = [
  http.get(`${DAEMON_URL}/api/v1/templates`, () => {
    return HttpResponse.json(createMockError('Failed to fetch templates', 500), { status: 500 });
  }),

  http.post(`${DAEMON_URL}/api/v1/instances/launch`, () => {
    return HttpResponse.json(createMockError('Instance launch failed', 400), { status: 400 });
  }),

  http.get(`${DAEMON_URL}/api/v1/health`, () => {
    return HttpResponse.json(createMockError('Daemon unavailable', 503), { status: 503 });
  }),
];

/**
 * Empty handlers for testing empty states
 */
export const emptyHandlers = [
  http.get(`${DAEMON_URL}/api/v1/templates`, () => {
    return HttpResponse.json([]);
  }),

  http.get(`${DAEMON_URL}/api/v1/instances`, () => {
    return HttpResponse.json([]);
  }),

  http.get(`${DAEMON_URL}/api/v1/profiles`, () => {
    return HttpResponse.json([]);
  }),

  http.get(`${DAEMON_URL}/api/v1/volumes`, () => {
    return HttpResponse.json([]);
  }),

  http.get(`${DAEMON_URL}/api/v1/storage`, () => {
    return HttpResponse.json([]);
  }),
];
