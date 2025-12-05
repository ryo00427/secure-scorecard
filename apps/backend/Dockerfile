# =============================================================================
# Go Backend Dockerfile - マルチステージビルド
# =============================================================================
#
# ビルドステージとランタイムステージを分離して最小イメージを実現します。
# 最終イメージサイズ: ~20MB（alpine base）
#
# ビルド方法:
#   docker build -t secure-scorecard-backend .
#
# 実行方法:
#   docker run -p 8080:8080 --env-file .env secure-scorecard-backend
#
# =============================================================================

# -----------------------------------------------------------------------------
# ステージ1: ビルドステージ
# -----------------------------------------------------------------------------
# Go公式イメージを使用してバイナリをコンパイルします。
# 依存関係キャッシュを活用して再ビルドを高速化します。
FROM golang:1.24-alpine AS builder

# ビルドに必要なツールをインストール
# - git: go mod downloadで必要な場合あり
# - ca-certificates: HTTPS通信用証明書
RUN apk add --no-cache git ca-certificates

# 作業ディレクトリを設定
WORKDIR /app

# 依存関係ファイルを先にコピー（キャッシュ活用）
# go.mod/go.sum が変更されない限り、このレイヤーはキャッシュされます
COPY go.mod go.sum ./

# Go toolchainの自動アップグレードを無効化
# go.mod で指定されたバージョンより新しいGoが必要な場合でも、
# 現在のバージョンでビルドを続行します
ENV GOTOOLCHAIN=local

# 依存関係をダウンロード
# このステップをソースコードコピーより前に行うことで、
# コード変更時の再ビルドが高速化されます
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド
# CGO_ENABLED=0: Cライブラリへの依存をなくし、alpine上で動作可能に
# -ldflags="-s -w": デバッグ情報を削除してバイナリサイズを削減
# -o: 出力ファイル名を指定
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/server \
    ./cmd/server

# -----------------------------------------------------------------------------
# ステージ2: ランタイムステージ
# -----------------------------------------------------------------------------
# 最小限のAlpineイメージを使用して最終イメージを作成します。
# セキュリティのため非rootユーザーで実行します。
FROM alpine:3.19

# 必要なCA証明書をインストール（HTTPS通信用）
# tzdata: タイムゾーン設定用
RUN apk add --no-cache ca-certificates tzdata

# タイムゾーンを設定（日本時間）
ENV TZ=Asia/Tokyo

# 非rootユーザーを作成（セキュリティベストプラクティス）
# UID/GIDは固定値を使用して一貫性を保つ
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -D -H appuser

# 作業ディレクトリを設定
WORKDIR /app

# ビルドステージからバイナリをコピー
COPY --from=builder /app/server /app/server

# 実行ファイルの所有者を変更
RUN chown appuser:appgroup /app/server

# 非rootユーザーに切り替え
USER appuser

# ヘルスチェック設定
# 30秒ごとにヘルスエンドポイントをチェック
# 3回連続で失敗するとunhealthyになる
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# アプリケーションが使用するポートを公開
EXPOSE 8080

# アプリケーションを実行
# execフォームを使用してシグナルを正しく処理
CMD ["/app/server"]
