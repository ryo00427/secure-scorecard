package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	JWT          JWTConfig
	CORS         CORSConfig
	S3           S3Config
	Scheduler    SchedulerConfig
	Notification NotificationConfig
}

// NotificationConfig は通知サービスの設定を保持します
type NotificationConfig struct {
	// AWS共通設定
	AWSRegion string // AWSリージョン（SNS/SES用）

	// SNS設定（プッシュ通知用）
	SNSPlatformARNiOS     string // iOS用 SNS Platform Application ARN
	SNSPlatformARNAndroid string // Android用 SNS Platform Application ARN

	// SES設定（メール通知用）
	SESFromEmail string // SES送信元メールアドレス
	SESFromName  string // 送信者名

	// リトライ設定
	MaxRetries       int // 最大リトライ回数（デフォルト: 3）
	InitialBackoffMs int // 初回リトライ待機時間(ms)（デフォルト: 1000）
}

// SchedulerConfig はスケジューラー関連の設定を保持します
type SchedulerConfig struct {
	AuthToken string // EventBridge Scheduler からの認証トークン
}

// S3Config はS3/CloudFront設定を保持します
type S3Config struct {
	Region          string // AWSリージョン
	BucketName      string // S3バケット名
	AccessKeyID     string // AWSアクセスキーID
	SecretAccessKey string // AWSシークレットアクセスキー
	CloudFrontURL   string // CloudFront DistributionのURL（オプション）
	Endpoint        string // カスタムエンドポイント（LocalStack等用、オプション）
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port string
	Env  string
}

// DatabaseConfig holds database-specific configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret     string
	ExpireHour int
}

// CORSConfig holds CORS-specific configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "home_garden"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "dev-secret-change-in-production"),
			ExpireHour: getEnvAsInt("JWT_EXPIRE_HOUR", 24),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8081"}),
		},
		S3: S3Config{
			Region:          getEnv("AWS_REGION", "ap-northeast-1"),
			BucketName:      getEnv("S3_BUCKET_NAME", ""),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			CloudFrontURL:   getEnv("CLOUDFRONT_URL", ""),
			Endpoint:        getEnv("S3_ENDPOINT", ""), // LocalStack用
		},
		Scheduler: SchedulerConfig{
			AuthToken: getEnv("SCHEDULER_AUTH_TOKEN", ""), // EventBridge用認証トークン
		},
		Notification: NotificationConfig{
			AWSRegion:             getEnv("AWS_REGION", "ap-northeast-1"),
			SNSPlatformARNiOS:     getEnv("SNS_PLATFORM_ARN_IOS", ""),
			SNSPlatformARNAndroid: getEnv("SNS_PLATFORM_ARN_ANDROID", ""),
			SESFromEmail:          getEnv("SES_FROM_EMAIL", ""),
			SESFromName:           getEnv("SES_FROM_NAME", "Home Garden"),
			MaxRetries:            getEnvAsInt("NOTIFICATION_MAX_RETRIES", 3),
			InitialBackoffMs:      getEnvAsInt("NOTIFICATION_INITIAL_BACKOFF_MS", 1000),
		},
	}

	return config, nil
}

// DSN returns the PostgreSQL connection string
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsSlice gets an environment variable as comma-separated slice or returns a default value
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		result := []string{}
		for i := 0; i < len(value); {
			start := i
			for i < len(value) && value[i] != ',' {
				i++
			}
			item := value[start:i]
			// Trim spaces
			for len(item) > 0 && (item[0] == ' ' || item[0] == '\t') {
				item = item[1:]
			}
			for len(item) > 0 && (item[len(item)-1] == ' ' || item[len(item)-1] == '\t') {
				item = item[:len(item)-1]
			}
			if len(item) > 0 {
				result = append(result, item)
			}
			if i < len(value) {
				i++ // skip comma
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
