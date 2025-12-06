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
    await execAsync('pkill -9 prismd || true')
    // Wait for processes to fully terminate
    await new Promise(resolve => setTimeout(resolve, 1000))
  } catch (error) {
    // Ignore errors - process might not exist
  }

  // Clean up PID files to prevent singleton takeover timeout
  try {
    console.log('Cleaning up PID lock files...')
    await execAsync('rm -f ~/.prism/daemon.pid || true')
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