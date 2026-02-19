#!/usr/bin/env node
/**
 * scripts/build-web.mjs
 *
 * Called by Wails as the "frontend:build" command.
 * Installs dependencies and builds the web app submodule.
 * Output lands in web/dist/ — read by Wails via "frontend:dir".
 */

import { execSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { join, dirname } from 'node:path';
import { existsSync } from 'node:fs';

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const webDir = join(root, 'web');

if (!existsSync(webDir)) {
  console.error(
    'ERROR: web/ submodule is missing.\n' +
    'Run: git submodule update --init --recursive'
  );
  process.exit(1);
}

console.log('→ Installing web app dependencies…');
execSync('npm ci --prefer-offline', { cwd: webDir, stdio: 'inherit' });

console.log('→ Building web app…');
execSync('npm run build', { cwd: webDir, stdio: 'inherit' });

console.log('✓ Web app built → web/dist/');
