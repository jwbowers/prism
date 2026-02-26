// Global teardown for Playwright tests
// Runs after all tests complete.
// Cleans up any test storage volumes that were not removed by individual tests
// (e.g., when tests timeout, fail, or the shared setup volumes are never deleted).
//
// This prevents zombie AWS resources from accumulating across test runs,
// which wastes money and causes subsequent test runs to fail with stale state.

import { stopDaemon, isDaemonRunning, cleanupTestStorage } from './setup-daemon.js'
import { exec } from 'child_process'
import { promisify } from 'util'

const execAsync = promisify(exec)

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

  // Stop the test daemon
  try {
    const { stdout } = await execAsync('pgrep -f prismd || true')
    const pids = stdout.trim().split('\n').filter(Boolean)
    if (pids.length > 0) {
      console.log(`[teardown] Stopping daemon processes: ${pids.join(', ')}`)
      await execAsync('pkill -SIGTERM prismd || true')
      // Give it a few seconds for graceful shutdown
      await new Promise(resolve => setTimeout(resolve, 3000))
      // Force kill if still running
      await execAsync('pkill -9 prismd || true').catch(() => {})
      console.log('[teardown] Daemon stopped')
    }
  } catch (error) {
    // Non-critical — process may have already stopped
  }

  console.log('[teardown] Post-test cleanup complete')
}

export default globalTeardown
