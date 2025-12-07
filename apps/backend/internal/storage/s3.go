// Package storage - S3ストレージサービス
//
// AWS S3を使用した画像アップロード機能を提供します。
// 機能:
//   - Presigned URLの生成（アップロード用、ダウンロード用）
//   - 画像バリデーション（サイズ、形式）
//   - Exponential backoffリトライ
//   - CloudFront CDN統合
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// =============================================================================
// 定数定義
// =============================================================================

const (
	// MaxImageSize は画像の最大サイズ（5MB）
	MaxImageSize = 5 * 1024 * 1024

	// PresignedURLExpiry はPresigned URLの有効期限（15分）
	PresignedURLExpiry = 15 * time.Minute

	// MaxRetryAttempts はリトライの最大回数
	MaxRetryAttempts = 3

	// InitialRetryDelay は最初のリトライ待機時間
	InitialRetryDelay = 1 * time.Second
)

// AllowedImageTypes は許可される画像形式のMIMEタイプ
var AllowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// =============================================================================
// エラー定義
// =============================================================================

var (
	// ErrFileTooLarge はファイルサイズが上限を超えている場合のエラー
	ErrFileTooLarge = errors.New("file size exceeds maximum allowed size (5MB)")

	// ErrInvalidImageType は画像形式が許可されていない場合のエラー
	ErrInvalidImageType = errors.New("invalid image type: only JPEG, PNG, and WEBP are allowed")

	// ErrUploadFailed はアップロードが失敗した場合のエラー
	ErrUploadFailed = errors.New("failed to upload image after retries")

	// ErrS3NotConfigured はS3が設定されていない場合のエラー
	ErrS3NotConfigured = errors.New("S3 storage is not configured")
)

// =============================================================================
// 設定構造体
// =============================================================================

// S3Config はS3接続設定を保持します
type S3Config struct {
	Region          string // AWSリージョン
	BucketName      string // S3バケット名
	AccessKeyID     string // AWSアクセスキーID
	SecretAccessKey string // AWSシークレットアクセスキー
	CloudFrontURL   string // CloudFront DistributionのURL（オプション）
	Endpoint        string // カスタムエンドポイント（LocalStack等用、オプション）
}

// IsConfigured はS3が設定されているかチェックします
func (c *S3Config) IsConfigured() bool {
	return c.BucketName != "" && c.Region != ""
}

// =============================================================================
// サービス構造体
// =============================================================================

// S3Service はS3ストレージ操作を提供するサービスです
type S3Service struct {
	client          *s3.Client
	presignClient   *s3.PresignClient
	config          *S3Config
}

// NewS3Service は新しいS3Serviceインスタンスを作成します
//
// 引数:
//   - cfg: S3設定
//
// 戻り値:
//   - *S3Service: S3サービスインスタンス
//   - error: 初期化に失敗した場合のエラー
func NewS3Service(cfg *S3Config) (*S3Service, error) {
	if cfg == nil || !cfg.IsConfigured() {
		return &S3Service{config: cfg}, nil // 設定なしでも動作（ダミーモード）
	}

	// AWS設定を構築
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// S3クライアントを作成
	var client *s3.Client
	if cfg.Endpoint != "" {
		// カスタムエンドポイント（LocalStack等）
		client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		})
	} else {
		client = s3.NewFromConfig(awsCfg)
	}

	// Presignクライアントを作成
	presignClient := s3.NewPresignClient(client)

	return &S3Service{
		client:        client,
		presignClient: presignClient,
		config:        cfg,
	}, nil
}

// =============================================================================
// Presigned URL生成
// =============================================================================

// PresignedUploadResult はPresigned URL生成結果を表します
type PresignedUploadResult struct {
	UploadURL   string    `json:"upload_url"`    // アップロード用Presigned URL
	ObjectKey   string    `json:"object_key"`    // S3オブジェクトキー
	ContentURL  string    `json:"content_url"`   // アップロード後の画像URL（CloudFront経由）
	ExpiresAt   time.Time `json:"expires_at"`    // URLの有効期限
}

// GenerateUploadURL はアップロード用のPresigned URLを生成します
//
// 引数:
//   - ctx: コンテキスト
//   - userID: ユーザーID（パス構成用）
//   - contentType: MIMEタイプ（image/jpeg等）
//
// 戻り値:
//   - *PresignedUploadResult: Presigned URL情報
//   - error: 生成に失敗した場合のエラー
func (s *S3Service) GenerateUploadURL(ctx context.Context, userID uint, contentType string) (*PresignedUploadResult, error) {
	// S3が設定されているかチェック
	if s.client == nil || s.config == nil || !s.config.IsConfigured() {
		return nil, ErrS3NotConfigured
	}

	// 画像形式をバリデーション
	ext, ok := AllowedImageTypes[contentType]
	if !ok {
		return nil, ErrInvalidImageType
	}

	// ユニークなオブジェクトキーを生成
	// パス形式: crops/images/{userID}/{year}/{month}/{uuid}.{ext}
	now := time.Now()
	objectKey := fmt.Sprintf("crops/images/%d/%d/%02d/%s%s",
		userID,
		now.Year(),
		now.Month(),
		uuid.New().String(),
		ext,
	)

	// Presigned URLを生成
	expiresAt := now.Add(PresignedURLExpiry)
	presignedReq, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(PresignedURLExpiry))
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// コンテンツURLを構築（CloudFront経由またはS3直接）
	var contentURL string
	if s.config.CloudFrontURL != "" {
		contentURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(s.config.CloudFrontURL, "/"), objectKey)
	} else {
		contentURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
			s.config.BucketName,
			s.config.Region,
			objectKey,
		)
	}

	return &PresignedUploadResult{
		UploadURL:  presignedReq.URL,
		ObjectKey:  objectKey,
		ContentURL: contentURL,
		ExpiresAt:  expiresAt,
	}, nil
}

// =============================================================================
// 画像アップロード（サーバーサイド）
// =============================================================================

// UploadResult はアップロード結果を表します
type UploadResult struct {
	ObjectKey  string `json:"object_key"`   // S3オブジェクトキー
	ContentURL string `json:"content_url"`  // 画像URL
	Size       int64  `json:"size"`         // ファイルサイズ（バイト）
}

// UploadImage はサーバーサイドで画像をS3にアップロードします
// Exponential backoffリトライを適用します
//
// 引数:
//   - ctx: コンテキスト
//   - userID: ユーザーID（パス構成用）
//   - reader: 画像データのReader
//   - contentType: MIMEタイプ
//   - size: ファイルサイズ
//
// 戻り値:
//   - *UploadResult: アップロード結果
//   - error: アップロードに失敗した場合のエラー
func (s *S3Service) UploadImage(ctx context.Context, userID uint, reader io.Reader, contentType string, size int64) (*UploadResult, error) {
	// S3が設定されているかチェック
	if s.client == nil || s.config == nil || !s.config.IsConfigured() {
		return nil, ErrS3NotConfigured
	}

	// サイズをバリデーション
	if size > MaxImageSize {
		return nil, ErrFileTooLarge
	}

	// 画像形式をバリデーション
	ext, ok := AllowedImageTypes[contentType]
	if !ok {
		return nil, ErrInvalidImageType
	}

	// オブジェクトキーを生成
	now := time.Now()
	objectKey := fmt.Sprintf("crops/images/%d/%d/%02d/%s%s",
		userID,
		now.Year(),
		now.Month(),
		uuid.New().String(),
		ext,
	)

	// Exponential backoffリトライでアップロード
	var lastErr error
	for attempt := 0; attempt < MaxRetryAttempts; attempt++ {
		if attempt > 0 {
			// 待機時間を計算（初回1秒、2回目2秒、3回目4秒）
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * InitialRetryDelay
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:        aws.String(s.config.BucketName),
			Key:           aws.String(objectKey),
			Body:          reader,
			ContentType:   aws.String(contentType),
			ContentLength: aws.Int64(size),
		})
		if err == nil {
			// 成功
			var contentURL string
			if s.config.CloudFrontURL != "" {
				contentURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(s.config.CloudFrontURL, "/"), objectKey)
			} else {
				contentURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
					s.config.BucketName,
					s.config.Region,
					objectKey,
				)
			}

			return &UploadResult{
				ObjectKey:  objectKey,
				ContentURL: contentURL,
				Size:       size,
			}, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("%w: %v", ErrUploadFailed, lastErr)
}

// =============================================================================
// バリデーションヘルパー
// =============================================================================

// ValidateImageFile は画像ファイルをバリデーションします
//
// 引数:
//   - fileHeader: アップロードされたファイルのヘッダー
//   - size: ファイルサイズ
//
// 戻り値:
//   - string: 検出されたMIMEタイプ
//   - error: バリデーションエラー
func ValidateImageFile(data []byte, size int64) (string, error) {
	// サイズチェック
	if size > MaxImageSize {
		return "", ErrFileTooLarge
	}

	// MIMEタイプを検出
	contentType := http.DetectContentType(data)

	// 許可されたタイプかチェック
	if _, ok := AllowedImageTypes[contentType]; !ok {
		return "", ErrInvalidImageType
	}

	return contentType, nil
}

// GetExtensionFromContentType はMIMEタイプから拡張子を取得します
func GetExtensionFromContentType(contentType string) string {
	if ext, ok := AllowedImageTypes[contentType]; ok {
		return ext
	}
	return ""
}

// IsAllowedExtension はファイル拡張子が許可されているかチェックします
func IsAllowedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowedExt := range AllowedImageTypes {
		if ext == allowedExt {
			return true
		}
	}
	return false
}
