
.PHONY: help dev db-up db-down db-logs clean install build test lint type-check docker-up docker-down docker-logs docker-clean

# デフォルトターゲット: ヘルプを表示
help:
	@echo "=================================="
	@echo "  家庭菜園管理アプリ - Makefile"
	@echo "=================================="
	@echo ""
	@echo "開発環境:"
	@echo "  make dev          - 開発環境全体を起動 (DB + Turborepo)"
	@echo "  make install      - 依存パッケージをインストール"
	@echo ""
	@echo "データベース:"
	@echo "  make db-up        - PostgreSQLコンテナを起動"
	@echo "  make db-down      - PostgreSQLコンテナを停止"
	@echo "  make db-logs      - PostgreSQLログを表示"
	@echo "  make db-shell     - PostgreSQLシェルに接続"
	@echo ""
	@echo "Docker Compose:"
	@echo "  make docker-up    - 全コンテナを起動 (DB + API)"
	@echo "  make docker-down  - 全コンテナを停止"
	@echo "  make docker-logs  - 全コンテナのログを表示"
	@echo "  make docker-clean - コンテナ/ボリューム/イメージを削除"
	@echo ""
	@echo "ビルド/テスト:"
	@echo "  make build        - 全パッケージをビルド"
	@echo "  make test         - 全テストを実行"
	@echo "  make lint         - Lintを実行"
	@echo "  make type-check   - 型チェックを実行"
	@echo ""
	@echo "クリーンアップ:"
	@echo "  make clean        - 全コンテナを停止してクリーンアップ"
	@echo ""

# 開発環境全体を起動（PostgreSQL + Turborepo）
dev: db-up
	@echo "Starting development environment..."
	pnpm dev

# 依存パッケージをインストール
install:
	@echo "Installing dependencies..."
	pnpm install

# PostgreSQLコンテナのみ起動
db-up:
	@echo "Starting PostgreSQL container..."
	docker compose up -d db
	@echo "Waiting for database to be ready..."
	@timeout 10 >nul 2>&1 || sleep 10
	@echo "PostgreSQL is ready!"

# PostgreSQLコンテナを停止
db-down:
	@echo "Stopping PostgreSQL container..."
	docker compose stop db

# PostgreSQLログを表示
db-logs:
	docker compose logs -f db

# PostgreSQLシェルに接続
db-shell:
	docker compose exec db psql -U postgres -d home_garden

# 全コンテナを起動（API + DB）
docker-up:
	@echo "Starting all containers..."
	docker compose up -d
	@echo "Containers started!"
	docker compose ps

# 全コンテナを停止
docker-down:
	@echo "Stopping all containers..."
	docker compose down

# 全コンテナのログを表示
docker-logs:
	docker compose logs -f

# コンテナ/ボリューム/イメージを完全削除
docker-clean:
	@echo "Cleaning up Docker resources..."
	docker compose down -v --rmi all
	@echo "Cleanup complete!"

# 全パッケージをビルド
build:
	@echo "Building all packages..."
	pnpm build

# 全テストを実行
test:
	@echo "Running all tests..."
	pnpm test

# Lintを実行
lint:
	@echo "Running lint..."
	pnpm lint

# 型チェックを実行
type-check:
	@echo "Running type check..."
	pnpm type-check

# 全コンテナを停止してクリーンアップ
clean: docker-down
	@echo "Cleaning up..."
	pnpm clean || true
	@echo "Cleanup complete!"
