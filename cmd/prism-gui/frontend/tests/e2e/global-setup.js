// Global setup for Playwright tests
import { startDaemon, stopDaemon, isDaemonRunning } from './setup-daemon.js'
import { exec } from 'child_process'
import { promisify } from 'util'

const execAsync = promisify(exec)

let daemonPid

async function globalSetup() {
  // Defense layer: Check for running playwright processes (catches direct npx playwright invocations)
  try {
    const { stdout } = await execAsync('pgrep -f "playwright test" || true')
    if (stdout.trim()) {
      console.error('❌ ERROR: Another Playwright test is already running!')
      console.error('   Process IDs:', stdout.trim())
      console.error('   Use "npm run test:e2e" instead of "npx playwright test" to enforce single-test execution.')
      throw new Error('Another test suite is already running')
    }
  } catch (error) {
    if (error.message.includes('already running')) {
      throw error
    }
    // pgrep command failed or no processes found - continue
  }
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
  
  // Return teardown function
  return async () => {
    if (daemonPid) {
      console.log('Stopping daemon after tests...')
      await stopDaemon(daemonPid)
    }
  }
}

export default globalSetup