# 無料デプロイ手順 (Render + Neon + Cloudflare R2)

バックエンドを **Render** (Web Service)、DBを **Neon** (PostgreSQL)、画像ストレージを **Cloudflare R2** (S3互換) に配置してすべて無料枠で運用する手順。

## 全体像

```
[Mobile (Expo)] ── HTTPS ──▶ [Render: Go Backend]
                                   │
                            ┌──────┴──────┐
                            │             │
                            ▼             ▼
                       [Neon: Postgres] [Cloudflare R2]
```

| サービス | 無料枠 | 注意点 |
|---------|--------|--------|
| Render Web Service | 750時間/月、512MB RAM | 15分無アクセスでスリープ（次リクエストで10秒程度の遅延） |
| Neon | 0.5GB ストレージ、コンピュート3GB-時/月 | 5分無アクセスで自動サスペンド |
| Cloudflare R2 | 10GB ストレージ、Class A 100万/月、Class B 1000万/月 | エグレス無料 |

---

## 1. Neon (PostgreSQL) セットアップ

1. https://console.neon.tech にサインアップ
2. **New Project** → Name: `secure-scorecard`、Region: `Asia Pacific (Singapore)` 推奨
3. プロジェクト作成後、ダッシュボードの **Connection string** 欄から **Pooled connection** をコピー
   ```
   postgresql://USER:PASS@ep-xxx-pooler.ap-southeast-1.aws.neon.tech/neondb?sslmode=require
   ```
4. この文字列を後で Render の `DATABASE_URL` に設定する

> **Tip**: Render の free tier ではコネクションプールが小さいため、Neon の **Pooled connection** (PgBouncer経由) を使うこと。

---

## 2. Cloudflare R2 セットアップ

1. https://dash.cloudflare.com → **R2 Object Storage** → サブスクリプション有効化（クレカ登録は必要だが無料枠内なら課金なし）
2. **Create bucket** → 名前: `secure-scorecard`、Location: `APAC` 推奨
3. バケット作成後 → **Settings** タブ → **Public access** → **R2.dev subdomain** を有効化
   - 公開URL `https://pub-xxxxx.r2.dev` が表示される（後で `CLOUDFRONT_URL` に使用）
4. R2 Dashboard 左サイドバー → **Manage R2 API Tokens** → **Create API Token**
   - Permissions: **Object Read & Write**
   - Bucket: 上で作ったバケットを指定
   - 発行された **Access Key ID** / **Secret Access Key** を控える
5. アカウントID（R2 トップページ右上）を控える → エンドポイントURL構築に使う:
   ```
   https://<ACCOUNT_ID>.r2.cloudflarestorage.com
   ```
6. **CORS 設定** (モバイルから直接 PUT する場合): バケット **Settings** → **CORS Policy** に:
   ```json
   [
     {
       "AllowedOrigins": ["*"],
       "AllowedMethods": ["GET", "PUT"],
       "AllowedHeaders": ["*"],
       "MaxAgeSeconds": 3600
     }
   ]
   ```

---

## 3. Render デプロイ

1. https://dashboard.render.com にサインアップ（GitHub連携推奨）
2. **New +** → **Blueprint** → このリポジトリを選択
3. Render が `render.yaml` を自動検出 → `secure-scorecard-backend` Web Service が作成される
4. **Environment** タブで `sync: false` の env 値を入力:

   | キー | 値 |
   |------|---|
   | `DATABASE_URL` | Neon の Pooled connection string |
   | `S3_BUCKET_NAME` | `secure-scorecard` |
   | `S3_ENDPOINT` | `https://<ACCOUNT_ID>.r2.cloudflarestorage.com` |
   | `AWS_ACCESS_KEY_ID` | R2 で発行した Access Key ID |
   | `AWS_SECRET_ACCESS_KEY` | R2 で発行した Secret Access Key |
   | `CLOUDFRONT_URL` | `https://pub-xxxxx.r2.dev` |

5. **Manual Deploy** → **Deploy latest commit**
6. ログを監視。`Database connected successfully` と migration 完了が見えれば成功
7. Render が割り当てた URL `https://secure-scorecard-backend-xxxx.onrender.com` を控える

### 動作確認

```bash
# ヘルスチェック
curl https://<your-render-url>.onrender.com/health
# → {"status":"ok"}

# DB込みヘルスチェック
curl https://<your-render-url>.onrender.com/health/db
# → {"status":"healthy", "stats":{...}}

# ユーザー登録
curl -X POST https://<your-render-url>.onrender.com/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Password123!","name":"Test"}'
```

---

## 4. モバイルアプリ更新

`apps/mobile/.env` を編集:

```
EXPO_PUBLIC_API_URL=https://<your-render-url>.onrender.com/api/v1
```

その後 Expo を再起動:

```bash
cd apps/mobile
pnpm start --clear
```

---

## トラブルシューティング

| 症状 | 原因 / 対処 |
|------|-------------|
| Render ビルド失敗 `Dockerfile not found` | render.yaml の `dockerfilePath` と `dockerContext` を確認 |
| `connection refused` (DB) | Neon が suspend 中。1-2秒待って再リクエスト |
| 画像 PUT で 403 | R2 API Token の権限が Read のみ。Read & Write に変更 |
| 画像 PUT で CORS エラー | R2 バケットの CORS Policy 未設定 |
| 初回リクエストが10秒以上かかる | Render free tier のコールドスタート（仕様） |
| Neon コンピュート時間超過 | スケジューラー等の常時クエリを停止、または有料プランへ |

---

## 通知機能を後で有効化する場合

現状 SNS/SES は env 未設定で no-op 動作している。有効化する場合は別タスクで以下を実装:

- **プッシュ通知**: `internal/service/notification_sender.go` の SNS 呼び出しを Expo Push API (`https://exp.host/--/api/v2/push/send`) に置き換え
- **メール**: SES 呼び出しを Resend (`https://api.resend.com/emails`) に置き換え

モバイル側は既に Expo Push Token 取得 → バックエンドへ登録のフローが実装済み (`apps/mobile/src/services/notifications.ts`)。
