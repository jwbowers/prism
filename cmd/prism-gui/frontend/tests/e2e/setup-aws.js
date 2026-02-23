// Global setup for real-AWS E2E tests
// Launches a real AWS instance before tests and writes its name to a temp file.
import { startDaemon } from './setup-daemon.js'
import fs from 'fs'

const DAEMON_URL = 'http://localhost:8947'
const INSTANCE_FILE = '/tmp/prism-e2e-instance.txt'
const WAIT_TIMEOUT_MS = 5 * 60 * 1000 // 5 minutes

async function waitForInstanceRunning(instanceName, timeoutMs) {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    try {
      const res = await fetch(`${DAEMON_URL}/api/v1/instances`)
      if (res.ok) {
        const data = await res.json()
        const instances = data.instances || data
        const inst = Array.isArray(instances) ? instances.find(i => i.name === instanceName) : null
        if (inst && inst.state === 'running') {
          console.log(`[setup-aws] Instance ${instanceName} is running`)
          return
        }
        if (inst) {
          console.log(`[setup-aws] Instance ${instanceName} state: ${inst.state}, waiting...`)
        }
      }
    } catch (e) {
      // Daemon not ready yet
    }
    await new Promise(r => setTimeout(r, 10000))
  }
  throw new Error(`Instance ${instanceName} did not become running within ${timeoutMs / 1000}s`)
}

async function globalSetup() {
  // Start daemon (builds if needed)
  await startDaemon()

  const instanceName = `prism-e2e-aws-${Date.now()}`
  console.log(`[setup-aws] Launching test instance: ${instanceName}`)

  const res = await fetch(`${DAEMON_URL}/api/v1/instances`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: instanceName,
      template: 'basic-ubuntu',
      instance_type: 't3.small',
    }),
  })

  if (!res.ok) {
    const body = await res.text()
    throw new Error(`Failed to launch test instance: ${res.status} ${body}`)
  }

  console.log(`[setup-aws] Waiting for ${instanceName} to be running...`)
  await waitForInstanceRunning(instanceName, WAIT_TIMEOUT_MS)

  fs.writeFileSync(INSTANCE_FILE, instanceName)
  console.log(`[setup-aws] Instance ${instanceName} ready, name saved to ${INSTANCE_FILE}`)
}

export default globalSetup
