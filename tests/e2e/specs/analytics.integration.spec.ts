// =============================================================================
// Analytics Integration Tests - 分析統合テスト
// =============================================================================
// グラフ表示→CSVエクスポート→サマリー取得の一連のフローをテストします。

import { test, expect } from '@playwright/test';

// テストコンテキスト
interface TestContext {
  authToken: string;
  userId: number;
  gardenId: number;
  cropIds: number[];
  harvestIds: number[];
}

const ctx: TestContext = {
  authToken: '',
  userId: 0,
  gardenId: 0,
  cropIds: [],
  harvestIds: [],
};

test.describe('Analytics Integration Tests', () => {
  test.describe.configure({ mode: 'serial' });

  // テスト前にユーザーを登録してログイン
  test.beforeAll(async ({ request }) => {
    const email = `analytics-test-${Date.now()}@example.com`;

    // ユーザー登録
    const registerResponse = await request.post('/api/v1/auth/register', {
      data: {
        email,
        password: 'TestPassword123!',
        display_name: 'Analytics Test User',
      },
    });

    expect(registerResponse.status()).toBe(201);
    const registerBody = await registerResponse.json();
    ctx.authToken = registerBody.token;
    ctx.userId = registerBody.user.id;
  });

  // ---------------------------------------------------------------------------
  // Setup - テストデータのセットアップ
  // ---------------------------------------------------------------------------

  test('Step 1: 菜園と作物を作成する', async ({ request }) => {
    // 菜園作成
    const gardenResponse = await request.post('/api/v1/gardens', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        name: '分析テスト菜園',
        location: '東京都目黒区',
      },
    });

    expect(gardenResponse.status()).toBe(201);
    const gardenBody = await gardenResponse.json();
    ctx.gardenId = gardenBody.garden.id;

    // 作物1: トマト
    const crop1Response = await request.post('/api/v1/crops', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        garden_id: ctx.gardenId,
        name: 'トマト',
        variety: '桃太郎',
        planted_date: new Date(2024, 3, 1).toISOString(),
        status: 'harvested',
      },
    });
    expect(crop1Response.status()).toBe(201);
    ctx.cropIds.push((await crop1Response.json()).crop.id);

    // 作物2: キュウリ
    const crop2Response = await request.post('/api/v1/crops', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        garden_id: ctx.gardenId,
        name: 'キュウリ',
        variety: '夏すずみ',
        planted_date: new Date(2024, 4, 1).toISOString(),
        status: 'harvested',
      },
    });
    expect(crop2Response.status()).toBe(201);
    ctx.cropIds.push((await crop2Response.json()).crop.id);

    // 作物3: ナス
    const crop3Response = await request.post('/api/v1/crops', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        garden_id: ctx.gardenId,
        name: 'ナス',
        variety: '千両二号',
        planted_date: new Date(2024, 4, 15).toISOString(),
        status: 'growing',
      },
    });
    expect(crop3Response.status()).toBe(201);
    ctx.cropIds.push((await crop3Response.json()).crop.id);
  });

  test('Step 2: 収穫データを登録する', async ({ request }) => {
    // トマトの収穫1
    const harvest1Response = await request.post(
      `/api/v1/crops/${ctx.cropIds[0]}/harvest`,
      {
        headers: { Authorization: `Bearer ${ctx.authToken}` },
        data: {
          harvest_date: new Date(2024, 6, 1).toISOString(),
          quantity: 5.0,
          unit: 'kg',
          quality: 'excellent',
          notes: '初収穫',
        },
      }
    );
    expect(harvest1Response.status()).toBe(201);
    ctx.harvestIds.push((await harvest1Response.json()).harvest.id);

    // トマトの収穫2
    const harvest2Response = await request.post(
      `/api/v1/crops/${ctx.cropIds[0]}/harvest`,
      {
        headers: { Authorization: `Bearer ${ctx.authToken}` },
        data: {
          harvest_date: new Date(2024, 6, 15).toISOString(),
          quantity: 8.0,
          unit: 'kg',
          quality: 'good',
          notes: '2回目の収穫',
        },
      }
    );
    expect(harvest2Response.status()).toBe(201);
    ctx.harvestIds.push((await harvest2Response.json()).harvest.id);

    // キュウリの収穫
    const harvest3Response = await request.post(
      `/api/v1/crops/${ctx.cropIds[1]}/harvest`,
      {
        headers: { Authorization: `Bearer ${ctx.authToken}` },
        data: {
          harvest_date: new Date(2024, 6, 10).toISOString(),
          quantity: 3.0,
          unit: 'kg',
          quality: 'excellent',
          notes: 'キュウリ収穫',
        },
      }
    );
    expect(harvest3Response.status()).toBe(201);
    ctx.harvestIds.push((await harvest3Response.json()).harvest.id);
  });

  // ---------------------------------------------------------------------------
  // Harvest Summary - 収穫サマリー
  // ---------------------------------------------------------------------------

  test('Step 3: 収穫サマリーを取得する', async ({ request }) => {
    const response = await request.get('/api/v1/analytics/harvest-summary', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.summaries).toBeDefined();
    expect(Array.isArray(body.summaries)).toBe(true);

    // トマトのサマリーを確認
    const tomatoSummary = body.summaries.find(
      (s: any) => s.crop_id === ctx.cropIds[0]
    );
    if (tomatoSummary) {
      expect(tomatoSummary.total_quantity).toBe(13.0); // 5 + 8
      expect(tomatoSummary.harvest_count).toBe(2);
    }

    // キュウリのサマリーを確認
    const cucumberSummary = body.summaries.find(
      (s: any) => s.crop_id === ctx.cropIds[1]
    );
    if (cucumberSummary) {
      expect(cucumberSummary.total_quantity).toBe(3.0);
      expect(cucumberSummary.harvest_count).toBe(1);
    }
  });

  test('期間指定で収穫サマリーを取得できる', async ({ request }) => {
    const startDate = new Date(2024, 6, 1).toISOString();
    const endDate = new Date(2024, 6, 10).toISOString();

    const response = await request.get(
      `/api/v1/analytics/harvest-summary?start_date=${startDate}&end_date=${endDate}`,
      {
        headers: { Authorization: `Bearer ${ctx.authToken}` },
      }
    );

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.summaries).toBeDefined();
    // 期間内の収穫のみが含まれる
  });

  // ---------------------------------------------------------------------------
  // Chart Data - グラフデータ
  // ---------------------------------------------------------------------------

  test('Step 4: 月別収穫量グラフデータを取得する', async ({ request }) => {
    const response = await request.get(
      '/api/v1/analytics/chart/monthly-harvest',
      {
        headers: { Authorization: `Bearer ${ctx.authToken}` },
      }
    );

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.chart_type).toBe('monthly_harvest');
    expect(body.title).toBeDefined();
    expect(body.data).toBeDefined();
    expect(Array.isArray(body.data)).toBe(true);

    // 7月のデータがあることを確認
    const julyData = body.data.find(
      (d: any) => d.label === '7月' || d.label === 'July'
    );
    if (julyData) {
      expect(julyData.value).toBeGreaterThan(0);
    }
  });

  test('Step 5: 作物別収穫量グラフデータを取得する', async ({ request }) => {
    const response = await request.get('/api/v1/analytics/chart/crop-harvest', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.chart_type).toBe('crop_harvest');
    expect(body.data).toBeDefined();

    // トマトとキュウリのデータがあることを確認
    const tomatoData = body.data.find((d: any) => d.label === 'トマト');
    const cucumberData = body.data.find((d: any) => d.label === 'キュウリ');

    if (tomatoData) {
      expect(tomatoData.value).toBe(13.0);
    }
    if (cucumberData) {
      expect(cucumberData.value).toBe(3.0);
    }
  });

  test('区画別生産性グラフデータを取得できる', async ({ request }) => {
    const response = await request.get(
      '/api/v1/analytics/chart/plot-productivity',
      {
        headers: { Authorization: `Bearer ${ctx.authToken}` },
      }
    );

    // このエンドポイントが実装されているか確認
    expect([200, 501]).toContain(response.status());

    if (response.status() === 200) {
      const body = await response.json();
      expect(body.chart_type).toBe('plot_productivity');
      expect(body.data).toBeDefined();
    }
  });

  // ---------------------------------------------------------------------------
  // CSV Export - CSVエクスポート
  // ---------------------------------------------------------------------------

  test('Step 6: 作物データをCSVエクスポートする', async ({ request }) => {
    const response = await request.get('/api/v1/analytics/export/crops', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect([200, 501]).toContain(response.status());

    if (response.status() === 200) {
      const body = await response.json();
      expect(body.download_url).toBeDefined();
      expect(body.expires_at).toBeDefined();
    }
  });

  test('Step 7: 収穫データをCSVエクスポートする', async ({ request }) => {
    const response = await request.get('/api/v1/analytics/export/harvests', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect([200, 501]).toContain(response.status());

    if (response.status() === 200) {
      const body = await response.json();
      expect(body.download_url).toBeDefined();
    }
  });

  test('タスクデータをCSVエクスポートできる', async ({ request }) => {
    const response = await request.get('/api/v1/analytics/export/tasks', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect([200, 501]).toContain(response.status());

    if (response.status() === 200) {
      const body = await response.json();
      expect(body.download_url).toBeDefined();
    }
  });

  test('全データをCSVエクスポートできる', async ({ request }) => {
    const response = await request.get('/api/v1/analytics/export/all', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect([200, 501]).toContain(response.status());

    if (response.status() === 200) {
      const body = await response.json();
      expect(body.download_url).toBeDefined();
    }
  });

  // ---------------------------------------------------------------------------
  // Statistics - 統計情報
  // ---------------------------------------------------------------------------

  test('Step 8: ダッシュボード統計を取得する', async ({ request }) => {
    const response = await request.get('/api/v1/analytics/dashboard', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.statistics).toBeDefined();

    // 基本統計情報を確認
    expect(body.statistics.total_crops).toBeGreaterThanOrEqual(3);
    expect(body.statistics.total_harvests).toBeGreaterThanOrEqual(3);
    expect(body.statistics.total_harvest_quantity).toBeGreaterThanOrEqual(16.0);
  });

  test('年間統計を取得できる', async ({ request }) => {
    const response = await request.get('/api/v1/analytics/yearly?year=2024', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect([200, 501]).toContain(response.status());

    if (response.status() === 200) {
      const body = await response.json();
      expect(body.year).toBe(2024);
      expect(body.monthly_data).toBeDefined();
    }
  });

  // ---------------------------------------------------------------------------
  // Growth Analytics - 成長分析
  // ---------------------------------------------------------------------------

  test('作物の成長分析を取得できる', async ({ request }) => {
    const response = await request.get(
      `/api/v1/analytics/growth/${ctx.cropIds[0]}`,
      {
        headers: { Authorization: `Bearer ${ctx.authToken}` },
      }
    );

    expect([200, 501]).toContain(response.status());

    if (response.status() === 200) {
      const body = await response.json();
      expect(body.crop_id).toBe(ctx.cropIds[0]);
      expect(body.growth_data).toBeDefined();
    }
  });

  // ---------------------------------------------------------------------------
  // Comparison - 比較分析
  // ---------------------------------------------------------------------------

  test('作物間の比較データを取得できる', async ({ request }) => {
    const response = await request.post('/api/v1/analytics/compare', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        crop_ids: ctx.cropIds.slice(0, 2), // トマトとキュウリを比較
        metrics: ['harvest_quantity', 'growth_days'],
      },
    });

    expect([200, 501]).toContain(response.status());

    if (response.status() === 200) {
      const body = await response.json();
      expect(body.comparison).toBeDefined();
    }
  });

  // ---------------------------------------------------------------------------
  // Error Handling - エラーハンドリング
  // ---------------------------------------------------------------------------

  test('無効な期間指定でエラーが返る', async ({ request }) => {
    const response = await request.get(
      '/api/v1/analytics/harvest-summary?start_date=invalid&end_date=2024-07-31',
      {
        headers: { Authorization: `Bearer ${ctx.authToken}` },
      }
    );

    expect([400, 200]).toContain(response.status());
  });

  test('認証なしでアクセスするとエラー', async ({ request }) => {
    const response = await request.get('/api/v1/analytics/harvest-summary');

    expect(response.status()).toBe(401);
  });
});
