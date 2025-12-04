package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"github.com/secure-scorecard/backend/internal/model"
)

type tokenBlacklistRepository struct {
	db *gorm.DB
}

// NewTokenBlacklistRepository creates a new token blacklist repository
func NewTokenBlacklistRepository(db *gorm.DB) TokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

// Add adds a token hash to the blacklist
func (r *tokenBlacklistRepository) Add(ctx context.Context, tokenHash string, expiresAt time.Time) error {
	token := &model.TokenBlacklist{
		TokenHash: tokenHash,
		RevokedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
	return GetDB(ctx, r.db).WithContext(ctx).Create(token).Error
}

// IsBlacklisted checks if a token hash is blacklisted
func (r *tokenBlacklistRepository) IsBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	var count int64
	err := GetDB(ctx, r.db).WithContext(ctx).
		Model(&model.TokenBlacklist{}).
		Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteExpired deletes expired tokens from the blacklist
func (r *tokenBlacklistRepository) DeleteExpired(ctx context.Context) error {
	return GetDB(ctx, r.db).WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&model.TokenBlacklist{}).Error
}
