// =============================================================================
// Playwright Configuration - E2Eテスト設定
// =============================================================================
// API統合テストとUI統合テストの設定を定義します。

import { defineConfig, devices } from '@playwright/test';

// 環境変数からベースURLを取得（デフォルトはローカル開発環境）
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:8080';

export default defineConfig({
  // テストディレクトリ
  testDir: './specs',

  // テストファイルのパターン
  testMatch: '**/*.spec.ts',

  // 並列実行の設定
  fullyParallel: true,

  // CI環境ではリトライなし、ローカルでは1回リトライ
  retries: process.env.CI ? 0 : 1,

  // 並列ワーカー数
  workers: process.env.CI ? 1 : undefined,

  // レポーター設定
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['json', { outputFile: 'test-results/results.json' }],
    ['list'],
  ],

  // 共通設定
  use: {
    // ベースURL
    baseURL: API_BASE_URL,

    // トレース設定（失敗時のみ記録）
    trace: 'on-first-retry',

    // スクリーンショット設定（失敗時のみ）
    screenshot: 'only-on-failure',

    // ビデオ設定（失敗時のみ）
    video: 'on-first-retry',

    // タイムアウト設定
    actionTimeout: 10000,
    navigationTimeout: 30000,

    // 追加のHTTPヘッダー
    extraHTTPHeaders: {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    },
  },

  // グローバルタイムアウト
  timeout: 60000,

  // テストの期待値のタイムアウト
  expect: {
    timeout: 5000,
  },

  // プロジェクト設定（APIテストのみ）
  projects: [
    {
      name: 'api-tests',
      testMatch: '**/*.api.spec.ts',
      use: {
        ...devices['Desktop Chrome'],
      },
    },
    {
      name: 'integration-tests',
      testMatch: '**/*.integration.spec.ts',
      use: {
        ...devices['Desktop Chrome'],
      },
    },
  ],

  // ローカル開発サーバーの設定（必要に応じて）
  // webServer: {
  //   command: 'cd ../../apps/backend && go run cmd/server/main.go',
  //   url: 'http://localhost:8080/api/health',
  //   reuseExistingServer: !process.env.CI,
  //   timeout: 120000,
  // },
});
