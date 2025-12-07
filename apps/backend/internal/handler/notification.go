package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/secure-scorecard/backend/internal/model"
)

// =============================================================================
// Notification Handler - 通知管理ハンドラー
// =============================================================================
// デバイストークン登録、通知設定更新などの通知関連エンドポイントを提供します。

// RegisterDeviceTokenRequest はデバイストークン登録リクエストの構造体です。
type RegisterDeviceTokenRequest struct {
	Token    string `json:"token" validate:"required"`             // FCM/APNSトークン
	Platform string `json:"platform" validate:"required,oneof=ios android web"` // ios, android, web
	DeviceID string `json:"device_id,omitempty"`                   // デバイス識別子（オプション）
}

// RegisterDeviceTokenResponse はデバイストークン登録レスポンスです。
type RegisterDeviceTokenResponse struct {
	ID       uint   `json:"id"`
	Platform string `json:"platform"`
	IsActive bool   `json:"is_active"`
	Message  string `json:"message"`
}

// UpdateNotificationSettingsRequest は通知設定更新リクエストの構造体です。
type UpdateNotificationSettingsRequest struct {
	PushEnabled               *bool `json:"push_enabled,omitempty"`
	EmailEnabled              *bool `json:"email_enabled,omitempty"`
	TaskReminders             *bool `json:"task_reminders,omitempty"`
	HarvestReminders          *bool `json:"harvest_reminders,omitempty"`
	GrowthRecordNotifications *bool `json:"growth_record_notifications,omitempty"`
}

// NotificationSettingsResponse は通知設定レスポンスです。
type NotificationSettingsResponse struct {
	PushEnabled               bool   `json:"push_enabled"`
	EmailEnabled              bool   `json:"email_enabled"`
	TaskReminders             bool   `json:"task_reminders"`
	HarvestReminders          bool   `json:"harvest_reminders"`
	GrowthRecordNotifications bool   `json:"growth_record_notifications"`
	Message                   string `json:"message,omitempty"`
}

// RegisterDeviceToken はデバイストークンを登録します。
// 同じユーザー・プラットフォームの既存トークンがある場合は更新します。
//
// エンドポイント: POST /api/v1/notifications/device-token
//
// リクエストボディ:
//
//	{
//	  "token": "fcm_or_apns_token_string",
//	  "platform": "ios",
//	  "device_id": "device-uuid" // optional
//	}
//
// レスポンス:
//
//	{
//	  "id": 1,
//	  "platform": "ios",
//	  "is_active": true,
//	  "message": "デバイストークンを登録しました"
//	}
func (h *Handler) RegisterDeviceToken(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザーIDを取得
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "認証が必要です",
		})
	}

	// リクエストをパース
	var req RegisterDeviceTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "invalid_request",
			"message": "リクエストの形式が正しくありません",
		})
	}

	// バリデーション
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "validation_error",
			"message": err.Error(),
		})
	}

	// サービス層でトークン登録/更新
	deviceToken, err := h.service.RegisterDeviceToken(ctx, userID, req.Token, req.Platform, req.DeviceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "registration_failed",
			"message": "デバイストークンの登録に失敗しました",
		})
	}

	return c.JSON(http.StatusOK, RegisterDeviceTokenResponse{
		ID:       deviceToken.ID,
		Platform: deviceToken.Platform,
		IsActive: deviceToken.IsActive,
		Message:  "デバイストークンを登録しました",
	})
}

// DeleteDeviceToken はデバイストークンを削除します。
//
// エンドポイント: DELETE /api/v1/notifications/device-token
//
// クエリパラメータ:
//   - platform: プラットフォーム（ios, android, web）
func (h *Handler) DeleteDeviceToken(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザーIDを取得
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "認証が必要です",
		})
	}

	platform := c.QueryParam("platform")
	if platform == "" {
		// プラットフォーム指定なしの場合は全削除
		if err := h.service.DeleteAllDeviceTokens(ctx, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "deletion_failed",
				"message": "デバイストークンの削除に失敗しました",
			})
		}
		return c.JSON(http.StatusOK, map[string]string{
			"message": "全てのデバイストークンを削除しました",
		})
	}

	// 特定プラットフォームのトークンを削除
	if err := h.service.DeleteDeviceTokenByPlatform(ctx, userID, platform); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "deletion_failed",
			"message": "デバイストークンの削除に失敗しました",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "デバイストークンを削除しました",
	})
}

// GetNotificationSettings は通知設定を取得します。
//
// エンドポイント: GET /api/v1/users/settings/notifications
func (h *Handler) GetNotificationSettings(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザーIDを取得
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "認証が必要です",
		})
	}

	// ユーザー情報を取得
	user, err := h.service.GetUserByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "fetch_failed",
			"message": "ユーザー情報の取得に失敗しました",
		})
	}

	// 通知設定を返す（デフォルト値を設定）
	settings := user.NotificationSettings
	if settings == nil {
		settings = &model.NotificationSettings{
			PushEnabled:               true,
			EmailEnabled:              true,
			TaskReminders:             true,
			HarvestReminders:          true,
			GrowthRecordNotifications: false,
		}
	}

	return c.JSON(http.StatusOK, NotificationSettingsResponse{
		PushEnabled:               settings.PushEnabled,
		EmailEnabled:              settings.EmailEnabled,
		TaskReminders:             settings.TaskReminders,
		HarvestReminders:          settings.HarvestReminders,
		GrowthRecordNotifications: settings.GrowthRecordNotifications,
	})
}

// UpdateNotificationSettings は通知設定を更新します。
//
// エンドポイント: PUT /api/v1/users/settings/notifications
//
// リクエストボディ:
//
//	{
//	  "push_enabled": true,
//	  "email_enabled": true,
//	  "task_reminders": true,
//	  "harvest_reminders": true,
//	  "growth_record_notifications": false
//	}
func (h *Handler) UpdateNotificationSettings(c echo.Context) error {
	ctx := c.Request().Context()

	// ユーザーIDを取得
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":   "unauthorized",
			"message": "認証が必要です",
		})
	}

	// リクエストをパース
	var req UpdateNotificationSettingsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "invalid_request",
			"message": "リクエストの形式が正しくありません",
		})
	}

	// サービス層で設定更新
	settings, err := h.service.UpdateNotificationSettings(ctx, userID, &model.NotificationSettings{
		PushEnabled:               getBoolValue(req.PushEnabled, true),
		EmailEnabled:              getBoolValue(req.EmailEnabled, true),
		TaskReminders:             getBoolValue(req.TaskReminders, true),
		HarvestReminders:          getBoolValue(req.HarvestReminders, true),
		GrowthRecordNotifications: getBoolValue(req.GrowthRecordNotifications, false),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "update_failed",
			"message": "通知設定の更新に失敗しました",
		})
	}

	return c.JSON(http.StatusOK, NotificationSettingsResponse{
		PushEnabled:               settings.PushEnabled,
		EmailEnabled:              settings.EmailEnabled,
		TaskReminders:             settings.TaskReminders,
		HarvestReminders:          settings.HarvestReminders,
		GrowthRecordNotifications: settings.GrowthRecordNotifications,
		Message:                   "通知設定を更新しました",
	})
}

// getBoolValue は *bool から bool を取得します（nilの場合はデフォルト値を返す）
func getBoolValue(ptr *bool, defaultValue bool) bool {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}
