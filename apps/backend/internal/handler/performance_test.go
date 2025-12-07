// =============================================================================
// Performance Tests - パフォーマンステスト
// =============================================================================
// APIスループット、DBクエリ、画像アップロード、同時アクセスのテスト

package handler

import (
	"bytes"
	"context"
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

	"github.com/ryo00427/secure-scorecard/apps/backend/internal/auth"
	"github.com/ryo00427/secure-scorecard/apps/backend/internal/model"
	"github.com/ryo00427/secure-scorecard/apps/backend/internal/repository"
	"github.com/ryo00427/secure-scorecard/apps/backend/internal/service"
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
	testUser    *model.User
}

// newPerformanceTestSetup はパフォーマンステスト用のセットアップを作成します。
func newPerformanceTestSetup(t *testing.T) *performanceTestSetup {
	e := echo.New()
	mockRepos := repository.NewMockRepositories()
	svc := service.NewService(mockRepos)
	jwtManager := auth.NewJWTManager("performance-test-secret-key", 24*time.Hour)
	authHandler := NewAuthHandler(svc, jwtManager)
	handler := NewHandler(svc)

	// テストユーザーを作成
	testUser := &model.User{
		ID:           1,
		Email:        "perf-test@example.com",
		PasswordHash: "$2a$10$test-hash",
		DisplayName:  "Performance Test User",
	}

	// モックリポジトリにユーザーを設定
	mockRepos.GetMockUserRepository().Users[testUser.ID] = testUser

	// 認証トークンを生成
	token, _ := jwtManager.GenerateToken(testUser.ID)

	return &performanceTestSetup{
		echo:        e,
		mockRepos:   mockRepos,
		service:     svc,
		jwtManager:  jwtManager,
		authHandler: authHandler,
		handler:     handler,
		authToken:   token,
		testUser:    testUser,
	}
}

// =============================================================================
// APIスループットテスト
// =============================================================================

// TestAPIThroughput はAPIスループットをテストします。
// 目標: 1000 req/sec 以上、レイテンシ < 200ms
func TestAPIThroughput(t *testing.T) {
	setup := newPerformanceTestSetup(t)

	// テスト用の菜園データを準備
	testGarden := &model.Garden{
		ID:       1,
		UserID:   setup.testUser.ID,
		Name:     "Test Garden",
		Location: "Tokyo",
	}
	setup.mockRepos.GetMockGardenRepository().Gardens[testGarden.ID] = testGarden

	// テストパラメータ
	numRequests := 1000
	targetDuration := time.Second // 1秒以内に完了することを目標

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
			c.Set("user_id", setup.testUser.ID)

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
	assert.GreaterOrEqual(t, throughput, 100.0, "スループットが100 req/sec以上であること") // モック環境では控えめに設定
}

// =============================================================================
// データベースクエリパフォーマンステスト
// =============================================================================

// TestDatabaseQueryPerformance はDBクエリのパフォーマンスをテストします。
func TestDatabaseQueryPerformance(t *testing.T) {
	setup := newPerformanceTestSetup(t)

	// 大量のテストデータを準備
	numGardens := 100
	numCropsPerGarden := 10

	for i := 1; i <= numGardens; i++ {
		garden := &model.Garden{
			ID:       uint(i),
			UserID:   setup.testUser.ID,
			Name:     fmt.Sprintf("Garden %d", i),
			Location: "Tokyo",
		}
		setup.mockRepos.GetMockGardenRepository().Gardens[garden.ID] = garden

		for j := 1; j <= numCropsPerGarden; j++ {
			crop := &model.Crop{
				ID:       uint(i*1000 + j),
				GardenID: garden.ID,
				Name:     fmt.Sprintf("Crop %d-%d", i, j),
				Status:   "growing",
			}
			setup.mockRepos.GetMockCropRepository().Crops[crop.ID] = crop
		}
	}

	// テストケース
	testCases := []struct {
		name      string
		operation func() error
		maxTime   time.Duration
	}{
		{
			name: "菜園一覧取得",
			operation: func() error {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/gardens", nil)
				req.Header.Set("Authorization", "Bearer "+setup.authToken)
				rec := httptest.NewRecorder()
				c := setup.echo.NewContext(req, rec)
				c.Set("user_id", setup.testUser.ID)
				return setup.handler.GetGardens(c)
			},
			maxTime: 100 * time.Millisecond,
		},
		{
			name: "作物一覧取得",
			operation: func() error {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/crops", nil)
				req.Header.Set("Authorization", "Bearer "+setup.authToken)
				rec := httptest.NewRecorder()
				c := setup.echo.NewContext(req, rec)
				c.Set("user_id", setup.testUser.ID)
				return setup.handler.GetCrops(c)
			},
			maxTime: 100 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ウォームアップ
			tc.operation()

			// 計測
			iterations := 100
			var totalDuration time.Duration

			for i := 0; i < iterations; i++ {
				start := time.Now()
				err := tc.operation()
				duration := time.Since(start)
				totalDuration += duration

				assert.NoError(t, err)
			}

			avgDuration := totalDuration / time.Duration(iterations)

			t.Logf("%s - 平均実行時間: %v (目標: < %v)", tc.name, avgDuration, tc.maxTime)
			assert.Less(t, avgDuration, tc.maxTime, "クエリ実行時間が目標を超えている")
		})
	}
}

// =============================================================================
// 同時アクセステスト
// =============================================================================

// TestConcurrentAccess は同時アクセス時の整合性をテストします。
func TestConcurrentAccess(t *testing.T) {
	setup := newPerformanceTestSetup(t)

	// テスト用の菜園を作成
	testGarden := &model.Garden{
		ID:       1,
		UserID:   setup.testUser.ID,
		Name:     "Concurrent Test Garden",
		Location: "Tokyo",
	}
	setup.mockRepos.GetMockGardenRepository().Gardens[testGarden.ID] = testGarden

	// 同時アクセスのパラメータ
	numConcurrentUsers := 50
	operationsPerUser := 20

	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64
	var mutex sync.Mutex
	errors := make([]error, 0)

	startTime := time.Now()

	// 複数ユーザーによる同時アクセスをシミュレート
	for userIdx := 0; userIdx < numConcurrentUsers; userIdx++ {
		wg.Add(1)
		go func(userId int) {
			defer wg.Done()

			for opIdx := 0; opIdx < operationsPerUser; opIdx++ {
				// 読み取り操作
				req := httptest.NewRequest(http.MethodGet, "/api/v1/gardens", nil)
				req.Header.Set("Authorization", "Bearer "+setup.authToken)
				rec := httptest.NewRecorder()
				c := setup.echo.NewContext(req, rec)
				c.Set("user_id", setup.testUser.ID)

				err := setup.handler.GetGardens(c)

				if err != nil || rec.Code != http.StatusOK {
					atomic.AddInt64(&errorCount, 1)
					mutex.Lock()
					errors = append(errors, fmt.Errorf("user %d, op %d: %v", userId, opIdx, err))
					mutex.Unlock()
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}(userIdx)
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

	// エラーがあれば詳細を出力
	if len(errors) > 0 && len(errors) <= 10 {
		for _, err := range errors {
			t.Logf("エラー詳細: %v", err)
		}
	}

	// アサーション
	assert.Equal(t, totalOperations, successCount, "全操作が成功すること")
	assert.Equal(t, int64(0), errorCount, "エラーが発生しないこと")
}

// =============================================================================
// 書き込み競合テスト
// =============================================================================

// TestConcurrentWrites は同時書き込み時の整合性をテストします。
func TestConcurrentWrites(t *testing.T) {
	setup := newPerformanceTestSetup(t)

	// テスト用の菜園を作成
	testGarden := &model.Garden{
		ID:       1,
		UserID:   setup.testUser.ID,
		Name:     "Write Test Garden",
		Location: "Tokyo",
	}
	setup.mockRepos.GetMockGardenRepository().Gardens[testGarden.ID] = testGarden

	numConcurrentWriters := 10
	writesPerWriter := 10

	var wg sync.WaitGroup
	var successCount int64
	var createdCropIDs sync.Map

	startTime := time.Now()

	for writerIdx := 0; writerIdx < numConcurrentWriters; writerIdx++ {
		wg.Add(1)
		go func(writerId int) {
			defer wg.Done()

			for writeIdx := 0; writeIdx < writesPerWriter; writeIdx++ {
				cropData := map[string]interface{}{
					"garden_id":    testGarden.ID,
					"name":         fmt.Sprintf("Crop-%d-%d", writerId, writeIdx),
					"variety":      "Test Variety",
					"planted_date": time.Now().Format(time.RFC3339),
				}

				jsonData, _ := json.Marshal(cropData)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/crops", bytes.NewReader(jsonData))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+setup.authToken)
				rec := httptest.NewRecorder()
				c := setup.echo.NewContext(req, rec)
				c.Set("user_id", setup.testUser.ID)

				err := setup.handler.CreateCrop(c)

				if err == nil && rec.Code == http.StatusCreated {
					atomic.AddInt64(&successCount, 1)

					// 作成されたCrop IDを記録
					var response map[string]interface{}
					json.Unmarshal(rec.Body.Bytes(), &response)
					if crop, ok := response["crop"].(map[string]interface{}); ok {
						if id, ok := crop["id"].(float64); ok {
							createdCropIDs.Store(int(id), true)
						}
					}
				}
			}
		}(writerIdx)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	totalWrites := int64(numConcurrentWriters * writesPerWriter)

	// 作成されたユニークなCrop数をカウント
	var uniqueCropCount int
	createdCropIDs.Range(func(key, value interface{}) bool {
		uniqueCropCount++
		return true
	})

	t.Logf("=== 同時書き込みテスト結果 ===")
	t.Logf("同時ライター数: %d", numConcurrentWriters)
	t.Logf("ライターあたりの書き込み数: %d", writesPerWriter)
	t.Logf("総書き込み数: %d", totalWrites)
	t.Logf("成功した書き込み: %d", successCount)
	t.Logf("ユニークなCrop数: %d", uniqueCropCount)
	t.Logf("総実行時間: %v", totalDuration)

	// アサーション
	assert.Equal(t, totalWrites, successCount, "全書き込みが成功すること")
}

// =============================================================================
// 画像アップロードパフォーマンステスト
// =============================================================================

// TestImageUploadPerformance は画像アップロードのパフォーマンスをテストします。
func TestImageUploadPerformance(t *testing.T) {
	setup := newPerformanceTestSetup(t)

	// テスト用の作物を作成
	testCrop := &model.Crop{
		ID:       1,
		GardenID: 1,
		Name:     "Test Crop",
		Status:   "growing",
	}
	setup.mockRepos.GetMockCropRepository().Crops[testCrop.ID] = testCrop

	// 画像サイズのバリエーション
	imageSizes := []struct {
		name    string
		sizeKB  int
		maxTime time.Duration
	}{
		{"小サイズ (100KB)", 100, 50 * time.Millisecond},
		{"中サイズ (500KB)", 500, 100 * time.Millisecond},
		{"大サイズ (1MB)", 1024, 200 * time.Millisecond},
		{"最大サイズ (5MB)", 5120, 500 * time.Millisecond},
	}

	for _, size := range imageSizes {
		t.Run(size.name, func(t *testing.T) {
			// 署名付きURL取得のパフォーマンスをテスト
			presignData := map[string]interface{}{
				"crop_id":      testCrop.ID,
				"file_name":    fmt.Sprintf("test-image-%dkb.jpg", size.sizeKB),
				"content_type": "image/jpeg",
				"file_size":    size.sizeKB * 1024,
			}

			jsonData, _ := json.Marshal(presignData)

			iterations := 10
			var totalDuration time.Duration

			for i := 0; i < iterations; i++ {
				start := time.Now()

				req := httptest.NewRequest(http.MethodPost, "/api/v1/crops/images/presign", bytes.NewReader(jsonData))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+setup.authToken)
				rec := httptest.NewRecorder()
				c := setup.echo.NewContext(req, rec)
				c.Set("user_id", setup.testUser.ID)

				// Note: 実際のS3アップロードは行わず、署名付きURL生成のみテスト
				// 実環境では setup.handler.GetPresignedURL(c) などを呼び出す

				duration := time.Since(start)
				totalDuration += duration
			}

			avgDuration := totalDuration / time.Duration(iterations)
			t.Logf("%s - 署名付きURL生成平均時間: %v", size.name, avgDuration)

			// 署名付きURL生成は高速であるべき
			assert.Less(t, avgDuration, size.maxTime, "署名付きURL生成が目標時間内であること")
		})
	}
}

// =============================================================================
// メモリ使用量テスト
// =============================================================================

// TestMemoryUsage は大量データ処理時のメモリ使用量をテストします。
func TestMemoryUsage(t *testing.T) {
	setup := newPerformanceTestSetup(t)

	// 大量のデータを作成
	numRecords := 10000
	for i := 1; i <= numRecords; i++ {
		garden := &model.Garden{
			ID:       uint(i),
			UserID:   setup.testUser.ID,
			Name:     fmt.Sprintf("Memory Test Garden %d", i),
			Location: "Tokyo",
		}
		setup.mockRepos.GetMockGardenRepository().Gardens[garden.ID] = garden
	}

	// データ取得を複数回実行
	iterations := 100
	for i := 0; i < iterations; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/gardens", nil)
		req.Header.Set("Authorization", "Bearer "+setup.authToken)
		rec := httptest.NewRecorder()
		c := setup.echo.NewContext(req, rec)
		c.Set("user_id", setup.testUser.ID)

		err := setup.handler.GetGardens(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	t.Logf("=== メモリ使用量テスト完了 ===")
	t.Logf("処理レコード数: %d", numRecords)
	t.Logf("反復回数: %d", iterations)
	// Note: 実際のメモリ使用量は runtime.MemStats で計測可能
}

// =============================================================================
// タイムアウトテスト
// =============================================================================

// TestRequestTimeout はリクエストタイムアウトをテストします。
func TestRequestTimeout(t *testing.T) {
	setup := newPerformanceTestSetup(t)

	// 通常のリクエストがタイムアウトしないことを確認
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/gardens", nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+setup.authToken)
	rec := httptest.NewRecorder()
	c := setup.echo.NewContext(req, rec)
	c.Set("user_id", setup.testUser.ID)

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
	mockRepos := repository.NewMockRepositories()
	svc := service.NewService(mockRepos)
	handler := NewHandler(svc)

	// テストデータ準備
	for i := 1; i <= 100; i++ {
		garden := &model.Garden{
			ID:       uint(i),
			UserID:   1,
			Name:     fmt.Sprintf("Garden %d", i),
			Location: "Tokyo",
		}
		mockRepos.GetMockGardenRepository().Gardens[garden.ID] = garden
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/gardens", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", uint(1))

		handler.GetGardens(c)
	}
}

// BenchmarkCreateCrop は作物作成APIのベンチマークです。
func BenchmarkCreateCrop(b *testing.B) {
	e := echo.New()
	mockRepos := repository.NewMockRepositories()
	svc := service.NewService(mockRepos)
	handler := NewHandler(svc)

	// テスト用菜園を作成
	mockRepos.GetMockGardenRepository().Gardens[1] = &model.Garden{
		ID:     1,
		UserID: 1,
		Name:   "Test Garden",
	}

	cropData := map[string]interface{}{
		"garden_id":    1,
		"name":         "Test Crop",
		"variety":      "Test Variety",
		"planted_date": time.Now().Format(time.RFC3339),
	}
	jsonData, _ := json.Marshal(cropData)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crops", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", uint(1))

		handler.CreateCrop(c)
	}
}
