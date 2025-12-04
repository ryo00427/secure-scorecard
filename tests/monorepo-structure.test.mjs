/**
 * Monorepo Structure Test
 * モノレポ構成が正しく設定されているかを検証するテスト
 *
 * ESM 形式（root package.json で "type": "module" 設定済み）
 */

import { existsSync, readFileSync } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';
import assert from 'assert';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const ROOT_DIR = join(__dirname, '..');

// テストユーティリティ
function test(name, fn) {
  try {
    fn();
    console.log(`✓ ${name}`);
    return true;
  } catch (error) {
    console.error(`✗ ${name}`);
    console.error(`  ${error.message}`);
    return false;
  }
}

function readJSON(path) {
  return JSON.parse(readFileSync(path, 'utf-8'));
}

console.log('=== Monorepo Structure Tests ===\n');

let passed = 0;
let failed = 0;

// 1. ルートディレクトリ構成テスト
if (
  test('ルートに必要な設定ファイルが存在する', () => {
    const requiredFiles = [
      'package.json',
      'pnpm-workspace.yaml',
      'turbo.json',
      'tsconfig.json',
      '.npmrc',
    ];
    for (const file of requiredFiles) {
      assert.ok(
        existsSync(join(ROOT_DIR, file)),
        `${file} が存在しません`
      );
    }
  })
) {
  passed++;
} else {
  failed++;
}

// 2. apps/ ディレクトリ構成テスト
if (
  test('apps/ ディレクトリに backend と mobile が存在する', () => {
    const apps = ['backend', 'mobile'];
    for (const app of apps) {
      assert.ok(
        existsSync(join(ROOT_DIR, 'apps', app)),
        `apps/${app} が存在しません`
      );
    }
  })
) {
  passed++;
} else {
  failed++;
}

// 3. packages/ ディレクトリ構成テスト
if (
  test('packages/ ディレクトリに shared と typescript-config が存在する', () => {
    const packages = ['shared', 'typescript-config'];
    for (const pkg of packages) {
      assert.ok(
        existsSync(join(ROOT_DIR, 'packages', pkg)),
        `packages/${pkg} が存在しません`
      );
    }
  })
) {
  passed++;
} else {
  failed++;
}

// 4. pnpm-workspace.yaml の内容テスト
if (
  test('pnpm-workspace.yaml が正しく設定されている', () => {
    const content = readFileSync(
      join(ROOT_DIR, 'pnpm-workspace.yaml'),
      'utf-8'
    );
    assert.ok(
      content.includes('apps/*'),
      'apps/* が pnpm-workspace.yaml に含まれていません'
    );
    assert.ok(
      content.includes('packages/*'),
      'packages/* が pnpm-workspace.yaml に含まれていません'
    );
  })
) {
  passed++;
} else {
  failed++;
}

// 5. turbo.json の形式テスト（v2: tasks 形式）
if (
  test('turbo.json が v2 形式（tasks）で設定されている', () => {
    const turboConfig = readJSON(join(ROOT_DIR, 'turbo.json'));
    assert.ok(
      turboConfig.tasks !== undefined,
      'turbo.json に tasks フィールドがありません（v1 の pipeline 形式は非推奨）'
    );
    assert.ok(
      turboConfig.tasks.build !== undefined,
      'tasks.build が定義されていません'
    );
    assert.ok(
      turboConfig.tasks.lint !== undefined,
      'tasks.lint が定義されていません'
    );
    assert.ok(
      turboConfig.tasks.test !== undefined,
      'tasks.test が定義されていません'
    );
  })
) {
  passed++;
} else {
  failed++;
}

// 6. ルート package.json の workspaces テスト
if (
  test('ルート package.json に workspaces が設定されている', () => {
    const pkg = readJSON(join(ROOT_DIR, 'package.json'));
    assert.ok(
      Array.isArray(pkg.workspaces),
      'workspaces が配列ではありません'
    );
    assert.ok(
      pkg.workspaces.includes('apps/*'),
      'workspaces に apps/* が含まれていません'
    );
    assert.ok(
      pkg.workspaces.includes('packages/*'),
      'workspaces に packages/* が含まれていません'
    );
  })
) {
  passed++;
} else {
  failed++;
}

// 7. backend の構成テスト
if (
  test('apps/backend に Go プロジェクトが設定されている', () => {
    assert.ok(
      existsSync(join(ROOT_DIR, 'apps', 'backend', 'go.mod')),
      'go.mod が存在しません'
    );
    assert.ok(
      existsSync(join(ROOT_DIR, 'apps', 'backend', 'cmd', 'server', 'main.go')),
      'cmd/server/main.go が存在しません'
    );
  })
) {
  passed++;
} else {
  failed++;
}

// 8. mobile の構成テスト
if (
  test('apps/mobile に Expo プロジェクトが設定されている', () => {
    assert.ok(
      existsSync(join(ROOT_DIR, 'apps', 'mobile', 'package.json')),
      'package.json が存在しません'
    );
    assert.ok(
      existsSync(join(ROOT_DIR, 'apps', 'mobile', 'app.json')),
      'app.json が存在しません'
    );
    const pkg = readJSON(join(ROOT_DIR, 'apps', 'mobile', 'package.json'));
    assert.ok(
      pkg.dependencies?.expo,
      'expo が dependencies に含まれていません'
    );
  })
) {
  passed++;
} else {
  failed++;
}

// 9. shared パッケージの構成テスト
if (
  test('packages/shared が正しく設定されている', () => {
    assert.ok(
      existsSync(join(ROOT_DIR, 'packages', 'shared', 'package.json')),
      'package.json が存在しません'
    );
    assert.ok(
      existsSync(join(ROOT_DIR, 'packages', 'shared', 'tsconfig.json')),
      'tsconfig.json が存在しません'
    );
    assert.ok(
      existsSync(join(ROOT_DIR, 'packages', 'shared', 'src')),
      'src/ ディレクトリが存在しません'
    );
  })
) {
  passed++;
} else {
  failed++;
}

// 10. typescript-config パッケージの構成テスト
if (
  test('packages/typescript-config が正しく設定されている', () => {
    const configDir = join(ROOT_DIR, 'packages', 'typescript-config');
    assert.ok(
      existsSync(join(configDir, 'base.json')),
      'base.json が存在しません'
    );
    assert.ok(
      existsSync(join(configDir, 'node.json')),
      'node.json が存在しません'
    );
    assert.ok(
      existsSync(join(configDir, 'react-native.json')),
      'react-native.json が存在しません'
    );
  })
) {
  passed++;
} else {
  failed++;
}

// 結果サマリー
console.log(`\n=== Results: ${passed} passed, ${failed} failed ===`);
process.exit(failed > 0 ? 1 : 0);
