// Setup script to start the actual daemon for E2E testing
import { exec, execSync, spawn } from 'child_process'
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

// Delete test projects that accumulated from previous test runs.
// Runs after daemon starts to keep the project table clean so pagination doesn't hide
// newly-created test projects (Cloudscape Table only renders the current page).
async function cleanupTestProjects() {
  const testProjectPatterns = [
    /^list-test-/,
    /^active-project-/,
    /^suspended-project-/,
    /^cancel-delete-test-/,
    /^delete-test-/,
    /^test-project-/,
    /^view-test-/,
    /^budget-view-test-/,
    /^budget-test-/,
    /^duplicate-test-/,
    /^spend-test-/,
    /^alert-test-/,
    /^exceeded-test-/,
    /^active-delete-test-/,
    /^gov-test-/,           // includes gov-test-e2e-shared
    // Invitation workflow test projects (created with timestamps appended)
    /^Membership Test /,
    /^Dialog Test /,
    /^Stats Test /,
    /^Expiration Test /,
    /^Email Validation Test /,
    /^Welcome Message Test /,
    /^Decline Test /,
    /^Decline Dialog Test /,
    /^Results Summary Test /,
    /^Decline Reason Test /,
    /^test-collab-project/,
    /^test-invitation-project/,
  ]

  try {
    const res = await fetch('http://localhost:8947/api/v1/projects')
    if (!res.ok) return

    // Guard against empty/invalid JSON during daemon startup
    const text = await res.text()
    if (!text || !text.trim()) return
    const data = JSON.parse(text)
    const projects = data.projects || []

    const testProjects = projects.filter(p =>
      p.name && testProjectPatterns.some(pat => pat.test(p.name))
    )

    if (testProjects.length === 0) {
      console.log('[setup-daemon] No leftover test projects found')
      return
    }

    console.log(`[setup-daemon] Cleaning up ${testProjects.length} leftover test projects...`)
    let deleted = 0
    for (const project of testProjects) {
      const delRes = await fetch(`http://localhost:8947/api/v1/projects/${project.id}`, {
        method: 'DELETE'
      }).catch(() => null)
      if (delRes && (delRes.status === 204 || delRes.status === 200 || delRes.status === 404)) {
        deleted++
      }
    }
    console.log(`[setup-daemon] Project cleanup: removed ${deleted}/${testProjects.length} test projects`)
  } catch (e) {
    // Non-critical — tests can still run if cleanup fails
    console.log(`[setup-daemon] Project cleanup skipped: ${e.message}`)
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

// Check for zombie Prism-managed EC2 instances left running in AWS after the test suite.
//
// Strategy:
//   - AUTO-TERMINATE: instances with prism:managed=true, older than ZOMBIE_THRESHOLD_HOURS,
//     whose name matches a test pattern (clearly temporary test resources)
//   - WARN ONLY: instances with prism:managed=true, older than ZOMBIE_THRESHOLD_HOURS,
//     whose name doesn't match a test pattern (may be intentional — don't auto-kill)
//
// Uses the same AWS_PROFILE and AWS_REGION as the daemon so credentials are consistent.
// Non-critical: failures are logged but do not fail the test run.
async function checkZombieInstances() {
  // Patterns that identify clearly-temporary test instances (auto-terminate)
  const TEST_NAME_PATTERN = /^(test-|backup-test-|target-instance-|python-ml-test-|simple-|final-test|context-fix|collab-.*-test|getinstance-|no-refresh-|.*-debug.*|.*-benchmark.*)/

  // Instances must be older than this to be considered zombies
  // (2h safely covers the longest full test run of ~50min with headroom)
  const ZOMBIE_THRESHOLD_HOURS = 2

  const awsProfile = process.env.AWS_PROFILE || 'aws'
  const awsRegion  = process.env.AWS_REGION  || 'us-west-2'

  console.log(`[teardown] Checking for zombie Prism instances (AWS_PROFILE=${awsProfile}, region=${awsRegion})...`)

  try {
    const raw = execSync(
      `aws ec2 describe-instances \
        --region ${awsRegion} \
        --filters "Name=instance-state-name,Values=running" \
                  "Name=tag:prism:managed,Values=true" \
        --query "Reservations[*].Instances[*].[InstanceId,LaunchTime,Tags[?Key=='Name']|[0].Value,InstanceType]" \
        --output json`,
      {
        encoding: 'utf8',
        env: { ...process.env, AWS_PROFILE: awsProfile, AWS_REGION: awsRegion },
        stdio: ['pipe', 'pipe', 'pipe'],
      }
    )

    const instances = JSON.parse(raw).flat()
    if (instances.length === 0) {
      console.log('[teardown] No zombie Prism instances found ✅')
      return
    }

    const now = Date.now()
    const thresholdMs = ZOMBIE_THRESHOLD_HOURS * 60 * 60 * 1000

    const toTerminate = []
    const toWarn = []

    for (const [id, launchTime, name, type] of instances) {
      const ageMs = now - new Date(launchTime).getTime()
      if (ageMs < thresholdMs) continue  // recent — skip

      const ageHours = (ageMs / 3600000).toFixed(1)
      if (name && TEST_NAME_PATTERN.test(name)) {
        toTerminate.push({ id, name, type, ageHours })
      } else {
        toWarn.push({ id, name, type, ageHours })
      }
    }

    if (toTerminate.length > 0) {
      const ids = toTerminate.map(i => i.id).join(' ')
      console.log(`[teardown] ⚠️  Terminating ${toTerminate.length} zombie test instance(s)...`)
      execSync(
        `aws ec2 terminate-instances --region ${awsRegion} --instance-ids ${ids} --output json`,
        {
          encoding: 'utf8',
          env: { ...process.env, AWS_PROFILE: awsProfile, AWS_REGION: awsRegion },
          stdio: ['pipe', 'pipe', 'pipe'],
        }
      )
      for (const { id, name, type, ageHours } of toTerminate) {
        console.log(`[teardown]   Terminated: ${name || id} (${id}, ${type}, ${ageHours}h old)`)
      }
    }

    if (toWarn.length > 0) {
      console.log(`[teardown] ⚠️  ${toWarn.length} Prism instance(s) still running — review manually:`)
      for (const { id, name, type, ageHours } of toWarn) {
        console.log(`[teardown]   ${name || '(unnamed)'} (${id}, ${type}, ${ageHours}h old)`)
      }
    }

    if (toTerminate.length === 0 && toWarn.length === 0) {
      console.log('[teardown] No zombie Prism instances found ✅')
    }
  } catch (e) {
    // Non-critical — don't fail the test run over a zombie check
    console.log(`[teardown] Zombie instance check skipped: ${e.message}`)
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

// Create the shared governance E2E test project.
// Called from global-setup.js after cleanupTestProjects() so the project
// exists for the ENTIRE test run without per-describe beforeAll/afterAll lifecycle issues.
async function createGovernanceTestProject() {
  const projectName = 'gov-test-e2e-shared'

  // Helper: parse JSON response safely, returns null on empty/invalid
  const safeJson = async (res) => {
    const text = await res.text()
    if (!text || !text.trim()) return null
    try { return JSON.parse(text) } catch { return null }
  }

  // Helper: list projects with retry (daemon may return empty briefly at startup)
  const listProjects = async (retries = 5, delayMs = 400) => {
    for (let i = 0; i < retries; i++) {
      const res = await fetch('http://localhost:8947/api/v1/projects')
      if (res.ok) {
        const data = await safeJson(res)
        if (data) return data.projects || []
      }
      if (i < retries - 1) await new Promise(r => setTimeout(r, delayMs))
    }
    return null
  }

  try {
    // Check if it already exists (persisted from previous run)
    const projects = await listProjects()
    if (projects) {
      const existing = projects.find(p => p.name === projectName)
      if (existing) {
        console.log(`[setup-daemon] Governance test project already exists: ${existing.id}`)
        return
      }
    }

    // Create new project
    const resp = await fetch('http://localhost:8947/api/v1/projects', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: projectName, description: 'Governance E2E test project', owner: 'test-user' })
    })
    if (resp.ok) {
      const project = await safeJson(resp)
      console.log(`[setup-daemon] Created governance test project: ${project?.id || '?'}`)
    } else if (resp.status === 409) {
      // Race condition: project was created between our list and create calls — look it up
      const projects2 = await listProjects()
      const existing = (projects2 || []).find(p => p.name === projectName)
      if (existing) {
        console.log(`[setup-daemon] Governance test project found after 409: ${existing.id}`)
      } else {
        console.log(`[setup-daemon] Governance test project conflict (409) but not found in list`)
      }
    } else {
      console.log(`[setup-daemon] Failed to create governance test project: ${resp.status}`)
    }
  } catch (e) {
    console.log(`[setup-daemon] Governance test project creation skipped: ${e.message}`)
  }
}

export { startDaemon, stopDaemon, isDaemonRunning, cleanupTestUsers, cleanupTestProjects, cleanupTestStorage, createGovernanceTestProject, checkZombieInstances }