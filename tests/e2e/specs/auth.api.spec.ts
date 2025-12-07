// =============================================================================
// Auth API E2E Tests - 認証APIテスト
// =============================================================================
// ユーザー登録・ログイン・ログアウトのE2Eテストを実行します。

import { test, expect, APIRequestContext } from '@playwright/test';

// テスト用のユニークなメールアドレスを生成
const generateEmail = () => `e2e-test-${Date.now()}@example.com`;

// テスト用の認証情報を保持
let authToken: string = '';
let testEmail: string = '';

test.describe('Auth API E2E Tests', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeAll(async () => {
    testEmail = generateEmail();
  });

  // ---------------------------------------------------------------------------
  // Registration Tests - ユーザー登録テスト
  // ---------------------------------------------------------------------------

  test('POST /api/v1/auth/register - 新規ユーザー登録', async ({ request }) => {
    const response = await request.post('/api/v1/auth/register', {
      data: {
        email: testEmail,
        password: 'TestPassword123!',
        display_name: 'E2E Test User',
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.token).toBeDefined();
    expect(body.user).toBeDefined();
    expect(body.user.email).toBe(testEmail);
    expect(body.user.display_name).toBe('E2E Test User');

    // トークンを保存
    authToken = body.token;
  });

  test('POST /api/v1/auth/register - 重複メールアドレスでエラー', async ({ request }) => {
    const response = await request.post('/api/v1/auth/register', {
      data: {
        email: testEmail, // 既に登録済み
        password: 'AnotherPassword123!',
        display_name: 'Duplicate User',
      },
    });

    expect(response.status()).toBe(409);

    const body = await response.json();
    expect(body.error).toBeDefined();
  });

  test('POST /api/v1/auth/register - 無効なメールアドレスでエラー', async ({ request }) => {
    const response = await request.post('/api/v1/auth/register', {
      data: {
        email: 'invalid-email',
        password: 'TestPassword123!',
        display_name: 'Invalid Email User',
      },
    });

    expect(response.status()).toBe(400);
  });

  test('POST /api/v1/auth/register - 弱いパスワードでエラー', async ({ request }) => {
    const response = await request.post('/api/v1/auth/register', {
      data: {
        email: generateEmail(),
        password: 'short',
        display_name: 'Weak Password User',
      },
    });

    expect(response.status()).toBe(400);
  });

  // ---------------------------------------------------------------------------
  // Login Tests - ログインテスト
  // ---------------------------------------------------------------------------

  test('POST /api/v1/auth/login - 正常なログイン', async ({ request }) => {
    const response = await request.post('/api/v1/auth/login', {
      data: {
        email: testEmail,
        password: 'TestPassword123!',
      },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.token).toBeDefined();
    expect(body.user).toBeDefined();
    expect(body.user.email).toBe(testEmail);

    // 新しいトークンを保存
    authToken = body.token;
  });

  test('POST /api/v1/auth/login - 間違ったパスワードでエラー', async ({ request }) => {
    const response = await request.post('/api/v1/auth/login', {
      data: {
        email: testEmail,
        password: 'WrongPassword123!',
      },
    });

    expect(response.status()).toBe(401);

    const body = await response.json();
    expect(body.error).toBeDefined();
  });

  test('POST /api/v1/auth/login - 存在しないユーザーでエラー', async ({ request }) => {
    const response = await request.post('/api/v1/auth/login', {
      data: {
        email: 'nonexistent@example.com',
        password: 'TestPassword123!',
      },
    });

    expect(response.status()).toBe(401);
  });

  // ---------------------------------------------------------------------------
  // Protected API Access Tests - 保護APIアクセステスト
  // ---------------------------------------------------------------------------

  test('GET /api/v1/auth/me - 認証済みユーザー情報取得', async ({ request }) => {
    const response = await request.get('/api/v1/auth/me', {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.user).toBeDefined();
    expect(body.user.email).toBe(testEmail);
  });

  test('GET /api/v1/auth/me - 認証なしでエラー', async ({ request }) => {
    const response = await request.get('/api/v1/auth/me');

    expect(response.status()).toBe(401);
  });

  test('GET /api/v1/auth/me - 無効なトークンでエラー', async ({ request }) => {
    const response = await request.get('/api/v1/auth/me', {
      headers: {
        Authorization: 'Bearer invalid-token',
      },
    });

    expect(response.status()).toBe(401);
  });

  // ---------------------------------------------------------------------------
  // Logout Tests - ログアウトテスト
  // ---------------------------------------------------------------------------

  test('POST /api/v1/auth/logout - 正常なログアウト', async ({ request }) => {
    const response = await request.post('/api/v1/auth/logout', {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.message).toBeDefined();
  });

  test('トークンがブラックリストに追加されている', async ({ request }) => {
    // ログアウト後のトークンでアクセス試行
    const response = await request.get('/api/v1/auth/me', {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    // トークンがブラックリストに入っていれば401
    expect(response.status()).toBe(401);
  });
});
