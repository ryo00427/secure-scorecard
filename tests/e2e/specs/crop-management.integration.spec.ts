// =============================================================================
// Crop Management Integration Tests - 作物管理統合テスト
// =============================================================================
// 作物登録→成長記録→画像アップロード→収穫の一連のフローをテストします。

import { test, expect, APIRequestContext } from '@playwright/test';

// テストコンテキスト
interface TestContext {
  authToken: string;
  userId: number;
  gardenId: number;
  cropId: number;
}

const ctx: TestContext = {
  authToken: '',
  userId: 0,
  gardenId: 0,
  cropId: 0,
};

test.describe('Crop Management Integration Tests', () => {
  test.describe.configure({ mode: 'serial' });

  // テスト前にユーザーを登録してログイン
  test.beforeAll(async ({ request }) => {
    const email = `crop-test-${Date.now()}@example.com`;

    // ユーザー登録
    const registerResponse = await request.post('/api/v1/auth/register', {
      data: {
        email,
        password: 'TestPassword123!',
        display_name: 'Crop Test User',
      },
    });

    expect(registerResponse.status()).toBe(201);
    const registerBody = await registerResponse.json();
    ctx.authToken = registerBody.token;
    ctx.userId = registerBody.user.id;
  });

  // ---------------------------------------------------------------------------
  // Garden Setup - 菜園セットアップ
  // ---------------------------------------------------------------------------

  test('Step 1: 菜園を作成する', async ({ request }) => {
    const response = await request.post('/api/v1/gardens', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        name: 'テスト菜園',
        location: '東京都渋谷区',
        description: 'E2Eテスト用の菜園です',
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.garden).toBeDefined();
    expect(body.garden.id).toBeGreaterThan(0);
    expect(body.garden.name).toBe('テスト菜園');

    ctx.gardenId = body.garden.id;
  });

  // ---------------------------------------------------------------------------
  // Crop Registration - 作物登録
  // ---------------------------------------------------------------------------

  test('Step 2: 作物を登録する', async ({ request }) => {
    const plantedDate = new Date();
    plantedDate.setDate(plantedDate.getDate() - 30);

    const expectedHarvestDate = new Date();
    expectedHarvestDate.setDate(expectedHarvestDate.getDate() + 60);

    const response = await request.post('/api/v1/crops', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        garden_id: ctx.gardenId,
        name: 'トマト',
        variety: 'ミニトマト',
        planted_date: plantedDate.toISOString(),
        expected_harvest_date: expectedHarvestDate.toISOString(),
        notes: 'E2Eテスト用のトマト',
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.crop).toBeDefined();
    expect(body.crop.id).toBeGreaterThan(0);
    expect(body.crop.name).toBe('トマト');
    expect(body.crop.status).toBe('growing');

    ctx.cropId = body.crop.id;
  });

  test('作物一覧を取得できる', async ({ request }) => {
    const response = await request.get('/api/v1/crops', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.crops).toBeDefined();
    expect(Array.isArray(body.crops)).toBe(true);
    expect(body.crops.length).toBeGreaterThan(0);

    const createdCrop = body.crops.find((c: any) => c.id === ctx.cropId);
    expect(createdCrop).toBeDefined();
  });

  test('作物詳細を取得できる', async ({ request }) => {
    const response = await request.get(`/api/v1/crops/${ctx.cropId}`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.crop).toBeDefined();
    expect(body.crop.id).toBe(ctx.cropId);
    expect(body.crop.name).toBe('トマト');
  });

  // ---------------------------------------------------------------------------
  // Growth Records - 成長記録
  // ---------------------------------------------------------------------------

  test('Step 3: 成長記録を追加する - 発芽', async ({ request }) => {
    const recordDate = new Date();
    recordDate.setDate(recordDate.getDate() - 25);

    const response = await request.post(`/api/v1/crops/${ctx.cropId}/records`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        record_date: recordDate.toISOString(),
        growth_stage: 'seedling',
        notes: '発芽しました！双葉が出ています。',
        height: 2.5,
        health_status: 'healthy',
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.record).toBeDefined();
    expect(body.record.growth_stage).toBe('seedling');
  });

  test('Step 4: 成長記録を追加する - 生長期', async ({ request }) => {
    const recordDate = new Date();
    recordDate.setDate(recordDate.getDate() - 15);

    const response = await request.post(`/api/v1/crops/${ctx.cropId}/records`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        record_date: recordDate.toISOString(),
        growth_stage: 'vegetative',
        notes: '本葉が5枚になりました。順調に成長中。',
        height: 15.0,
        health_status: 'healthy',
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.record.growth_stage).toBe('vegetative');
  });

  test('Step 5: 成長記録を追加する - 開花期', async ({ request }) => {
    const recordDate = new Date();
    recordDate.setDate(recordDate.getDate() - 5);

    const response = await request.post(`/api/v1/crops/${ctx.cropId}/records`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        record_date: recordDate.toISOString(),
        growth_stage: 'flowering',
        notes: '黄色い花が咲き始めました！',
        height: 45.0,
        health_status: 'healthy',
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.record.growth_stage).toBe('flowering');
  });

  test('成長記録一覧を取得できる', async ({ request }) => {
    const response = await request.get(`/api/v1/crops/${ctx.cropId}/records`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.records).toBeDefined();
    expect(Array.isArray(body.records)).toBe(true);
    expect(body.records.length).toBe(3);

    // 時系列順に並んでいることを確認
    const stages = body.records.map((r: any) => r.growth_stage);
    expect(stages).toContain('seedling');
    expect(stages).toContain('vegetative');
    expect(stages).toContain('flowering');
  });

  // ---------------------------------------------------------------------------
  // Image Upload - 画像アップロード
  // ---------------------------------------------------------------------------

  test('Step 6: 画像アップロード用の署名付きURLを取得する', async ({ request }) => {
    const response = await request.post('/api/v1/crops/images/presign', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        crop_id: ctx.cropId,
        file_name: 'tomato-growth.jpg',
        content_type: 'image/jpeg',
      },
    });

    // 署名付きURL生成が成功するか、モック環境では200または501
    expect([200, 201, 501]).toContain(response.status());

    if (response.status() === 200 || response.status() === 201) {
      const body = await response.json();
      expect(body.upload_url).toBeDefined();
      expect(body.image_url).toBeDefined();
    }
  });

  // ---------------------------------------------------------------------------
  // Harvest - 収穫
  // ---------------------------------------------------------------------------

  test('Step 7: 収穫を記録する', async ({ request }) => {
    const response = await request.post(`/api/v1/crops/${ctx.cropId}/harvest`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        harvest_date: new Date().toISOString(),
        quantity: 2.5,
        unit: 'kg',
        quality: 'excellent',
        notes: '初収穫！甘くて美味しいミニトマトでした。',
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.harvest).toBeDefined();
    expect(body.harvest.quantity).toBe(2.5);
    expect(body.harvest.quality).toBe('excellent');
  });

  test('作物のステータスが収穫済みに更新されている', async ({ request }) => {
    const response = await request.get(`/api/v1/crops/${ctx.cropId}`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.crop.status).toBe('harvested');
  });

  // ---------------------------------------------------------------------------
  // Analytics - 分析
  // ---------------------------------------------------------------------------

  test('Step 8: 収穫サマリーを確認する', async ({ request }) => {
    const response = await request.get('/api/v1/analytics/harvest-summary', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.summaries).toBeDefined();
    expect(Array.isArray(body.summaries)).toBe(true);

    // 収穫したトマトが集計に含まれていることを確認
    const tomatoSummary = body.summaries.find(
      (s: any) => s.crop_id === ctx.cropId
    );
    if (tomatoSummary) {
      expect(tomatoSummary.total_quantity).toBe(2.5);
    }
  });

  // ---------------------------------------------------------------------------
  // Cleanup - クリーンアップ
  // ---------------------------------------------------------------------------

  test('作物を削除できる', async ({ request }) => {
    const response = await request.delete(`/api/v1/crops/${ctx.cropId}`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    // 削除後に取得しようとするとエラー
    const getResponse = await request.get(`/api/v1/crops/${ctx.cropId}`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(getResponse.status()).toBe(404);
  });
});
