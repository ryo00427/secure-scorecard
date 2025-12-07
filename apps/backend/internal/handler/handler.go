package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/auth"
	"github.com/secure-scorecard/backend/internal/service"
	"github.com/secure-scorecard/backend/internal/storage"
)

// Handler holds all HTTP handlers
type Handler struct {
	service    *service.Service
	jwtManager *auth.JWTManager
	s3Service  *storage.S3Service
}

// NewHandler creates a new Handler instance
func NewHandler(svc *service.Service, jwtManager *auth.JWTManager, s3Svc *storage.S3Service) *Handler {
	return &Handler{
		service:    svc,
		jwtManager: jwtManager,
		s3Service:  s3Svc,
	}
}

// RegisterRoutes registers all routes
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	// Health check (public)
	e.GET("/health", h.Health)
	e.GET("/", h.Hello)

	// API v1 group
	api := e.Group("/api/v1")

	// Auth endpoints (public)
	authHandler := NewAuthHandler(h.service, h.jwtManager)
	authGroup := api.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/firebase-login", authHandler.FirebaseLogin)
	authGroup.POST("/logout", authHandler.Logout)

	// Protected auth endpoints
	authProtected := authGroup.Group("")
	authProtected.Use(auth.AuthMiddleware(h.jwtManager, h.service))
	authProtected.POST("/refresh", authHandler.RefreshToken)
	authProtected.GET("/me", authHandler.Me)

	// Protected API endpoints
	protected := api.Group("")
	protected.Use(auth.AuthMiddleware(h.jwtManager, h.service))

	// Gardens endpoints (protected)
	gardens := protected.Group("/gardens")
	gardens.GET("", h.GetGardens)
	gardens.POST("", h.CreateGarden)
	gardens.GET("/:id", h.GetGarden)
	gardens.PUT("/:id", h.UpdateGarden)
	gardens.DELETE("/:id", h.DeleteGarden)

	// Plants endpoints (nested under gardens, protected)
	gardens.GET("/:id/plants", h.GetGardenPlants)
	gardens.POST("/:id/plants", h.CreatePlant)

	// Plants endpoints (direct access, protected)
	plants := protected.Group("/plants")
	plants.GET("/:id", h.GetPlant)
	plants.PUT("/:id", h.UpdatePlant)
	plants.DELETE("/:id", h.DeletePlant)

	// Care logs endpoints (nested under plants, protected)
	plants.GET("/:id/care-logs", h.GetPlantCareLogs)
	plants.POST("/:id/care-logs", h.CreateCareLog)

	// User endpoints (protected)
	users := protected.Group("/users")
	users.GET("/me", h.GetCurrentUser)

	// Task endpoints (protected)
	// タスク管理エンドポイント - やることリストのCRUD操作
	tasks := protected.Group("/tasks")
	tasks.GET("", h.GetTasks)                   // 全タスク取得（statusクエリパラメータでフィルタ可能）
	tasks.GET("/today", h.GetTodayTasks)        // 今日のタスク取得
	tasks.GET("/overdue", h.GetOverdueTasks)    // 期限切れタスク取得
	tasks.POST("", h.CreateTask)                // 新規タスク作成
	tasks.GET("/:id", h.GetTask)                // 特定タスク取得
	tasks.PUT("/:id", h.UpdateTask)             // タスク更新
	tasks.DELETE("/:id", h.DeleteTask)          // タスク削除
	tasks.POST("/:id/complete", h.CompleteTask) // タスク完了

	// Crop endpoints (protected)
	// 作物管理エンドポイント - 作物の植え付けから収穫までのライフサイクル管理
	crops := protected.Group("/crops")
	crops.GET("", h.GetCrops)        // 全作物取得（statusクエリパラメータでフィルタ可能）
	crops.POST("", h.CreateCrop)     // 新規作物登録
	crops.GET("/:id", h.GetCrop)     // 特定作物取得
	crops.PUT("/:id", h.UpdateCrop)  // 作物更新
	crops.DELETE("/:id", h.DeleteCrop) // 作物削除

	// Image upload endpoints (nested under crops)
	// 画像アップロードエンドポイント - S3 Presigned URL生成・直接アップロード
	crops.POST("/images/presign", h.GenerateImageUploadURL) // Presigned URL生成（クライアント直接アップロード用）
	crops.POST("/images", h.UploadImage)                     // サーバー経由アップロード（multipart/form-data）

	// Growth records endpoints (nested under crops)
	// 成長記録エンドポイント - 作物の成長観察記録
	crops.GET("/:id/growth-records", h.GetGrowthRecords)   // 成長記録一覧取得
	crops.POST("/:id/growth-records", h.CreateGrowthRecord) // 成長記録追加

	// Harvest endpoints (nested under crops)
	// 収穫記録エンドポイント - 収穫量と品質の記録
	crops.GET("/:id/harvests", h.GetHarvests)   // 収穫記録一覧取得
	crops.POST("/:id/harvests", h.CreateHarvest) // 収穫記録追加

	// Plot endpoints (protected)
	// 区画管理エンドポイント - 菜園のグリッドレイアウト管理
	plots := protected.Group("/plots")
	plots.GET("", h.GetPlots)         // 全区画取得（statusクエリパラメータでフィルタ可能）
	plots.POST("", h.CreatePlot)      // 新規区画作成
	plots.GET("/layout", h.GetPlotLayout) // 全区画のレイアウトデータ取得（グリッド表示用）
	plots.GET("/:id", h.GetPlot)      // 特定区画取得
	plots.PUT("/:id", h.UpdatePlot)   // 区画更新
	plots.DELETE("/:id", h.DeletePlot) // 区画削除

	// Plot assignment endpoints (nested under plots)
	// 区画配置エンドポイント - 作物の配置管理
	plots.POST("/:id/assign", h.AssignCrop)               // 作物を区画に配置
	plots.DELETE("/:id/assign", h.UnassignCrop)           // 配置解除
	plots.GET("/:id/assignments", h.GetPlotAssignments)   // 配置履歴取得
	plots.GET("/:id/assignment", h.GetActivePlotAssignment) // アクティブな配置取得
	plots.GET("/:id/history", h.GetPlotHistory) // 区画の栽培履歴取得（作物情報付き）

	// Analytics endpoints (protected)
	// 分析データエンドポイント - 収穫量・成長データなどの集計・分析
	analytics := protected.Group("/analytics")
	analytics.GET("/harvest", h.GetHarvestSummary)         // 収穫量集計取得
	analytics.GET("/charts/:type", h.GetChartData)         // グラフデータ取得（月別、作物別、区画別）
	analytics.GET("/export/:dataType", h.ExportCSV)        // CSVエクスポート（作物、収穫、タスク、全部）

	// Notification endpoints (protected)
	// 通知管理エンドポイント - デバイストークン登録、通知設定
	notifications := protected.Group("/notifications")
	notifications.POST("/device-token", h.RegisterDeviceToken)    // デバイストークン登録（FCM/APNS）
	notifications.DELETE("/device-token", h.DeleteDeviceToken)    // デバイストークン削除

	// User notification settings (protected)
	// ユーザー通知設定エンドポイント
	users.GET("/settings/notifications", h.GetNotificationSettings)    // 通知設定取得
	users.PUT("/settings/notifications", h.UpdateNotificationSettings) // 通知設定更新
}
