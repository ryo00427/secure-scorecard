package service

import (
	"github.com/labstack/echo/v4"
)

// IsTokenBlacklisted implements auth.TokenBlacklistChecker interface
func (s *Service) IsTokenBlacklisted(c echo.Context, tokenHash string) (bool, error) {
	ctx := c.Request().Context()
	return s.repos.TokenBlacklist().IsBlacklisted(ctx, tokenHash)
}
