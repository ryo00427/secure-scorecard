// =============================================================================
// Stress Test - ストレステスト (k6)
// =============================================================================
// システムの限界を測定するストレステスト
// 実行方法: k6 run stress-test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// =============================================================================
// カスタムメトリクス
// =============================================================================

const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');

// =============================================================================
// テスト設定
// =============================================================================

export const options = {
  stages: [
    // ランプアップ: 徐々に負荷を増加
    { duration: '2m', target: 100 },   // 2分で100VUまで
    { duration: '5m', target: 100 },   // 5分間100VU維持
    { duration: '2m', target: 200 },   // 2分で200VUまで
    { duration: '5m', target: 200 },   // 5分間200VU維持
    { duration: '2m', target: 300 },   // 2分で300VUまで
    { duration: '5m', target: 300 },   // 5分間300VU維持
    { duration: '2m', target: 400 },   // 2分で400VUまで
    { duration: '5m', target: 400 },   // 5分間400VU維持
    { duration: '10m', target: 0 },    // 10分でクールダウン
  ],
  thresholds: {
    'http_req_duration': ['p(95)<2000'], // ストレス下でも2秒以内
    'http_req_failed': ['rate<0.1'],     // エラーレート10%未満
    'errors': ['rate<0.1'],
  },
};

// =============================================================================
// 設定
// =============================================================================

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// =============================================================================
// セットアップ
// =============================================================================

export function setup() {
  // 複数のテストユーザーを作成
  const users = [];

  for (let i = 0; i < 10; i++) {
    const email = `stress-test-${Date.now()}-${i}@example.com`;
    const res = http.post(
      `${BASE_URL}/api/v1/auth/register`,
      JSON.stringify({
        email: email,
        password: 'StressTest123!',
        display_name: `Stress Test User ${i}`,
      }),
      { headers: { 'Content-Type': 'application/json' } }
    );

    if (res.status === 201) {
      const data = JSON.parse(res.body);
      users.push({
        email: email,
        token: data.token,
        userId: data.user.id,
      });
    }
  }

  console.log(`${users.length}ユーザーを作成しました`);
  return { users };
}

// =============================================================================
// メインテストシナリオ
// =============================================================================

export default function(data) {
  if (!data || !data.users || data.users.length === 0) {
    console.error('セットアップデータがありません');
    return;
  }

  // ランダムにユーザーを選択
  const user = data.users[Math.floor(Math.random() * data.users.length)];
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${user.token}`,
  };

  // 混合ワークロード
  const operations = [
    { weight: 40, fn: () => getGardens(headers) },
    { weight: 30, fn: () => getCrops(headers) },
    { weight: 20, fn: () => getTasks(headers) },
    { weight: 10, fn: () => getAnalytics(headers) },
  ];

  // 重み付きランダム選択
  const random = Math.random() * 100;
  let cumulative = 0;

  for (const op of operations) {
    cumulative += op.weight;
    if (random < cumulative) {
      op.fn();
      break;
    }
  }

  sleep(Math.random() * 2); // 0-2秒のランダムスリープ
}

// =============================================================================
// 操作関数
// =============================================================================

function getGardens(headers) {
  const res = http.get(`${BASE_URL}/api/v1/gardens`, { headers });
  responseTime.add(res.timings.duration);

  const success = check(res, {
    'gardens status 200': (r) => r.status === 200,
    'gardens response time < 1s': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
  return res;
}

function getCrops(headers) {
  const res = http.get(`${BASE_URL}/api/v1/crops`, { headers });
  responseTime.add(res.timings.duration);

  const success = check(res, {
    'crops status 200': (r) => r.status === 200,
    'crops response time < 1s': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
  return res;
}

function getTasks(headers) {
  const res = http.get(`${BASE_URL}/api/v1/tasks`, { headers });
  responseTime.add(res.timings.duration);

  const success = check(res, {
    'tasks status 200': (r) => r.status === 200,
    'tasks response time < 1s': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
  return res;
}

function getAnalytics(headers) {
  const res = http.get(`${BASE_URL}/api/v1/analytics/harvest-summary`, { headers });
  responseTime.add(res.timings.duration);

  const success = check(res, {
    'analytics status 200': (r) => r.status === 200,
    'analytics response time < 2s': (r) => r.timings.duration < 2000,
  });

  errorRate.add(!success);
  return res;
}

// =============================================================================
// サマリー
// =============================================================================

export function handleSummary(data) {
  const report = {
    test_type: 'stress_test',
    timestamp: new Date().toISOString(),
    metrics: {
      total_requests: data.metrics.http_reqs?.values?.count || 0,
      failed_requests: data.metrics.http_req_failed?.values?.rate || 0,
      avg_duration_ms: data.metrics.http_req_duration?.values?.avg || 0,
      p95_duration_ms: data.metrics.http_req_duration?.values?.['p(95)'] || 0,
      p99_duration_ms: data.metrics.http_req_duration?.values?.['p(99)'] || 0,
      max_duration_ms: data.metrics.http_req_duration?.values?.max || 0,
      max_vus: data.metrics.vus?.values?.max || 0,
    },
    thresholds_passed: !Object.values(data.root_group?.checks || {}).some(
      (c) => c.fails > 0
    ),
  };

  console.log('=== ストレステストレポート ===');
  console.log(`最大VU: ${report.metrics.max_vus}`);
  console.log(`総リクエスト: ${report.metrics.total_requests}`);
  console.log(`エラー率: ${(report.metrics.failed_requests * 100).toFixed(2)}%`);
  console.log(`平均レスポンス時間: ${report.metrics.avg_duration_ms.toFixed(2)}ms`);
  console.log(`P95レスポンス時間: ${report.metrics.p95_duration_ms.toFixed(2)}ms`);
  console.log(`P99レスポンス時間: ${report.metrics.p99_duration_ms.toFixed(2)}ms`);
  console.log(`最大レスポンス時間: ${report.metrics.max_duration_ms.toFixed(2)}ms`);

  return {
    'stdout': JSON.stringify(report, null, 2),
    'stress-test-report.json': JSON.stringify(data, null, 2),
  };
}
