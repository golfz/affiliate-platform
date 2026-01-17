package service

import (
	"context"
	"fmt"
	"net"

	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/validator"
)

// RedirectService handles redirect business logic
type RedirectService struct {
	linkRepo LinkRepositoryInterface
	clickSvc *ClickService
	logger   logger.Logger
}

// NewRedirectService creates a new redirect service
func NewRedirectService(linkRepo LinkRepositoryInterface, clickSvc *ClickService, log logger.Logger) *RedirectService {
	return &RedirectService{
		linkRepo: linkRepo,
		clickSvc: clickSvc,
		logger:   log,
	}
}

// Redirect handles redirect logic: finds link, validates URL, tracks click
func (s *RedirectService) Redirect(ctx context.Context, shortCode string, ipAddress net.IP, userAgent, referrer string) (string, error) {
	// Find link by short code
	link, err := s.linkRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("link not found: %w", err)
	}

	// Validate redirect URL (whitelist domains)
	if !validator.ValidateRedirectURL(link.TargetURL) {
		s.logger.Error("Invalid redirect URL", logger.String("url", link.TargetURL), logger.String("short_code", shortCode))
		return "", fmt.Errorf("invalid redirect URL")
	}

	// Track click (async - don't block redirect)
	go func() {
		if err := s.clickSvc.TrackClick(context.Background(), link.ID, ipAddress, userAgent, referrer); err != nil {
			s.logger.Error("Failed to track click", logger.String("error", err.Error()), logger.String("link_id", link.ID.String()))
		}
	}()

	return link.TargetURL, nil
}
