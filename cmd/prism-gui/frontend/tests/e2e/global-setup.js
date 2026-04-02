// Global setup for Playwright tests
import { startDaemon, stopDaemon, isDaemonRunning, cleanupTestUsers, cleanupTestProjects, cleanupTestStorage, cleanupTestCourses, createGovernanceTestProject } from './setup-daemon.js'
import { exec } from 'child_process'
import { promisify } from 'util'

const execAsync = promisify(exec)

let daemonPid

async function globalSetup() {
  // ALWAYS kill existing daemons to ensure clean test environment
  // This prevents tests from connecting to a production daemon without PRISM_TEST_MODE
  try {
    console.log('Killing any existing daemon processes...')

    // Kill aggressively - multiple rounds to catch any stubborn processes
    for (let i = 0; i < 3; i++) {
      await execAsync('pkill -9 prismd || true')
      await new Promise(resolve => setTimeout(resolve, 500))
    }

    // Verify all daemons are dead
    let attempts = 0
    while (attempts < 10) {
      const { stdout } = await execAsync('pgrep -f prismd || true')
      if (!stdout.trim()) {
        console.log('✓ All daemon processes terminated')
        break
      }
      console.log(`Waiting for daemon processes to terminate... (${stdout.trim()})`)
      await new Promise(resolve => setTimeout(resolve, 1000))
      attempts++

      if (attempts === 10) {
        console.warn('Warning: Some daemon processes may still be running')
      }
    }

    // Additional wait for ports to be released
    // (OS needs time to clean up TCP connections and release port 8947)
    await new Promise(resolve => setTimeout(resolve, 2000))
  } catch (error) {
    // Ignore errors - process might not exist
  }

  // Clean up daemon state files to prevent singleton takeover timeout
  // The daemon uses multiple methods to detect existing daemons:
  // 1. PID file: ~/.prism/prismd.pid (see pkg/daemon/singleton.go:20)
  // 2. Registry file: ~/.prism/daemon_registry.json (see pkg/daemon/process_manager.go:94)
  // 3. System-wide process scan: ps/pgrep for prismd processes
  try {
    console.log('Cleaning up daemon state files...')
    await execAsync('rm -f ~/.prism/prismd.pid ~/.prism/daemon_registry.json || true')
  } catch (error) {
    // Ignore errors - files might not exist
  }

  // Start a fresh daemon with test mode enabled
  console.log('Starting daemon for tests with PRISM_TEST_MODE...')
  daemonPid = await startDaemon()

  // Brief pause to allow the daemon's HTTP handler registration to fully complete.
  // The ping endpoint responds quickly, but project/storage handlers may need
  // a moment to finish initializing in-memory data structures.
  await new Promise(resolve => setTimeout(resolve, 500))

  // Remove test users left over from previous runs before tests start.
  // Prevents cleanupTestUsers() helpers from overflowing when 100+ users accumulate.
  await cleanupTestUsers()

  // Remove test projects left over from previous runs before tests start.
  // Cloudscape Table only renders the current page (~20 items), so accumulated test projects
  // (e.g., 600+) push newly-created projects off page 1, making row lookups fail.
  await cleanupTestProjects()

  // Create the shared governance E2E test project so it exists for the entire run.
  // governance-workflows.spec.ts tests look up this project by name; having it pre-created
  // avoids beforeAll/afterAll lifecycle issues when Playwright re-evaluates the spec per describe group.
  await createGovernanceTestProject()

  // Remove test courses left over from previous runs.
  // Prevents strict-mode violations: openCourse(prefix) finds multiple rows when stale
  // courses from interrupted runs share the same code prefix (e.g. PWR-TA-12345).
  await cleanupTestCourses()

  // Remove test storage volumes (EFS + EBS) left over from previous runs.
  // Stale volumes in transitional states (creating, deleting) cause storage test failures:
  //   - Delete button disabled (aria-disabled="true") for volumes still being created/deleted
  //   - waitForEFSVolumeToExist times out when previous run's volume row is missing
  // Cleanup ensures each run starts with a known-clean set of storage resources.
  await cleanupTestStorage()

  // Return teardown function
  return async () => {
    if (daemonPid) {
      console.log('Stopping daemon after tests...')
      await stopDaemon(daemonPid)
    }
  }
}

export default globalSetup