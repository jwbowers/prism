#!/usr/bin/env node
/**
 * Test Execution Isolation Wrapper
 *
 * Ensures only ONE Playwright test suite runs at a time using file-based locking.
 * Prevents daemon startup conflicts and resource contention.
 */

import { execSync } from 'child_process';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const LOCK_FILE = path.join(__dirname, '../../.playwright-lock');
const LOCK_TIMEOUT = 30000; // 30 seconds

function acquireLock() {
  const startTime = Date.now();

  while (fs.existsSync(LOCK_FILE)) {
    if (Date.now() - startTime > LOCK_TIMEOUT) {
      console.error('❌ Another test suite is already running. Timeout waiting for lock.');
      console.error('   If no tests are running, remove: ' + LOCK_FILE);
      process.exit(1);
    }

    // Check if the process holding the lock is still alive
    try {
      const lockPid = fs.readFileSync(LOCK_FILE, 'utf8').trim();
      try {
        process.kill(lockPid, 0); // Check if process exists
      } catch (e) {
        // Process doesn't exist, remove stale lock
        console.log('⚠️  Removing stale lock file from dead process');
        fs.unlinkSync(LOCK_FILE);
        break;
      }
    } catch (e) {
      // Can't read lock file, try to remove it
      try {
        fs.unlinkSync(LOCK_FILE);
        break;
      } catch (e2) {}
    }

    execSync('sleep 1', { stdio: 'ignore' });
  }

  // Create lock file with our PID
  fs.writeFileSync(LOCK_FILE, process.pid.toString());
  console.log('✅ Acquired test execution lock (PID: ' + process.pid + ')');

  // Cleanup on exit
  const cleanup = () => {
    try {
      const currentLockPid = fs.readFileSync(LOCK_FILE, 'utf8').trim();
      if (currentLockPid === process.pid.toString()) {
        fs.unlinkSync(LOCK_FILE);
        console.log('✅ Released test execution lock');
      }
    } catch (e) {
      // Lock file already removed
    }
  };

  process.on('exit', cleanup);
  process.on('SIGINT', () => {
    cleanup();
    process.exit(130);
  });
  process.on('SIGTERM', () => {
    cleanup();
    process.exit(143);
  });
}

// Acquire lock before running tests
acquireLock();

// Run Playwright with all passed arguments
const args = process.argv.slice(2);

// Quote args that contain special characters
const quotedArgs = args.map(arg => {
  if (arg.includes(' ') || arg.includes('|') || arg.includes('&') || arg.includes(';')) {
    return `"${arg}"`;
  }
  return arg;
});

const command = 'npx playwright test ' + quotedArgs.join(' ');

console.log('🎭 Running: ' + command);
console.log('');

try {
  execSync(command, { stdio: 'inherit', shell: '/bin/bash' });
} catch (error) {
  process.exit(error.status || 1);
}
