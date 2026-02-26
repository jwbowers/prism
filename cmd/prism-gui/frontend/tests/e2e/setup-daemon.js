// Setup script to start the actual daemon for E2E testing
import { exec, spawn } from 'child_process'
import { promisify } from 'util'
import path from 'path'
import fs from 'fs'

const execAsync = promisify(exec)

// Function to check if daemon is running
async function isDaemonRunning() {
  try {
    const response = await fetch('http://localhost:8947/api/v1/ping')
    return response.ok
  } catch {
    return false
  }
}

// Function to start the daemon
async function startDaemon() {
  const daemonPath = path.join(process.cwd(), '..', '..', '..', 'bin', 'prismd')

  // Check if daemon binary exists and is up-to-date
  let needsBuild = !fs.existsSync(daemonPath)

  if (!needsBuild) {
    // Rebuild if any Go source file is newer than the binary
    const binaryMtime = fs.statSync(daemonPath).mtimeMs
    try {
      const { stdout } = await execAsync(
        `find ../../../pkg ../../../cmd/prismd ../../../internal -name "*.go" -newer "${daemonPath}" 2>/dev/null | head -1`,
        { cwd: process.cwd() }
      )
      if (stdout.trim()) {
        console.log(`Go source files changed (e.g. ${stdout.trim()}), rebuilding daemon...`)
        needsBuild = true
      }
    } catch {
      // find failed - skip stale check, use existing binary
    }
  } else {
    console.error(`Daemon binary not found at ${daemonPath}`)
    console.log('Building daemon...')
  }

  if (needsBuild) {
    // Build the daemon
    const buildCmd = 'cd ../../.. && go build -o bin/prismd ./cmd/prismd'
    await execAsync(buildCmd)

    if (!fs.existsSync(daemonPath)) {
      throw new Error('Failed to build daemon')
    }
    console.log('Daemon binary built successfully')
  }
  
  // Start daemon in background
  console.log('Starting Prism daemon for testing...')

  // Calculate absolute path to templates directory (repository root + /templates)
  const repoRoot = path.join(process.cwd(), '..', '..', '..')
  const templatesPath = path.join(repoRoot, 'templates')

  console.log(`Setting PRISM_TEMPLATE_DIR=${templatesPath}`)

  // Check if LocalStack mode is enabled
  const useLocalStack = process.env.PRISM_USE_LOCALSTACK === 'true'

  if (useLocalStack) {
    console.log('🚀 LocalStack mode enabled - tests will use local AWS emulation')
  } else {
    console.log('☁️  Real AWS mode - tests will use actual AWS resources')
  }

  const daemon = spawn(daemonPath, [], {
    detached: true,
    stdio: ['ignore', 'pipe', 'pipe'],
    env: {
      ...process.env,
      PRISM_TEST_MODE: 'true',
      PRISM_TEMPLATE_DIR: templatesPath,
      // LocalStack configuration (pkg/aws/localstack/config.go handles endpoint setup)
      PRISM_USE_LOCALSTACK: useLocalStack ? 'true' : undefined,
      // AWS configuration (only set if NOT using LocalStack)
      AWS_PROFILE: useLocalStack ? undefined : 'aws',
      AWS_REGION: 'us-west-2',
      AWS_SDK_LOAD_CONFIG: useLocalStack ? undefined : '1',
      AWS_EC2_METADATA_DISABLED: 'true'
    }
  })
  
  // Log daemon output for debugging
  daemon.stdout.on('data', (data) => {
    console.log(`[Daemon] ${data.toString()}`)
  })
  
  daemon.stderr.on('data', (data) => {
    console.error(`[Daemon Error] ${data.toString()}`)
  })
  
  daemon.unref()
  
  // Wait for daemon to be ready
  // NOTE: Daemon may take up to 20 seconds for singleton takeover if old process exists
  // (10s graceful shutdown + 5s SIGINT + 5s SIGKILL)
  // Allow 60 seconds total: 20s for takeover + 10s for startup + 30s buffer
  let attempts = 0
  const maxAttempts = 60
  while (attempts < maxAttempts) {
    if (await isDaemonRunning()) {
      console.log(`Daemon is ready! (took ${attempts} seconds)`)
      return daemon.pid
    }
    await new Promise(resolve => setTimeout(resolve, 1000))
    attempts++

    // Progress indicator every 5 seconds
    if (attempts % 5 === 0) {
      console.log(`Still waiting for daemon... (${attempts}/${maxAttempts}s)`)
    }
  }

  throw new Error(`Daemon failed to start within ${maxAttempts} seconds. Check daemon logs above for errors.`)
}

// Delete test users that accumulated from previous test runs.
// Runs after daemon starts to keep the user table clean so cleanup helpers don't overflow.
async function cleanupTestUsers() {
  const testUserPattern = /^(test-|status-update-test-|role-test-|list-test-|delete-test-|update-test-|bulk-|deactivate-|reactivate-|invite-|workspace-test-)/

  try {
    const res = await fetch('http://localhost:8947/api/v1/users')
    if (!res.ok) return

    const users = await res.json()
    if (!Array.isArray(users)) return

    const testUsers = users.filter(u => u.username && testUserPattern.test(u.username))
    if (testUsers.length === 0) return

    console.log(`[setup-daemon] Cleaning up ${testUsers.length} leftover test users...`)
    let deleted = 0
    for (const user of testUsers) {
      const delRes = await fetch(`http://localhost:8947/api/v1/users/${user.username}`, {
        method: 'DELETE'
      }).catch(() => null)
      if (delRes && (delRes.status === 204 || delRes.status === 404)) {
        deleted++
      }
    }
    console.log(`[setup-daemon] Removed ${deleted}/${testUsers.length} test users`)
  } catch {
    // Non-critical — tests can still run if cleanup fails
  }
}

// Delete test storage volumes (EFS + EBS) that accumulated from previous test runs.
// Runs after daemon starts to ensure a clean slate before storage tests begin.
// Prevents failures caused by stale volumes in transitional states (creating, deleting, etc.)
//
// API routes:
//   GET  /api/v1/storage         → list ALL volumes (EFS + EBS)
//   DELETE /api/v1/volumes/{name} → delete EFS (shared) volume
//   DELETE /api/v1/storage/{name} → delete EBS (workspace) volume
async function cleanupTestStorage() {
  const testVolumePattern = /^(test-|delete-test-|mount-test-|unmount-test-|attach-test-|detach-test-|test-setup-|test-efs-|test-ebs-)/

  try {
    const res = await fetch('http://localhost:8947/api/v1/storage')
    if (!res.ok) return

    const volumes = await res.json()
    if (!Array.isArray(volumes)) return

    const testVolumes = volumes.filter(v => v.name && testVolumePattern.test(v.name))
    if (testVolumes.length === 0) {
      console.log('[setup-daemon] No leftover test storage volumes found')
      return
    }

    console.log(`[setup-daemon] Cleaning up ${testVolumes.length} leftover test storage volumes...`)
    let deleted = 0
    for (const vol of testVolumes) {
      // Route to correct endpoint based on AWS service type:
      // EFS (shared) → /api/v1/volumes/{name}
      // EBS (workspace) → /api/v1/storage/{name}
      const isEFS = vol.aws_service === 'efs' || vol.type === 'shared'
      const endpoint = isEFS
        ? `http://localhost:8947/api/v1/volumes/${encodeURIComponent(vol.name)}`
        : `http://localhost:8947/api/v1/storage/${encodeURIComponent(vol.name)}`

      const delRes = await fetch(endpoint, { method: 'DELETE' }).catch(() => null)
      if (delRes && (delRes.status === 204 || delRes.status === 200 || delRes.status === 404)) {
        deleted++
        console.log(`[setup-daemon]   Deleted ${isEFS ? 'EFS' : 'EBS'}: ${vol.name}`)
      } else if (delRes) {
        console.log(`[setup-daemon]   Skipped ${vol.name} (status ${delRes.status} - may be in transitional state)`)
      }
    }
    console.log(`[setup-daemon] Storage cleanup: removed ${deleted}/${testVolumes.length} test volumes`)
  } catch (e) {
    // Non-critical — tests can still run if cleanup fails
    console.log(`[setup-daemon] Storage cleanup skipped: ${e.message}`)
  }
}

// Function to stop the daemon
async function stopDaemon(pid) {
  if (pid) {
    try {
      process.kill(pid, 'SIGTERM')
      console.log(`Stopped daemon with PID ${pid}`)
    } catch (error) {
      console.error(`Failed to stop daemon: ${error.message}`)
    }
  }
}

export { startDaemon, stopDaemon, isDaemonRunning, cleanupTestUsers, cleanupTestStorage }