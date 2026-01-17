package service

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"

	"github.com/jonosize/affiliate-platform/internal/logger"
	"github.com/jonosize/affiliate-platform/internal/model"
)

// ClickService handles click tracking business logic
type ClickService struct {
	clickRepo ClickRepositoryInterface
	linkRepo  LinkRepositoryInterface
	logger    logger.Logger
}

// NewClickService creates a new click service
func NewClickService(clickRepo ClickRepositoryInterface, linkRepo LinkRepositoryInterface, log logger.Logger) *ClickService {
	return &ClickService{
		clickRepo: clickRepo,
		linkRepo:  linkRepo,
		logger:    log,
	}
}

// TrackClick records a click event
func (s *ClickService) TrackClick(ctx context.Context, linkID uuid.UUID, ipAddress net.IP, userAgent, referrer string) error {
	ipStr := ""
	if ipAddress != nil {
		ipStr = ipAddress.String()
	}

	click := &model.Click{
		LinkID:    linkID,
		Timestamp: time.Now(),
		IPAddress: ipStr,
		UserAgent: userAgent,
		Referrer:  referrer,
	}

	if err := s.clickRepo.Create(ctx, click); err != nil {
		return fmt.Errorf("failed to track click: %w", err)
	}

	return nil
}

// GetClickStats returns click statistics for a link
func (s *ClickService) GetClickStats(ctx context.Context, linkID uuid.UUID) (int64, error) {
	return s.clickRepo.CountByLinkID(ctx, linkID)
}
