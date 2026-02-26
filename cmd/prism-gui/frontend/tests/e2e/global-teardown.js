// Global teardown for Playwright tests
// Runs after all tests complete.
// Cleans up any test storage volumes that were not removed by individual tests
// (e.g., when tests timeout, fail, or the shared setup volumes are never deleted).
//
// This prevents zombie AWS resources from accumulating across test runs,
// which wastes money and causes subsequent test runs to fail with stale state.
//
// NOTE: Daemon lifecycle is managed by global-setup.js (it returns a teardown
// function that stops the daemon it started). This file must NOT kill the daemon —
// doing so would terminate a co-running tier (e.g., killing serial's daemon when
// fast tier finishes after ~10 minutes).

import { isDaemonRunning, cleanupTestStorage, checkZombieInstances } from './setup-daemon.js'

async function globalTeardown() {
  console.log('[teardown] Running post-test cleanup...')

  // Clean up any remaining test storage volumes.
  // This catches the shared volumes (test-setup-efs, test-setup-ebs) which are never
  // deleted by individual tests, plus any volumes from tests that timed out.
  if (await isDaemonRunning()) {
    await cleanupTestStorage()
  } else {
    console.log('[teardown] Daemon not running - skipping storage cleanup (already cleaned up)')
  }

  // Check for zombie Prism-managed EC2 instances left running in AWS.
  // Auto-terminates test-pattern instances older than 2h; warns about others.
  await checkZombieInstances()

  console.log('[teardown] Post-test cleanup complete')
}

export default globalTeardown
