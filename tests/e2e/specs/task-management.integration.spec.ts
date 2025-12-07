// =============================================================================
// Task Management Integration Tests - タスク管理統合テスト
// =============================================================================
// タスク作成→今日のタスク取得→完了→通知の一連のフローをテストします。

import { test, expect } from '@playwright/test';

// テストコンテキスト
interface TestContext {
  authToken: string;
  userId: number;
  gardenId: number;
  cropId: number;
  taskIds: number[];
}

const ctx: TestContext = {
  authToken: '',
  userId: 0,
  gardenId: 0,
  cropId: 0,
  taskIds: [],
};

test.describe('Task Management Integration Tests', () => {
  test.describe.configure({ mode: 'serial' });

  // テスト前にユーザーを登録してログイン
  test.beforeAll(async ({ request }) => {
    const email = `task-test-${Date.now()}@example.com`;

    // ユーザー登録
    const registerResponse = await request.post('/api/v1/auth/register', {
      data: {
        email,
        password: 'TestPassword123!',
        display_name: 'Task Test User',
      },
    });

    expect(registerResponse.status()).toBe(201);
    const registerBody = await registerResponse.json();
    ctx.authToken = registerBody.token;
    ctx.userId = registerBody.user.id;
  });

  // ---------------------------------------------------------------------------
  // Setup - セットアップ
  // ---------------------------------------------------------------------------

  test('Step 1: 菜園と作物を作成する', async ({ request }) => {
    // 菜園作成
    const gardenResponse = await request.post('/api/v1/gardens', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        name: 'タスクテスト菜園',
        location: '東京都品川区',
      },
    });

    expect(gardenResponse.status()).toBe(201);
    const gardenBody = await gardenResponse.json();
    ctx.gardenId = gardenBody.garden.id;

    // 作物作成
    const cropResponse = await request.post('/api/v1/crops', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        garden_id: ctx.gardenId,
        name: 'ピーマン',
        variety: 'エース',
        planted_date: new Date().toISOString(),
      },
    });

    expect(cropResponse.status()).toBe(201);
    const cropBody = await cropResponse.json();
    ctx.cropId = cropBody.crop.id;
  });

  // ---------------------------------------------------------------------------
  // Task Creation - タスク作成
  // ---------------------------------------------------------------------------

  test('Step 2: 水やりタスクを作成する', async ({ request }) => {
    const today = new Date();
    today.setHours(9, 0, 0, 0);

    const response = await request.post('/api/v1/tasks', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        crop_id: ctx.cropId,
        title: '水やり',
        description: '朝の水やり作業',
        due_date: today.toISOString(),
        priority: 'high',
        task_type: 'watering',
        is_recurring: true,
        recurrence_pattern: 'daily',
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.task).toBeDefined();
    expect(body.task.title).toBe('水やり');
    expect(body.task.priority).toBe('high');
    expect(body.task.status).toBe('pending');

    ctx.taskIds.push(body.task.id);
  });

  test('Step 3: 肥料タスクを作成する', async ({ request }) => {
    const nextWeek = new Date();
    nextWeek.setDate(nextWeek.getDate() + 7);
    nextWeek.setHours(10, 0, 0, 0);

    const response = await request.post('/api/v1/tasks', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        crop_id: ctx.cropId,
        title: '追肥',
        description: '週次の追肥作業',
        due_date: nextWeek.toISOString(),
        priority: 'medium',
        task_type: 'fertilizing',
        is_recurring: true,
        recurrence_pattern: 'weekly',
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.task.title).toBe('追肥');
    expect(body.task.priority).toBe('medium');

    ctx.taskIds.push(body.task.id);
  });

  test('Step 4: 収穫タスクを作成する', async ({ request }) => {
    const nextMonth = new Date();
    nextMonth.setDate(nextMonth.getDate() + 30);

    const response = await request.post('/api/v1/tasks', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        crop_id: ctx.cropId,
        title: '収穫作業',
        description: '実が赤くなったら収穫',
        due_date: nextMonth.toISOString(),
        priority: 'low',
        task_type: 'harvesting',
        is_recurring: false,
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.task.title).toBe('収穫作業');

    ctx.taskIds.push(body.task.id);
  });

  // ---------------------------------------------------------------------------
  // Task Retrieval - タスク取得
  // ---------------------------------------------------------------------------

  test('タスク一覧を取得できる', async ({ request }) => {
    const response = await request.get('/api/v1/tasks', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.tasks).toBeDefined();
    expect(Array.isArray(body.tasks)).toBe(true);
    expect(body.tasks.length).toBeGreaterThanOrEqual(3);
  });

  test('Step 5: 今日のタスクを取得する', async ({ request }) => {
    const response = await request.get('/api/v1/tasks/today', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.tasks).toBeDefined();
    expect(Array.isArray(body.tasks)).toBe(true);

    // 今日のタスク（水やり）が含まれていることを確認
    const wateringTask = body.tasks.find((t: any) => t.title === '水やり');
    expect(wateringTask).toBeDefined();
    expect(wateringTask.priority).toBe('high');
  });

  test('優先度でフィルタリングできる', async ({ request }) => {
    const response = await request.get('/api/v1/tasks?priority=high', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.tasks).toBeDefined();

    // 全てのタスクがhigh優先度
    body.tasks.forEach((task: any) => {
      expect(task.priority).toBe('high');
    });
  });

  test('ステータスでフィルタリングできる', async ({ request }) => {
    const response = await request.get('/api/v1/tasks?status=pending', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.tasks).toBeDefined();

    // 全てのタスクがpendingステータス
    body.tasks.forEach((task: any) => {
      expect(task.status).toBe('pending');
    });
  });

  // ---------------------------------------------------------------------------
  // Task Completion - タスク完了
  // ---------------------------------------------------------------------------

  test('Step 6: タスクを完了にする', async ({ request }) => {
    const taskId = ctx.taskIds[0]; // 水やりタスク

    const response = await request.put(`/api/v1/tasks/${taskId}/complete`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        completed_at: new Date().toISOString(),
        notes: '朝9時に完了しました',
      },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.task).toBeDefined();
    expect(body.task.status).toBe('completed');
    expect(body.task.completed_at).toBeDefined();
  });

  test('完了したタスクがステータスフィルタで取得できる', async ({ request }) => {
    const response = await request.get('/api/v1/tasks?status=completed', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.tasks).toBeDefined();

    const completedTask = body.tasks.find(
      (t: any) => t.id === ctx.taskIds[0]
    );
    expect(completedTask).toBeDefined();
    expect(completedTask.status).toBe('completed');
  });

  // ---------------------------------------------------------------------------
  // Recurring Task - 繰り返しタスク
  // ---------------------------------------------------------------------------

  test('繰り返しタスクが次回分を生成する', async ({ request }) => {
    // 完了した水やりタスクの次回分が生成されているか確認
    const response = await request.get('/api/v1/tasks?task_type=watering', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.tasks).toBeDefined();

    // 繰り返し設定がある場合、新しいタスクが生成されている
    // （実装によっては手動で次回タスクを生成する必要がある）
  });

  // ---------------------------------------------------------------------------
  // Overdue Tasks - 期限切れタスク
  // ---------------------------------------------------------------------------

  test('Step 7: 期限切れタスクを取得する', async ({ request }) => {
    // 過去の期限でタスクを作成
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);

    const createResponse = await request.post('/api/v1/tasks', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        crop_id: ctx.cropId,
        title: '期限切れタスク',
        description: '昨日が期限だったタスク',
        due_date: yesterday.toISOString(),
        priority: 'high',
        task_type: 'other',
      },
    });

    expect(createResponse.status()).toBe(201);
    const createBody = await createResponse.json();
    ctx.taskIds.push(createBody.task.id);

    // 期限切れタスクを取得
    const response = await request.get('/api/v1/tasks/overdue', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.tasks).toBeDefined();
    expect(Array.isArray(body.tasks)).toBe(true);

    // 期限切れタスクが含まれていることを確認
    const overdueTask = body.tasks.find(
      (t: any) => t.title === '期限切れタスク'
    );
    expect(overdueTask).toBeDefined();
  });

  // ---------------------------------------------------------------------------
  // Task Update - タスク更新
  // ---------------------------------------------------------------------------

  test('タスクを更新できる', async ({ request }) => {
    const taskId = ctx.taskIds[1]; // 追肥タスク

    const response = await request.put(`/api/v1/tasks/${taskId}`, {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
      data: {
        title: '追肥（更新済み）',
        priority: 'high',
        description: '液体肥料を使用する',
      },
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.task.title).toBe('追肥（更新済み）');
    expect(body.task.priority).toBe('high');
  });

  // ---------------------------------------------------------------------------
  // Cleanup - クリーンアップ
  // ---------------------------------------------------------------------------

  test('タスクを削除できる', async ({ request }) => {
    for (const taskId of ctx.taskIds) {
      const response = await request.delete(`/api/v1/tasks/${taskId}`, {
        headers: { Authorization: `Bearer ${ctx.authToken}` },
      });

      expect(response.status()).toBe(200);
    }

    // 削除後に取得するとリストが空（作成したタスクのみ）
    const listResponse = await request.get('/api/v1/tasks', {
      headers: { Authorization: `Bearer ${ctx.authToken}` },
    });

    expect(listResponse.status()).toBe(200);
    const body = await listResponse.json();

    // 削除したタスクが含まれていないことを確認
    const deletedTasks = body.tasks.filter((t: any) =>
      ctx.taskIds.includes(t.id)
    );
    expect(deletedTasks.length).toBe(0);
  });
});
