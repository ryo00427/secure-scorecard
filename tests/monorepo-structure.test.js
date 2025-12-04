#!/usr/bin/env node

/**
 * モノレポ構造検証テスト (Task 1.1)
 * TDD RED Phase: このテストは最初失敗するはず
 */

const fs = require('fs');
const path = require('path');

const rootDir = path.join(__dirname, '..');
let testsPassed = 0;
let testsFailed = 0;

function test(description, fn) {
  try {
    fn();
    console.log(`✅ PASS: ${description}`);
    testsPassed++;
  } catch (error) {
    console.error(`❌ FAIL: ${description}`);
    console.error(`   ${error.message}`);
    testsFailed++;
  }
}

function fileExists(filepath) {
  if (!fs.existsSync(filepath)) {
    throw new Error(`File not found: ${filepath}`);
  }
}

function directoryExists(dirpath) {
  if (!fs.existsSync(dirpath) || !fs.statSync(dirpath).isDirectory()) {
    throw new Error(`Directory not found: ${dirpath}`);
  }
}

function jsonContains(filepath, expectedContent) {
  const content = JSON.parse(fs.readFileSync(filepath, 'utf-8'));
  for (const [key, value] of Object.entries(expectedContent)) {
    if (JSON.stringify(content[key]) !== JSON.stringify(value)) {
      throw new Error(`Expected ${key} to be ${JSON.stringify(value)}, but got ${JSON.stringify(content[key])}`);
    }
  }
}

function fileContains(filepath, expectedString) {
  const content = fs.readFileSync(filepath, 'utf-8');
  if (!content.includes(expectedString)) {
    throw new Error(`File ${filepath} does not contain: ${expectedString}`);
  }
}

console.log('🧪 Running Monorepo Structure Tests (Task 1.1)\n');

// Test 1: package.json exists and has workspace configuration
test('package.json exists in root', () => {
  fileExists(path.join(rootDir, 'package.json'));
});

test('package.json has correct name', () => {
  jsonContains(path.join(rootDir, 'package.json'), {
    name: 'secure-scorecard'
  });
});

test('package.json has workspaces field', () => {
  const packageJson = JSON.parse(fs.readFileSync(path.join(rootDir, 'package.json'), 'utf-8'));
  if (!packageJson.workspaces) {
    throw new Error('package.json does not have workspaces field');
  }
});

test('package.json includes required devDependencies', () => {
  const packageJson = JSON.parse(fs.readFileSync(path.join(rootDir, 'package.json'), 'utf-8'));
  const requiredDeps = ['turbo'];
  for (const dep of requiredDeps) {
    if (!packageJson.devDependencies || !packageJson.devDependencies[dep]) {
      throw new Error(`Missing devDependency: ${dep}`);
    }
  }
});

// Test 2: pnpm-workspace.yaml exists
test('pnpm-workspace.yaml exists', () => {
  fileExists(path.join(rootDir, 'pnpm-workspace.yaml'));
});

test('pnpm-workspace.yaml includes apps/*', () => {
  fileContains(path.join(rootDir, 'pnpm-workspace.yaml'), "apps/*");
});

test('pnpm-workspace.yaml includes packages/*', () => {
  fileContains(path.join(rootDir, 'pnpm-workspace.yaml'), "packages/*");
});

// Test 3: turbo.json exists and has tasks configuration (Turbo v2)
test('turbo.json exists', () => {
  fileExists(path.join(rootDir, 'turbo.json'));
});

test('turbo.json has $schema field', () => {
  const turboJson = JSON.parse(fs.readFileSync(path.join(rootDir, 'turbo.json'), 'utf-8'));
  if (!turboJson.$schema) {
    throw new Error('turbo.json does not have $schema field');
  }
});

test('turbo.json has tasks configuration (Turbo v2)', () => {
  const turboJson = JSON.parse(fs.readFileSync(path.join(rootDir, 'turbo.json'), 'utf-8'));
  if (!turboJson.tasks) {
    throw new Error('turbo.json does not have tasks field (Turbo v2 format)');
  }
});

// Test 4: apps directory structure (deployable applications)
test('apps directory exists', () => {
  directoryExists(path.join(rootDir, 'apps'));
});

test('apps/backend directory exists', () => {
  directoryExists(path.join(rootDir, 'apps', 'backend'));
});

test('apps/mobile directory exists', () => {
  directoryExists(path.join(rootDir, 'apps', 'mobile'));
});

// Test 5: packages directory structure (shared libraries)
test('packages directory exists', () => {
  directoryExists(path.join(rootDir, 'packages'));
});

test('packages/shared directory exists', () => {
  directoryExists(path.join(rootDir, 'packages', 'shared'));
});

// Test 6: Each app/package has package.json
test('apps/backend/package.json exists', () => {
  fileExists(path.join(rootDir, 'apps', 'backend', 'package.json'));
});

test('apps/mobile/package.json exists', () => {
  fileExists(path.join(rootDir, 'apps', 'mobile', 'package.json'));
});

test('packages/shared/package.json exists', () => {
  fileExists(path.join(rootDir, 'packages', 'shared', 'package.json'));
});

// Summary
console.log('\n' + '='.repeat(50));
console.log(`✅ Passed: ${testsPassed}`);
console.log(`❌ Failed: ${testsFailed}`);
console.log(`📊 Total: ${testsPassed + testsFailed}`);
console.log('='.repeat(50));

process.exit(testsFailed > 0 ? 1 : 0);
