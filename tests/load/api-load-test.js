import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// エラーレート
const errorRate = new Rate('errors');

// レイテンシトレンド
const loginLatency = new Trend('login_latency');
const getGardensLatency = new Trend('get_gardens_latency');
const getCropsLatency = new Trend('get_crops_latency');
const getTasksLatency = new Trend('get_tasks_latency');

// リクエストカウンター
const requestsTotal = new Counter('requests_total');

export const options = {
  // シナリオ定義
  scenarios: {
    // ウォームアップ
    warmup: {
      executor: 'constant-vus',
      vus: 5,
      duration: '30s',
      startTime: '0s',
      tags: { phase: 'warmup' },
    },
    // 通常負荷
    normal_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 50 },   // 1分で50VUまで増加
        { duration: '3m', target: 50 },   // 3分間50VU維持
        { duration: '1m', target: 0 },    // 1分でクールダウン
      ],
      startTime: '30s',
      tags: { phase: 'normal' },
    },
    // スパイク負荷
    spike_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 100 },  // 急激に100VUまで増加
        { duration: '30s', target: 100 },  // 30秒維持
        { duration: '10s', target: 0 },    // クールダウン
      ],
      startTime: '6m',
      tags: { phase: 'spike' },
    },
  },
  // しきい値
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],  // 95%のリクエストが500ms未満
    'http_req_failed': ['rate<0.01'],                   // エラーレート1%未満
    'errors': ['rate<0.05'],                            // カスタムエラーレート5%未満
    'login_latency': ['p(95)<300'],                     // ログインは300ms未満
    'get_gardens_latency': ['p(95)<200'],               // 菜園取得は200ms未満
    'get_crops_latency': ['p(95)<200'],                 // 作物取得は200ms未満
    'get_tasks_latency': ['p(95)<200'],                 // タスク取得は200ms未満
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TEST_USER_EMAIL = `loadtest-${Date.now()}@example.com`;
const TEST_USER_PASSWORD = 'LoadTest123!';

// レスポンスチェック
function checkResponse(res, name, expectedStatus = 200) {
  const success = check(res, {
    [`${name} status is ${expectedStatus}`]: (r) => r.status === expectedStatus,
    [`${name} response time < 1s`]: (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
  requestsTotal.add(1);

  return success;
}

// JSONパース
function parseJSON(res) {
  try {
    return JSON.parse(res.body);
  } catch (e) {
    return null;
  }
}

export function setup() {
  // テストユーザーを登録
  const registerRes = http.post(
    `${BASE_URL}/api/v1/auth/register`,
    JSON.stringify({
      email: TEST_USER_EMAIL,
      password: TEST_USER_PASSWORD,
      display_name: 'Load Test User',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  if (registerRes.status !== 201) {
    console.error('ユーザー登録失敗:', registerRes.body);
    return null;
  }

  const data = parseJSON(registerRes);

  // テスト用の菜園を作成
  const gardenRes = http.post(
    `${BASE_URL}/api/v1/gardens`,
    JSON.stringify({
      name: 'Load Test Garden',
      location: 'Tokyo',
    }),
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${data.token}`,
      },
    }
  );

  const gardenData = parseJSON(gardenRes);

  return {
    token: data.token,
    userId: data.user.id,
    gardenId: gardenData?.garden?.id,
  };
}

export default function(data) {
  if (!data || !data.token) {
    console.error('セットアップデータがありません');
    return;
  }

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
  };

  group('Authentication', () => {
    const loginRes = http.post(
      `${BASE_URL}/api/v1/auth/login`,
      JSON.stringify({
        email: TEST_USER_EMAIL,
        password: TEST_USER_PASSWORD,
      }),
      { headers: { 'Content-Type': 'application/json' } }
    );

    loginLatency.add(loginRes.timings.duration);
    checkResponse(loginRes, 'Login', 200);
  });

  sleep(0.5);

  group('Gardens', () => {
    // 一覧取得
    const listRes = http.get(`${BASE_URL}/api/v1/gardens`, { headers });
    getGardensLatency.add(listRes.timings.duration);
    checkResponse(listRes, 'Get Gardens', 200);

    // 詳細取得
    if (data.gardenId) {
      const detailRes = http.get(
        `${BASE_URL}/api/v1/gardens/${data.gardenId}`,
        { headers }
      );
      checkResponse(detailRes, 'Get Garden Detail', 200);
    }
  });

  sleep(0.5);

  group('Crops', () => {
    // 一覧取得
    const listRes = http.get(`${BASE_URL}/api/v1/crops`, { headers });
    getCropsLatency.add(listRes.timings.duration);
    checkResponse(listRes, 'Get Crops', 200);

    // 作物作成（低頻度）
    if (Math.random() < 0.1) {
      const createRes = http.post(
        `${BASE_URL}/api/v1/crops`,
        JSON.stringify({
          garden_id: data.gardenId,
          name: `Load Test Crop ${Date.now()}`,
          variety: 'Test',
          planted_date: new Date().toISOString(),
        }),
        { headers }
      );
      checkResponse(createRes, 'Create Crop', 201);
    }
  });

  sleep(0.5);

  group('Tasks', () => {
    // 一覧取得
    const listRes = http.get(`${BASE_URL}/api/v1/tasks`, { headers });
    getTasksLatency.add(listRes.timings.duration);
    checkResponse(listRes, 'Get Tasks', 200);

    // 今日のタスク取得
    const todayRes = http.get(`${BASE_URL}/api/v1/tasks/today`, { headers });
    checkResponse(todayRes, 'Get Today Tasks', 200);
  });

  sleep(0.5);

  group('Analytics', () => {
    const summaryRes = http.get(
      `${BASE_URL}/api/v1/analytics/harvest-summary`,
      { headers }
    );
    checkResponse(summaryRes, 'Get Harvest Summary', 200);
  });

  sleep(1);
}

export function teardown(data) {
  if (!data) return;

  console.log('=== 負荷テスト完了 ===');
  console.log(`テストユーザー: ${TEST_USER_EMAIL}`);
}

export function handleSummary(data) {
  console.log('=== テストサマリー ===');

  // メトリクスを出力
  const metrics = {
    total_requests: data.metrics.http_reqs?.values?.count || 0,
    failed_requests: data.metrics.http_req_failed?.values?.rate || 0,
    avg_duration: data.metrics.http_req_duration?.values?.avg || 0,
    p95_duration: data.metrics.http_req_duration?.values?.['p(95)'] || 0,
    p99_duration: data.metrics.http_req_duration?.values?.['p(99)'] || 0,
  };

  console.log(`総リクエスト数: ${metrics.total_requests}`);
  console.log(`失敗率: ${(metrics.failed_requests * 100).toFixed(2)}%`);
  console.log(`平均レイテンシ: ${metrics.avg_duration.toFixed(2)}ms`);
  console.log(`P95レイテンシ: ${metrics.p95_duration.toFixed(2)}ms`);
  console.log(`P99レイテンシ: ${metrics.p99_duration.toFixed(2)}ms`);

  return {
    'stdout': JSON.stringify(metrics, null, 2),
    'summary.json': JSON.stringify(data, null, 2),
  };
}
