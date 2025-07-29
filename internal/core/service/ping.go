package service

import (
	"context"

	"leeta/internal/core/domain"
	"leeta/internal/core/port"
)

/**
 * PingService implements port.PingService interface
 */
type PingService struct {
	repo port.PingRepository
}

// NewAuthService creates a new auth service instance
func NewPingService(repo port.PingRepository) *PingService {
	return &PingService{
		repo,
	}
}

// Login gives a registered user an access token if the credentials are valid
func (ps *PingService) Ping(ctx context.Context, ping *domain.Ping) (domain.Ping, domain.CError) {
	_ = ps.repo.CreatePing(ctx, ping)
	return *ping, nil
}
