package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// =============================================================================
// Notification Sender - 通知送信サービス
// =============================================================================
// AWS SNS（プッシュ通知）とSES（メール通知）を使用して通知を配信します。
// Exponential backoffリトライ機構を備えています。

// Config は通知サービスの設定を保持します
type Config struct {
	AWSRegion        string // AWSリージョン
	SNSPlatformARN   string // SNS Platform Application ARN
	SESFromEmail     string // SES送信元メールアドレス
	MaxRetries       int    // 最大リトライ回数（デフォルト: 3）
	InitialBackoffMs int    // 初回リトライ待機時間（デフォルト: 1000ms）
}

// Sender は通知送信サービスです
type Sender struct {
	snsClient *sns.Client
	sesClient *ses.Client
	config    Config
}

// PushMessage はプッシュ通知のメッセージ構造です
type PushMessage struct {
	Title    string                 `json:"title"`
	Body     string                 `json:"body"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Priority string                 `json:"priority,omitempty"` // high または normal
}

// EmailMessage はメール通知のメッセージ構造です
type EmailMessage struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	BodyHTML    string `json:"body_html,omitempty"`
	BodyText    string `json:"body_text"`
	ReplyTo     string `json:"reply_to,omitempty"`
}

// SendResult は送信結果を表します
type SendResult struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NewSender は新しい通知送信サービスを作成します
func NewSender(ctx context.Context, cfg Config) (*Sender, error) {
	// デフォルト値設定
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.InitialBackoffMs <= 0 {
		cfg.InitialBackoffMs = 1000
	}

	// AWS SDK設定を読み込み
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Sender{
		snsClient: sns.NewFromConfig(awsCfg),
		sesClient: ses.NewFromConfig(awsCfg),
		config:    cfg,
	}, nil
}

// =============================================================================
// Push Notification Methods - プッシュ通知メソッド
// =============================================================================

// SendPush はプッシュ通知を送信します
// FCM/APNS向けにdata-onlyメッセージとして送信します
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - endpointARN: SNS Endpoint ARN（デバイストークンから生成）
//   - message: 送信するメッセージ
//
// 戻り値:
//   - *SendResult: 送信結果
func (s *Sender) SendPush(ctx context.Context, endpointARN string, message PushMessage) *SendResult {
	// デフォルト優先度設定
	if message.Priority == "" {
		message.Priority = "high"
	}

	// SNSメッセージフォーマット作成
	snsMessage, err := s.buildSNSMessage(message)
	if err != nil {
		return &SendResult{Success: false, Error: err.Error()}
	}

	// リトライ付きで送信
	var lastErr error
	for attempt := 0; attempt < s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(s.config.InitialBackoffMs*(1<<(attempt-1))) * time.Millisecond
			time.Sleep(backoff)
		}

		result, err := s.snsClient.Publish(ctx, &sns.PublishInput{
			TargetArn:        aws.String(endpointARN),
			Message:          aws.String(snsMessage),
			MessageStructure: aws.String("json"),
		})
		if err == nil {
			return &SendResult{
				Success:   true,
				MessageID: *result.MessageId,
			}
		}
		lastErr = err
	}

	return &SendResult{
		Success: false,
		Error:   fmt.Sprintf("failed after %d retries: %v", s.config.MaxRetries, lastErr),
	}
}

// buildSNSMessage はSNS用のJSONメッセージを構築します
// FCMとAPNS両方に対応したフォーマットを生成します
func (s *Sender) buildSNSMessage(msg PushMessage) (string, error) {
	// FCM (Android) 用メッセージ
	fcmMessage := map[string]interface{}{
		"data": map[string]interface{}{
			"title": msg.Title,
			"body":  msg.Body,
		},
		"priority": msg.Priority,
	}
	if msg.Data != nil {
		for k, v := range msg.Data {
			fcmMessage["data"].(map[string]interface{})[k] = v
		}
	}

	// APNS (iOS) 用メッセージ
	apnsMessage := map[string]interface{}{
		"aps": map[string]interface{}{
			"content-available": 1,
			"alert": map[string]string{
				"title": msg.Title,
				"body":  msg.Body,
			},
			"sound": "default",
		},
	}
	if msg.Data != nil {
		for k, v := range msg.Data {
			apnsMessage[k] = v
		}
	}

	// SNS用のマルチプラットフォームメッセージ
	fcmJSON, err := json.Marshal(fcmMessage)
	if err != nil {
		return "", err
	}
	apnsJSON, err := json.Marshal(apnsMessage)
	if err != nil {
		return "", err
	}

	snsMessage := map[string]string{
		"default": msg.Body,
		"GCM":     string(fcmJSON),
		"APNS":    string(apnsJSON),
	}

	result, err := json.Marshal(snsMessage)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// CreateEndpoint はデバイストークンからSNS Endpointを作成します
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - deviceToken: FCM/APNSトークン
//   - platform: プラットフォーム（ios, android）
//
// 戻り値:
//   - string: 作成されたEndpoint ARN
//   - error: 作成に失敗した場合のエラー
func (s *Sender) CreateEndpoint(ctx context.Context, deviceToken, platform string) (string, error) {
	result, err := s.snsClient.CreatePlatformEndpoint(ctx, &sns.CreatePlatformEndpointInput{
		PlatformApplicationArn: aws.String(s.config.SNSPlatformARN),
		Token:                  aws.String(deviceToken),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create platform endpoint: %w", err)
	}
	return *result.EndpointArn, nil
}

// DeleteEndpoint はSNS Endpointを削除します
// 無効なデバイストークン検出時に使用します
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - endpointARN: 削除するEndpoint ARN
//
// 戻り値:
//   - error: 削除に失敗した場合のエラー
func (s *Sender) DeleteEndpoint(ctx context.Context, endpointARN string) error {
	_, err := s.snsClient.DeleteEndpoint(ctx, &sns.DeleteEndpointInput{
		EndpointArn: aws.String(endpointARN),
	})
	return err
}

// =============================================================================
// Email Notification Methods - メール通知メソッド
// =============================================================================

// SendEmail はメール通知を送信します
// AWS SESを使用してトランザクションメールを送信します
//
// 引数:
//   - ctx: リクエストコンテキスト
//   - message: 送信するメッセージ
//
// 戻り値:
//   - *SendResult: 送信結果
func (s *Sender) SendEmail(ctx context.Context, message EmailMessage) *SendResult {
	// リトライ付きで送信
	var lastErr error
	for attempt := 0; attempt < s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(s.config.InitialBackoffMs*(1<<(attempt-1))) * time.Millisecond
			time.Sleep(backoff)
		}

		input := s.buildSESInput(message)
		result, err := s.sesClient.SendEmail(ctx, input)
		if err == nil {
			return &SendResult{
				Success:   true,
				MessageID: *result.MessageId,
			}
		}
		lastErr = err
	}

	return &SendResult{
		Success: false,
		Error:   fmt.Sprintf("failed after %d retries: %v", s.config.MaxRetries, lastErr),
	}
}

// buildSESInput はSES送信用の入力を構築します
func (s *Sender) buildSESInput(msg EmailMessage) *ses.SendEmailInput {
	input := &ses.SendEmailInput{
		Source: aws.String(s.config.SESFromEmail),
		Destination: &types.Destination{
			ToAddresses: []string{msg.To},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data:    aws.String(msg.Subject),
				Charset: aws.String("UTF-8"),
			},
			Body: &types.Body{},
		},
	}

	// テキスト本文
	if msg.BodyText != "" {
		input.Message.Body.Text = &types.Content{
			Data:    aws.String(msg.BodyText),
			Charset: aws.String("UTF-8"),
		}
	}

	// HTML本文
	if msg.BodyHTML != "" {
		input.Message.Body.Html = &types.Content{
			Data:    aws.String(msg.BodyHTML),
			Charset: aws.String("UTF-8"),
		}
	}

	// Reply-To設定
	if msg.ReplyTo != "" {
		input.ReplyToAddresses = []string{msg.ReplyTo}
	}

	return input
}

// =============================================================================
// Email Templates - メールテンプレート
// =============================================================================

// TaskReminderEmailTemplate はタスクリマインダーメールのテンプレートです
func TaskReminderEmailTemplate(taskTitle string, dueDate string, taskCount int) EmailMessage {
	subject := "【家庭菜園】今日のタスクリマインダー"
	bodyText := fmt.Sprintf("本日のタスクがあります。\n\nタスク: %s\n期限: %s\n\n合計 %d 件のタスクが予定されています。",
		taskTitle, dueDate, taskCount)
	bodyHTML := fmt.Sprintf(`
		<h2>今日のタスクリマインダー</h2>
		<p>本日のタスクがあります。</p>
		<ul>
			<li><strong>タスク:</strong> %s</li>
			<li><strong>期限:</strong> %s</li>
		</ul>
		<p>合計 <strong>%d 件</strong>のタスクが予定されています。</p>
		<p><a href="#">アプリで確認する</a></p>
	`, taskTitle, dueDate, taskCount)

	return EmailMessage{
		Subject:  subject,
		BodyText: bodyText,
		BodyHTML: bodyHTML,
	}
}

// OverdueAlertEmailTemplate は期限切れ警告メールのテンプレートです
func OverdueAlertEmailTemplate(overdueCount int) EmailMessage {
	subject := "【家庭菜園】期限切れタスクの警告"
	bodyText := fmt.Sprintf("%d 件のタスクが期限切れです。早めに対応してください。", overdueCount)
	bodyHTML := fmt.Sprintf(`
		<h2>⚠️ 期限切れタスクの警告</h2>
		<p><strong>%d 件</strong>のタスクが期限切れです。</p>
		<p>早めに対応してください。</p>
		<p><a href="#">アプリで確認する</a></p>
	`, overdueCount)

	return EmailMessage{
		Subject:  subject,
		BodyText: bodyText,
		BodyHTML: bodyHTML,
	}
}

// HarvestReminderEmailTemplate は収穫リマインダーメールのテンプレートです
func HarvestReminderEmailTemplate(cropName string, daysUntilHarvest int) EmailMessage {
	subject := "【家庭菜園】収穫リマインダー"
	bodyText := fmt.Sprintf("%s があと %d 日で収穫予定です。準備をお忘れなく！", cropName, daysUntilHarvest)
	bodyHTML := fmt.Sprintf(`
		<h2>🌱 収穫リマインダー</h2>
		<p><strong>%s</strong> があと <strong>%d 日</strong>で収穫予定です。</p>
		<p>準備をお忘れなく！</p>
		<p><a href="#">アプリで確認する</a></p>
	`, cropName, daysUntilHarvest)

	return EmailMessage{
		Subject:  subject,
		BodyText: bodyText,
		BodyHTML: bodyHTML,
	}
}
