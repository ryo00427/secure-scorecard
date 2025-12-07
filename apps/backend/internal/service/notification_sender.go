package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/secure-scorecard/backend/internal/config"
	"github.com/secure-scorecard/backend/internal/model"
)

// =============================================================================
// Notification Sender - 通知送信サービス
// =============================================================================
// AWS SNS（プッシュ通知）とAWS SES（メール通知）を使用して通知を送信します。
// Exponential backoffによるリトライ機構を実装しています。

// NotificationSender は通知送信インターフェースです。
type NotificationSender interface {
	// SendPushNotification はプッシュ通知を送信します。
	SendPushNotification(ctx context.Context, token *model.DeviceToken, title, body string, data map[string]interface{}) error

	// SendEmailNotification はメール通知を送信します。
	SendEmailNotification(ctx context.Context, toEmail, subject, htmlBody, textBody string) error

	// SendNotificationEvent は通知イベントを処理して送信します。
	SendNotificationEvent(ctx context.Context, event NotificationEvent, user *model.User, tokens []model.DeviceToken) error
}

// notificationSender はNotificationSenderの実装です。
type notificationSender struct {
	snsClient *sns.Client
	sesClient *ses.Client
	cfg       *config.NotificationConfig
}

// NewNotificationSender は新しいNotificationSenderを作成します。
//
// 引数:
//   - cfg: 通知設定（AWS設定を含む）
//
// 戻り値:
//   - NotificationSender: 通知送信インターフェース
//   - error: 初期化に失敗した場合のエラー
func NewNotificationSender(cfg *config.NotificationConfig) (NotificationSender, error) {
	// AWS設定をロード
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &notificationSender{
		snsClient: sns.NewFromConfig(awsCfg),
		sesClient: ses.NewFromConfig(awsCfg),
		cfg:       cfg,
	}, nil
}

// =============================================================================
// Push Notification - プッシュ通知
// =============================================================================

// PushMessage はプッシュ通知のメッセージ構造体です。
// FCM/APNSの両方に対応したフォーマットを定義します。
type PushMessage struct {
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

// FCMMessage はFirebase Cloud Messaging向けのメッセージ構造体です。
type FCMMessage struct {
	Notification *FCMNotification       `json:"notification,omitempty"`
	Data         map[string]string      `json:"data,omitempty"`
	Priority     string                 `json:"priority,omitempty"`
}

// FCMNotification はFCM通知部分の構造体です。
type FCMNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// APNSMessage はApple Push Notification Service向けのメッセージ構造体です。
type APNSMessage struct {
	APS  APNSPayload            `json:"aps"`
	Data map[string]interface{} `json:"data,omitempty"`
}

// APNSPayload はAPNS通知ペイロードです。
type APNSPayload struct {
	Alert            APNSAlert `json:"alert"`
	ContentAvailable int       `json:"content-available,omitempty"`
	MutableContent   int       `json:"mutable-content,omitempty"`
}

// APNSAlert はAPNSアラート部分の構造体です。
type APNSAlert struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// SendPushNotification はプッシュ通知を送信します。
// プラットフォームに応じてFCMまたはAPNS形式でメッセージを構築します。
//
// 引数:
//   - ctx: コンテキスト
//   - token: デバイストークン
//   - title: 通知タイトル
//   - body: 通知本文
//   - data: カスタムデータ（任意）
//
// 戻り値:
//   - error: 送信に失敗した場合のエラー
func (n *notificationSender) SendPushNotification(ctx context.Context, token *model.DeviceToken, title, body string, data map[string]interface{}) error {
	// プラットフォームに応じたARNを取得
	var platformARN string
	switch token.Platform {
	case "ios":
		platformARN = n.cfg.SNSPlatformARNiOS
	case "android", "web":
		platformARN = n.cfg.SNSPlatformARNAndroid
	default:
		return fmt.Errorf("unsupported platform: %s", token.Platform)
	}

	if platformARN == "" {
		return fmt.Errorf("platform ARN not configured for %s", token.Platform)
	}

	// エンドポイントを作成または取得
	endpointARN, err := n.getOrCreateEndpoint(ctx, platformARN, token.Token)
	if err != nil {
		return fmt.Errorf("failed to get/create endpoint: %w", err)
	}

	// メッセージを構築
	message, err := n.buildPushMessage(token.Platform, title, body, data)
	if err != nil {
		return fmt.Errorf("failed to build message: %w", err)
	}

	// リトライ付きで送信
	return n.sendWithRetry(ctx, func() error {
		_, err := n.snsClient.Publish(ctx, &sns.PublishInput{
			TargetArn:        aws.String(endpointARN),
			Message:          aws.String(message),
			MessageStructure: aws.String("json"),
		})
		return err
	})
}

// getOrCreateEndpoint はSNSエンドポイントを取得または作成します。
func (n *notificationSender) getOrCreateEndpoint(ctx context.Context, platformARN, token string) (string, error) {
	result, err := n.snsClient.CreatePlatformEndpoint(ctx, &sns.CreatePlatformEndpointInput{
		PlatformApplicationArn: aws.String(platformARN),
		Token:                  aws.String(token),
	})
	if err != nil {
		return "", err
	}
	return *result.EndpointArn, nil
}

// buildPushMessage はプラットフォームに応じたメッセージを構築します。
func (n *notificationSender) buildPushMessage(platform, title, body string, data map[string]interface{}) (string, error) {
	// SNSはプラットフォームごとに異なるフォーマットを期待する
	messageMap := make(map[string]string)

	switch platform {
	case "ios":
		apnsMessage := APNSMessage{
			APS: APNSPayload{
				Alert: APNSAlert{
					Title: title,
					Body:  body,
				},
				ContentAvailable: 1,
				MutableContent:   1,
			},
			Data: data,
		}
		apnsJSON, err := json.Marshal(apnsMessage)
		if err != nil {
			return "", err
		}
		messageMap["APNS"] = string(apnsJSON)
		messageMap["APNS_SANDBOX"] = string(apnsJSON)

	case "android", "web":
		// dataをstring mapに変換
		stringData := make(map[string]string)
		for k, v := range data {
			stringData[k] = fmt.Sprintf("%v", v)
		}

		fcmMessage := FCMMessage{
			Notification: &FCMNotification{
				Title: title,
				Body:  body,
			},
			Data:     stringData,
			Priority: "high",
		}
		fcmJSON, err := json.Marshal(fcmMessage)
		if err != nil {
			return "", err
		}
		messageMap["GCM"] = string(fcmJSON)
	}

	// デフォルトメッセージ
	defaultMessage := fmt.Sprintf("%s: %s", title, body)
	messageMap["default"] = defaultMessage

	result, err := json.Marshal(messageMap)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// =============================================================================
// Email Notification - メール通知
// =============================================================================

// SendEmailNotification はメール通知を送信します。
//
// 引数:
//   - ctx: コンテキスト
//   - toEmail: 送信先メールアドレス
//   - subject: 件名
//   - htmlBody: HTML形式の本文
//   - textBody: テキスト形式の本文
//
// 戻り値:
//   - error: 送信に失敗した場合のエラー
func (n *notificationSender) SendEmailNotification(ctx context.Context, toEmail, subject, htmlBody, textBody string) error {
	if n.cfg.SESFromEmail == "" {
		return fmt.Errorf("SES from email not configured")
	}

	fromAddress := n.cfg.SESFromEmail
	if n.cfg.SESFromName != "" {
		fromAddress = fmt.Sprintf("%s <%s>", n.cfg.SESFromName, n.cfg.SESFromEmail)
	}

	return n.sendWithRetry(ctx, func() error {
		_, err := n.sesClient.SendEmail(ctx, &ses.SendEmailInput{
			Source: aws.String(fromAddress),
			Destination: &sestypes.Destination{
				ToAddresses: []string{toEmail},
			},
			Message: &sestypes.Message{
				Subject: &sestypes.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(subject),
				},
				Body: &sestypes.Body{
					Html: &sestypes.Content{
						Charset: aws.String("UTF-8"),
						Data:    aws.String(htmlBody),
					},
					Text: &sestypes.Content{
						Charset: aws.String("UTF-8"),
						Data:    aws.String(textBody),
					},
				},
			},
		})
		return err
	})
}

// =============================================================================
// Notification Event Handler - 通知イベント処理
// =============================================================================

// SendNotificationEvent は通知イベントを処理して送信します。
// ユーザーの通知設定に基づいて、プッシュ通知とメール通知を送信します。
//
// 引数:
//   - ctx: コンテキスト
//   - event: 通知イベント
//   - user: 対象ユーザー
//   - tokens: ユーザーのデバイストークン
//
// 戻り値:
//   - error: 送信に失敗した場合のエラー
func (n *notificationSender) SendNotificationEvent(ctx context.Context, event NotificationEvent, user *model.User, tokens []model.DeviceToken) error {
	settings := user.NotificationSettings
	if settings == nil {
		// デフォルト設定
		settings = &model.NotificationSettings{
			PushEnabled:   true,
			EmailEnabled:  true,
			TaskReminders: true,
			HarvestReminders: true,
		}
	}

	// イベントタイプに応じた通知設定チェック
	var shouldSend bool
	switch event.Type {
	case NotificationEventTaskDueReminder, NotificationEventTaskOverdueAlert:
		shouldSend = settings.TaskReminders
	case NotificationEventHarvestReminder:
		shouldSend = settings.HarvestReminders
	default:
		shouldSend = true
	}

	if !shouldSend {
		return nil // 通知設定で無効化されている
	}

	var lastErr error

	// プッシュ通知を送信
	if settings.PushEnabled && len(tokens) > 0 {
		for _, token := range tokens {
			if token.IsActive {
				if err := n.SendPushNotification(ctx, &token, event.Title, event.Body, event.Data); err != nil {
					lastErr = err
					// エラーでも他のトークンへの送信を継続
				}
			}
		}
	}

	// メール通知を送信
	if settings.EmailEnabled && user.Email != "" {
		htmlBody := n.buildEmailHTML(event)
		textBody := fmt.Sprintf("%s\n\n%s", event.Title, event.Body)

		if err := n.SendEmailNotification(ctx, user.Email, event.Title, htmlBody, textBody); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// buildEmailHTML はメール通知用のHTML本文を生成します。
func (n *notificationSender) buildEmailHTML(event NotificationEvent) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #16a34a; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background-color: #f9fafb; padding: 20px; border-radius: 0 0 8px 8px; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s</h1>
        </div>
        <div class="content">
            <p>%s</p>
        </div>
        <div class="footer">
            <p>Home Garden アプリからの通知</p>
        </div>
    </div>
</body>
</html>
`, event.Title, event.Body)
}

// =============================================================================
// Retry Logic - リトライ機構
// =============================================================================

// sendWithRetry はExponential backoffでリトライを行います。
//
// リトライ条件:
//   - 最大リトライ回数: MaxRetries（デフォルト3回）
//   - 初回待機時間: InitialBackoffMs（デフォルト1000ms）
//   - 待機時間は毎回2倍に増加
func (n *notificationSender) sendWithRetry(ctx context.Context, fn func() error) error {
	maxRetries := n.cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	backoffMs := n.cfg.InitialBackoffMs
	if backoffMs <= 0 {
		backoffMs = 1000
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := fn(); err != nil {
			lastErr = err

			// 最後のリトライの場合はリトライしない
			if attempt == maxRetries {
				break
			}

			// 指数バックオフで待機
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(backoffMs) * time.Millisecond):
			}

			// 次回の待機時間を2倍に
			backoffMs *= 2
		} else {
			return nil // 成功
		}
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// =============================================================================
// Mock Implementation - テスト用モック
// =============================================================================

// MockNotificationSender はテスト用のモック実装です。
type MockNotificationSender struct {
	SentPushNotifications  []PushNotificationRecord
	SentEmailNotifications []EmailNotificationRecord
	ShouldFail             bool
}

// PushNotificationRecord はプッシュ通知の送信記録です。
type PushNotificationRecord struct {
	Token string
	Title string
	Body  string
	Data  map[string]interface{}
}

// EmailNotificationRecord はメール通知の送信記録です。
type EmailNotificationRecord struct {
	ToEmail  string
	Subject  string
	HTMLBody string
	TextBody string
}

// NewMockNotificationSender は新しいモック通知送信者を作成します。
func NewMockNotificationSender() *MockNotificationSender {
	return &MockNotificationSender{
		SentPushNotifications:  make([]PushNotificationRecord, 0),
		SentEmailNotifications: make([]EmailNotificationRecord, 0),
	}
}

// SendPushNotification はプッシュ通知をモックで記録します。
func (m *MockNotificationSender) SendPushNotification(ctx context.Context, token *model.DeviceToken, title, body string, data map[string]interface{}) error {
	if m.ShouldFail {
		return fmt.Errorf("mock error: push notification failed")
	}
	m.SentPushNotifications = append(m.SentPushNotifications, PushNotificationRecord{
		Token: token.Token,
		Title: title,
		Body:  body,
		Data:  data,
	})
	return nil
}

// SendEmailNotification はメール通知をモックで記録します。
func (m *MockNotificationSender) SendEmailNotification(ctx context.Context, toEmail, subject, htmlBody, textBody string) error {
	if m.ShouldFail {
		return fmt.Errorf("mock error: email notification failed")
	}
	m.SentEmailNotifications = append(m.SentEmailNotifications, EmailNotificationRecord{
		ToEmail:  toEmail,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
	return nil
}

// SendNotificationEvent はイベントをモックで処理します。
func (m *MockNotificationSender) SendNotificationEvent(ctx context.Context, event NotificationEvent, user *model.User, tokens []model.DeviceToken) error {
	if m.ShouldFail {
		return fmt.Errorf("mock error: notification event failed")
	}

	// プッシュ通知を記録
	for _, token := range tokens {
		if token.IsActive {
			m.SentPushNotifications = append(m.SentPushNotifications, PushNotificationRecord{
				Token: token.Token,
				Title: event.Title,
				Body:  event.Body,
				Data:  event.Data,
			})
		}
	}

	// メール通知を記録
	if user.Email != "" {
		m.SentEmailNotifications = append(m.SentEmailNotifications, EmailNotificationRecord{
			ToEmail: user.Email,
			Subject: event.Title,
			TextBody: event.Body,
		})
	}

	return nil
}
