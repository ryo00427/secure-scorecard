// =============================================================================
// Performance Tests - パフォーマンステスト
// =============================================================================
// APIスループット、DBクエリ、同時アクセスのテスト

package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/secure-scorecard/backend/internal/auth"
	"github.com/secure-scorecard/backend/internal/model"
	"github.com/secure-scorecard/backend/internal/repository"
	"github.com/secure-scorecard/backend/internal/service"
	"github.com/secure-scorecard/backend/internal/validator"
)

// =============================================================================
// パフォーマンステスト用のセットアップ
// =============================================================================

// performanceTestSetup はパフォーマンステスト用のセットアップ構造体です。
type performanceTestSetup struct {
	echo        *echo.Echo
	mockRepos   *repository.MockRepositories
	service     *service.Service
	jwtManager  *auth.JWTManager
	authHandler *AuthHandler
	handler     *Handler
	authToken   string
	testUserID  uint
}

// newPerformanceTestSetup はパフォーマンステスト用のセットアップを作成します。
func newPerformanceTestSetup(t *testing.T) *performanceTestSetup {
	e := echo.New()
	e.Validator = validator.NewValidator()
	mockRepos := repository.NewMockRepositories()
	svc := service.NewService(mockRepos)
	jwtManager := auth.NewJWTManager("performance-test-secret-key-32ch", 24)
	authHandler := NewAuthHandler(svc, jwtManager)
	handler := NewHandler(svc, jwtManager, nil)

	// テストユーザーを作成
	testUserID := uint(1)

	// モックリポジトリにユーザーを設定
	mockRepos.GetMockUserRepository().Users[testUserID] = &model.User{
		Email:        "perf-test@example.com",
		PasswordHash: "$2a$10$test-hash",
		DisplayName:  "Performance Test User",
		IsActive:     true,
	}
	mockRepos.GetMockUserRepository().Users[testUserID].ID = testUserID

	// 認証トークンを生成
	token, _ := jwtManager.GenerateToken(testUserID, "", "perf-test@example.com")

	return &performanceTestSetup{
		echo:        e,
		mockRepos:   mockRepos,
		service:     svc,
		jwtManager:  jwtManager,
		authHandler: authHandler,
		handler:     handler,
		authToken:   token,
		testUserID:  testUserID,
	}
}

// =============================================================================
// APIスループットテスト
// =============================================================================

// TestAPIThroughput はAPIスループットをテストします。
// 目標: 高速なレスポンス、エラーなし
func TestAPIThroughput(t *testing.T) {
	setup := newPerformanceTestSetup(t)

	// テストパラメータ
	numRequests := 100

	// レイテンシ計測用
	var totalLatency int64
	var maxLatency int64
	var successCount int64

	startTime := time.Now()

	// 並列でリクエストを送信
	var wg sync.WaitGroup
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			reqStart := time.Now()

			req := httptest.NewRequest(http.MethodGet, "/api/v1/gardens", nil)
			req.Header.Set("Authorization", "Bearer "+setup.authToken)
			rec := httptest.NewRecorder()

			c := setup.echo.NewContext(req, rec)
			// Set user claims in context (required by auth.GetUserIDFromContext)
			c.Set("user", &auth.Claims{UserID: setup.testUserID, Email: "perf-test@example.com"})

			err := setup.handler.GetGardens(c)
			latency := time.Since(reqStart).Nanoseconds()

			if err == nil && rec.Code == http.StatusOK {
				atomic.AddInt64(&successCount, 1)
			}

			atomic.AddInt64(&totalLatency, latency)

			// 最大レイテンシを更新
			for {
				current := atomic.LoadInt64(&maxLatency)
				if latency <= current || atomic.CompareAndSwapInt64(&maxLatency, current, latency) {
					break
				}
			}
		}()
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// メトリクスを計算
	avgLatency := time.Duration(totalLatency / int64(numRequests))
	maxLatencyDuration := time.Duration(maxLatency)
	throughput := float64(numRequests) / totalDuration.Seconds()

	// 結果を出力
	t.Logf("=== APIスループットテスト結果 ===")
	t.Logf("総リクエスト数: %d", numRequests)
	t.Logf("成功リクエスト数: %d", successCount)
	t.Logf("総実行時間: %v", totalDuration)
	t.Logf("スループット: %.2f req/sec", throughput)
	t.Logf("平均レイテンシ: %v", avgLatency)
	t.Logf("最大レイテンシ: %v", maxLatencyDuration)

	// アサーション
	assert.Equal(t, int64(numRequests), successCount, "全リクエストが成功すること")
	assert.Less(t, avgLatency, 200*time.Millisecond, "平均レイテンシが200ms未満であること")
}

// =============================================================================
// 同時アクセステスト
// =============================================================================

// TestConcurrentAccess は同時アクセス時の整合性をテストします。
func TestConcurrentAccess(t *testing.T) {
	setup := newPerformanceTestSetup(t)

	// 同時アクセスのパラメータ
	numConcurrentUsers := 10
	operationsPerUser := 10

	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64

	startTime := time.Now()

	// 複数ユーザーによる同時アクセスをシミュレート
	for userIdx := 0; userIdx < numConcurrentUsers; userIdx++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for opIdx := 0; opIdx < operationsPerUser; opIdx++ {
				// 読み取り操作
				req := httptest.NewRequest(http.MethodGet, "/api/v1/gardens", nil)
				req.Header.Set("Authorization", "Bearer "+setup.authToken)
				rec := httptest.NewRecorder()
				c := setup.echo.NewContext(req, rec)
				c.Set("user", &auth.Claims{UserID: setup.testUserID, Email: "perf-test@example.com"})

				err := setup.handler.GetGardens(c)

				if err != nil || rec.Code != http.StatusOK {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	totalOperations := int64(numConcurrentUsers * operationsPerUser)

	t.Logf("=== 同時アクセステスト結果 ===")
	t.Logf("同時ユーザー数: %d", numConcurrentUsers)
	t.Logf("ユーザーあたりの操作数: %d", operationsPerUser)
	t.Logf("総操作数: %d", totalOperations)
	t.Logf("成功: %d, エラー: %d", successCount, errorCount)
	t.Logf("総実行時間: %v", totalDuration)
	t.Logf("スループット: %.2f ops/sec", float64(totalOperations)/totalDuration.Seconds())

	// アサーション
	assert.Equal(t, totalOperations, successCount, "全操作が成功すること")
	assert.Equal(t, int64(0), errorCount, "エラーが発生しないこと")
}

// =============================================================================
// 書き込み競合テスト
// =============================================================================

// TestConcurrentWrites は同時書き込み時の整合性をテストします。
// 注意: このテストはスレッドセーフなリポジトリ実装が必要です。
// MockRepository はマップを使用しており並行書き込みに対応していないためスキップします。
func TestConcurrentWrites(t *testing.T) {
	t.Skip("MockRepository does not support concurrent writes - requires thread-safe implementation")
	setup := newPerformanceTestSetup(t)

	numConcurrentWriters := 5
	writesPerWriter := 5

	var wg sync.WaitGroup
	var successCount int64

	startTime := time.Now()

	for writerIdx := 0; writerIdx < numConcurrentWriters; writerIdx++ {
		wg.Add(1)
		go func(writerId int) {
			defer wg.Done()

			for writeIdx := 0; writeIdx < writesPerWriter; writeIdx++ {
				cropData := map[string]interface{}{
					"name":                  fmt.Sprintf("Crop-%d-%d", writerId, writeIdx),
					"variety":               "Test Variety",
					"planted_date":          time.Now().Format(time.RFC3339),
					"expected_harvest_date": time.Now().AddDate(0, 2, 0).Format(time.RFC3339),
				}

				jsonData, _ := json.Marshal(cropData)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/crops", bytes.NewReader(jsonData))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+setup.authToken)
				rec := httptest.NewRecorder()
				c := setup.echo.NewContext(req, rec)
				c.Set("user", &auth.Claims{UserID: setup.testUserID, Email: "perf-test@example.com"})

				err := setup.handler.CreateCrop(c)

				if err == nil && rec.Code == http.StatusCreated {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(writerIdx)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	totalWrites := int64(numConcurrentWriters * writesPerWriter)

	t.Logf("=== 同時書き込みテスト結果 ===")
	t.Logf("同時ライター数: %d", numConcurrentWriters)
	t.Logf("ライターあたりの書き込み数: %d", writesPerWriter)
	t.Logf("総書き込み数: %d", totalWrites)
	t.Logf("成功した書き込み: %d", successCount)
	t.Logf("総実行時間: %v", totalDuration)

	// アサーション
	assert.Equal(t, totalWrites, successCount, "全書き込みが成功すること")
}

// =============================================================================
// タイムアウトテスト
// =============================================================================

// TestRequestTimeout はリクエストタイムアウトをテストします。
func TestRequestTimeout(t *testing.T) {
	setup := newPerformanceTestSetup(t)

	// 通常のリクエストがタイムアウトしないことを確認
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gardens", nil)
	req.Header.Set("Authorization", "Bearer "+setup.authToken)
	rec := httptest.NewRecorder()
	c := setup.echo.NewContext(req, rec)
	c.Set("user", &auth.Claims{UserID: setup.testUserID, Email: "perf-test@example.com"})

	start := time.Now()
	err := setup.handler.GetGardens(c)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Less(t, duration, 5*time.Second, "リクエストがタイムアウト前に完了すること")

	t.Logf("リクエスト完了時間: %v", duration)
}

// =============================================================================
// ベンチマークテスト
// =============================================================================

// BenchmarkGetGardens は菜園取得APIのベンチマークです。
func BenchmarkGetGardens(b *testing.B) {
	e := echo.New()
	e.Validator = validator.NewValidator()
	mockRepos := repository.NewMockRepositories()
	svc := service.NewService(mockRepos)
	jwtManager := auth.NewJWTManager("benchmark-test-secret-key-32ch", 24)
	handler := NewHandler(svc, jwtManager, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/gardens", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user", &auth.Claims{UserID: uint(1), Email: "benchmark@example.com"})

		handler.GetGardens(c)
	}
}

// BenchmarkCreateCrop は作物作成APIのベンチマークです。
func BenchmarkCreateCrop(b *testing.B) {
	e := echo.New()
	e.Validator = validator.NewValidator()
	mockRepos := repository.NewMockRepositories()
	svc := service.NewService(mockRepos)
	jwtManager := auth.NewJWTManager("benchmark-test-secret-key-32ch", 24)
	handler := NewHandler(svc, jwtManager, nil)

	cropData := map[string]interface{}{
		"name":                  "Test Crop",
		"variety":               "Test Variety",
		"planted_date":          time.Now().Format(time.RFC3339),
		"expected_harvest_date": time.Now().AddDate(0, 2, 0).Format(time.RFC3339),
	}
	jsonData, _ := json.Marshal(cropData)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crops", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user", &auth.Claims{UserID: uint(1), Email: "benchmark@example.com"})

		handler.CreateCrop(c)
	}
}
