/**
 * ProfileManager.test.tsx
 *
 * Tests the Profile Management functionality in the Prism GUI App.
 * Profiles are accessible via Settings > Profiles (settingsSection === 'profiles').
 *
 * Navigation: Click "Settings" in nav → then click "Profiles" in Settings sub-nav.
 *
 * The ProfileSelectorView shows:
 * - data-testid="profiles-table"
 * - data-testid="create-profile-button" → "Create Profile"
 * - data-testid="switch-profile-${name}" per non-current profile
 * - data-testid="edit-profile-${name}" per profile
 * - data-testid="delete-profile-${name}" per non-current profile
 * - "No profiles" empty state
 * - "Profile Management" header
 *
 * Uses vi.stubGlobal('fetch', ...) to mock SafePrismAPI HTTP calls.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import App from './App';

// ── Mock Data ─────────────────────────────────────────────────────────────

const mockProfiles = [
  {
    id: 'prof-default',
    name: 'default',
    aws_profile: 'default',
    region: 'us-west-2',
    type: 'standard',
    default: true,
  },
  {
    id: 'prof-research',
    name: 'research-profile',
    aws_profile: 'research',
    region: 'us-east-1',
    type: 'standard',
    default: false,
  },
];

function buildFetchMock(overrides?: { profiles?: unknown[]; failProfiles?: boolean }) {
  return vi.fn().mockImplementation((url: string) => {
    if (url.includes('/api/v1/profiles')) {
      if (overrides?.failProfiles) {
        return Promise.reject(new Error('Failed to fetch profiles'));
      }
      const profiles = overrides?.profiles ?? mockProfiles;
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve(profiles) });
    }
    if (url.includes('/api/v1/templates')) {
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({}) });
    }
    if (url.includes('/api/v1/instances')) {
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({ instances: [] }) });
    }
    if (url.includes('/api/v1/snapshots')) {
      return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({ snapshots: [], count: 0 }) });
    }
    return Promise.resolve({ ok: true, status: 200, headers: { get: () => null }, json: () => Promise.resolve({}) });
  });
}

// ── Setup ─────────────────────────────────────────────────────────────────

beforeEach(() => {
  vi.stubGlobal('fetch', buildFetchMock());
  localStorage.setItem('cws_onboarding_complete', 'true');
});

afterEach(() => {
  vi.unstubAllGlobals();
  localStorage.removeItem('cws_onboarding_complete');
});

// ── Helpers ────────────────────────────────────────────────────────────────

async function navigateToProfiles() {
  const user = userEvent.setup();
  render(<App />);

  // Navigate to Settings
  const settingsLinks = screen.getAllByText('Settings');
  await user.click(settingsLinks[0]);

  // Navigate to Profiles sub-section
  await waitFor(() => {
    const profilesLinks = screen.getAllByText('Profiles');
    expect(profilesLinks.length).toBeGreaterThan(0);
  });

  const profilesLinks = screen.getAllByText('Profiles');
  await user.click(profilesLinks[0]);

  await waitFor(() => {
    expect(screen.getByText('Profile Management')).toBeInTheDocument();
  }, { timeout: 5000 });

  return user;
}

// ── Tests ─────────────────────────────────────────────────────────────────

describe('ProfileManager', () => {
  describe('Navigation', () => {
    it('renders the app with Settings nav item', () => {
      render(<App />);
      expect(screen.getAllByText('Settings').length).toBeGreaterThan(0);
    });

    it('can navigate to Settings view', async () => {
      const user = userEvent.setup();
      render(<App />);

      const settingsLinks = screen.getAllByText('Settings');
      await user.click(settingsLinks[0]);

      await waitFor(() => {
        // Settings view shows General Settings by default
        const content = screen.queryByText('General Settings') ||
                        screen.queryByText('Profile Management') ||
                        screen.queryByText('System Status');
        expect(content).toBeTruthy();
      });
    });

    it('can navigate to Profiles section within Settings', async () => {
      await navigateToProfiles();
    });

    it('shows Profile Management header in Profiles section', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        expect(screen.getByText('Profile Management')).toBeInTheDocument();
      });
    });
  });

  describe('Profile List Rendering', () => {
    it('should show the profiles table', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        expect(screen.getByTestId('profiles-table')).toBeInTheDocument();
      });
    });

    it('should display profile names', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        // 'default' may appear multiple times (name + aws_profile + nav items)
        expect(screen.getAllByText('default').length).toBeGreaterThan(0);
        expect(screen.getAllByText('research-profile').length).toBeGreaterThan(0);
      });
    });

    it('should display profile regions', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        // Region column shows 'us-west-2' and 'us-east-1'
        expect(screen.getAllByText('us-west-2').length).toBeGreaterThan(0);
        expect(screen.getAllByText('us-east-1').length).toBeGreaterThan(0);
      });
    });

    it('should display AWS profile names', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        // aws_profile column shows 'default' and 'research'
        expect(screen.getAllByText('default').length).toBeGreaterThan(0);
        expect(screen.getAllByText('research').length).toBeGreaterThan(0);
      });
    });

    it('should show profile counter', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        expect(screen.getByText('(2 profiles)')).toBeInTheDocument();
      });
    });

    it('should handle empty profile list', async () => {
      vi.stubGlobal('fetch', buildFetchMock({ profiles: [] }));

      const user = userEvent.setup();
      render(<App />);

      const settingsLinks = screen.getAllByText('Settings');
      await user.click(settingsLinks[0]);

      const profilesLinks = screen.getAllByText('Profiles');
      await user.click(profilesLinks[0]);

      await waitFor(() => {
        expect(screen.getByText('No profiles')).toBeInTheDocument();
      });
    });
  });

  describe('Profile Actions', () => {
    it('should show Create Profile button', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        expect(screen.getByTestId('create-profile-button')).toBeInTheDocument();
      });
    });

    it('should show Edit button for each profile', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        expect(screen.getByTestId('edit-profile-default')).toBeInTheDocument();
        expect(screen.getByTestId('edit-profile-research-profile')).toBeInTheDocument();
      });
    });

    it('should show Switch button for non-current profile', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        expect(screen.getByTestId('switch-profile-research-profile')).toBeInTheDocument();
      });
    });

    it('should show Delete button for non-current profile', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        expect(screen.getByTestId('delete-profile-research-profile')).toBeInTheDocument();
      });
    });

    it('should not show Switch button for current profile', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        // default is current (default: true) → no switch button
        expect(screen.queryByTestId('switch-profile-default')).not.toBeInTheDocument();
      });
    });
  });

  describe('Create Profile Modal', () => {
    it('should open Create Profile modal when button is clicked', async () => {
      const user = await navigateToProfiles();

      const createButton = screen.getByTestId('create-profile-button');
      await user.click(createButton);

      await waitFor(() => {
        // Modal header "Create Profile"
        const headers = screen.getAllByText('Create Profile');
        expect(headers.length).toBeGreaterThan(0);
      });
    });

    it('should show Profile Name field in modal', async () => {
      const user = await navigateToProfiles();

      const createButton = screen.getByTestId('create-profile-button');
      await user.click(createButton);

      await waitFor(() => {
        // "Profile Name" appears in table header AND in modal label
        expect(screen.getAllByText('Profile Name').length).toBeGreaterThan(0);
      });
    });

    it('should show validation error when creating without name', async () => {
      const user = await navigateToProfiles();

      const createButton = screen.getByTestId('create-profile-button');
      await user.click(createButton);

      // Modal opens — wait for it to appear
      await waitFor(() => {
        const headers = screen.getAllByText('Create Profile');
        expect(headers.length).toBeGreaterThan(0);
      });

      // The modal is now open. Find all buttons with text "Create" —
      // the last visible one should be the modal's primary submit button.
      // Click all of them until validation error appears.
      const allSpans = screen.getAllByText('Create');
      // Use the last one which is the primary button in the modal footer
      await user.click(allSpans[allSpans.length - 1]);

      // After clicking, validation should fire since name is empty.
      // The validation error is shown via data-testid="validation-error"
      await waitFor(
        () => {
          const validationEl = screen.queryByTestId('validation-error');
          const validationText = screen.queryByText(/Profile name is required/);
          expect(validationEl || validationText).not.toBeNull();
        },
        { timeout: 3000 }
      );
    });

    it('should call profiles API when creating a profile', async () => {
      const fetchSpy = buildFetchMock();
      vi.stubGlobal('fetch', fetchSpy);

      const user = userEvent.setup();
      render(<App />);

      const settingsLinks = screen.getAllByText('Settings');
      await user.click(settingsLinks[0]);
      const profilesLinks = screen.getAllByText('Profiles');
      await user.click(profilesLinks[0]);

      await waitFor(() => {
        expect(screen.getByTestId('create-profile-button')).toBeInTheDocument();
      });

      const createButton = screen.getByTestId('create-profile-button');
      await user.click(createButton);

      await waitFor(() => {
        const profileNameInput = screen.queryByTestId('profile-name-input');
        expect(profileNameInput).toBeTruthy();
      });

      // Fill in the form using testid
      const nameInput = screen.getByTestId('profile-name-input');
      await user.type(nameInput, 'new-profile');

      // Click Create
      const createButtons = screen.getAllByText('Create');
      await user.click(createButtons[createButtons.length - 1]);

      await waitFor(() => {
        const calls = (fetchSpy as ReturnType<typeof vi.fn>).mock.calls.map((c: unknown[]) => c[0] as string);
        expect(calls.some((url: string) => url.includes('/api/v1/profiles'))).toBe(true);
      });
    });
  });

  describe('Edit Profile', () => {
    it('should open edit modal when Edit button is clicked', async () => {
      const user = await navigateToProfiles();

      await waitFor(() => {
        expect(screen.getByTestId('edit-profile-research-profile')).toBeInTheDocument();
      });

      const editButton = screen.getByTestId('edit-profile-research-profile');
      await user.click(editButton);

      await waitFor(() => {
        // Edit Profile modal header
        expect(screen.getByText('Edit Profile')).toBeInTheDocument();
      });
    });
  });

  describe('Switch Profile', () => {
    it('should call activate API when Switch is clicked', async () => {
      const fetchSpy = buildFetchMock();
      vi.stubGlobal('fetch', fetchSpy);

      const user = userEvent.setup();
      render(<App />);

      const settingsLinks = screen.getAllByText('Settings');
      await user.click(settingsLinks[0]);
      const profilesLinks = screen.getAllByText('Profiles');
      await user.click(profilesLinks[0]);

      await waitFor(() => {
        expect(screen.getByTestId('switch-profile-research-profile')).toBeInTheDocument();
      });

      const switchButton = screen.getByTestId('switch-profile-research-profile');
      await user.click(switchButton);

      await waitFor(() => {
        const calls = (fetchSpy as ReturnType<typeof vi.fn>).mock.calls.map((c: unknown[]) => c[0] as string);
        expect(calls.some((url: string) => url.includes('/api/v1/profiles') && url.includes('activate'))).toBe(true);
      });
    });
  });

  describe('Delete Profile', () => {
    it('should show delete confirmation modal when Delete is clicked', async () => {
      const user = await navigateToProfiles();

      await waitFor(() => {
        expect(screen.getByTestId('delete-profile-research-profile')).toBeInTheDocument();
      });

      const deleteButton = screen.getByTestId('delete-profile-research-profile');
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByText('Delete Profile')).toBeInTheDocument();
      });
    });
  });

  describe('API Integration', () => {
    it('calls /api/v1/profiles on mount', async () => {
      const fetchSpy = buildFetchMock();
      vi.stubGlobal('fetch', fetchSpy);

      render(<App />);

      // Navigate to profiles to trigger load
      const user = userEvent.setup();
      const settingsLinks = screen.getAllByText('Settings');
      await user.click(settingsLinks[0]);
      const profilesLinks = screen.getAllByText('Profiles');
      await user.click(profilesLinks[0]);

      await waitFor(() => {
        const calls = (fetchSpy as ReturnType<typeof vi.fn>).mock.calls.map((c: unknown[]) => c[0] as string);
        expect(calls.some((url: string) => url.includes('/api/v1/profiles'))).toBe(true);
      });
    });

    it('handles profiles API failure gracefully', async () => {
      vi.stubGlobal('fetch', buildFetchMock({ failProfiles: true }));

      userEvent.setup();
      render(<App />);

      // App should still render
      expect(screen.getAllByText('Settings').length).toBeGreaterThan(0);
    });

    it('shows Refresh button in profiles view', async () => {
      await navigateToProfiles();

      await waitFor(() => {
        expect(screen.getByText('Refresh')).toBeInTheDocument();
      });
    });
  });
});
