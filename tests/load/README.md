# 負荷テスト / Load Tests

このディレクトリには、k6を使用した負荷テストスクリプトが含まれています。

## 前提条件

- [k6](https://k6.io/docs/getting-started/installation/) がインストールされていること
- バックエンドサーバーが起動していること

## テストスクリプト

### api-load-test.js

通常の負荷テストスクリプトです。以下のシナリオを実行します：

1. **ウォームアップ**: 5VUで30秒
2. **通常負荷**: 50VUで3分間
3. **スパイク負荷**: 100VUで30秒

```bash
# 実行方法
k6 run api-load-test.js

# カスタムベースURL
k6 run -e BASE_URL=http://localhost:8080 api-load-test.js
```

### stress-test.js

ストレステストスクリプトです。システムの限界を測定します。

- 100VU → 200VU → 300VU → 400VU と段階的に負荷を増加
- 各段階で5分間維持

```bash
# 実行方法
k6 run stress-test.js
```

## しきい値

| メトリクス | しきい値 | 説明 |
|-----------|---------|------|
| http_req_duration (p95) | < 500ms | 95%のリクエストが500ms未満 |
| http_req_duration (p99) | < 1000ms | 99%のリクエストが1秒未満 |
| http_req_failed | < 1% | エラーレート1%未満 |
| login_latency (p95) | < 300ms | ログインが300ms未満 |

## レポート

テスト完了後、以下のファイルが生成されます：

- `summary.json`: 詳細なテスト結果
- `stress-test-report.json`: ストレステストの詳細レポート

## CI/CD統合

GitHub Actionsでの実行例：

```yaml
- name: Run load tests
  uses: grafana/k6-action@v0.3.0
  with:
    filename: tests/load/api-load-test.js
  env:
    BASE_URL: ${{ secrets.API_URL }}
```
