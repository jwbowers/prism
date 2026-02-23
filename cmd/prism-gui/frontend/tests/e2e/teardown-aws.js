// Global teardown for real-AWS E2E tests
// Terminates the test instance launched by setup-aws.js.
import fs from 'fs'

const DAEMON_URL = 'http://localhost:8947'
const INSTANCE_FILE = '/tmp/prism-e2e-instance.txt'

async function globalTeardown() {
  if (!fs.existsSync(INSTANCE_FILE)) {
    console.log('[teardown-aws] No instance file found, nothing to terminate')
    return
  }

  const instanceName = fs.readFileSync(INSTANCE_FILE, 'utf8').trim()
  if (!instanceName) {
    console.log('[teardown-aws] Instance file is empty, nothing to terminate')
    return
  }

  console.log(`[teardown-aws] Terminating test instance: ${instanceName}`)
  try {
    const res = await fetch(`${DAEMON_URL}/api/v1/instances/${instanceName}/terminate`, {
      method: 'POST',
    })
    if (res.ok || res.status === 404) {
      fs.unlinkSync(INSTANCE_FILE)
      console.log(`[teardown-aws] Instance ${instanceName} terminated`)
    } else {
      const body = await res.text()
      console.error(`[teardown-aws] Failed to terminate: ${res.status} ${body}`)
      console.error(`[teardown-aws] Manually terminate: prism terminate ${instanceName}`)
    }
  } catch (e) {
    console.error(`[teardown-aws] Failed to terminate instance: ${e.message}`)
    console.error(`[teardown-aws] Manually terminate: prism terminate ${instanceName}`)
  }
}

export default globalTeardown
