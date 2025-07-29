package service

import (
	"context"

	"leeta/internal/adapter/logger"
	"leeta/internal/core/domain"
	"leeta/internal/core/port"

	"go.uber.org/zap"
)

/**
 * LocationService implements port.LocationService interface
 */
type LocationService struct {
	repo port.LocationRepository
}

// NewLocationService creates a new location service instance
func NewLocationService(repo port.LocationRepository) *LocationService {
	return &LocationService{
		repo,
	}
}

func (ls *LocationService) RegisterLocation(ctx context.Context, location *domain.RegisterLocationRequest) (*domain.Location, domain.CError) {
	locationToCreate := domain.Location{
		Name:      location.Name,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
	}

	locationResponse, cerr := ls.repo.CreateLocation(ctx, &locationToCreate)
	if cerr != nil {

		if cerr.Code() == 409 { // conflict
			return nil, domain.NewCError(cerr.Code(), "location already exists")
		}

		logger.FromCtx(ctx).Error("Error creating location", zap.Error(cerr))
		return nil, domain.ErrInternal
	}

	return locationResponse, nil
}

func (ls *LocationService) GetLocation(ctx context.Context, name string) (*domain.Location, domain.CError) {
	location, cerr := ls.repo.GetLocationByName(ctx, name)
	if cerr != nil {
		if cerr.Code() == 500 {

			logger.FromCtx(ctx).Error("Error getting location", zap.Error(cerr))
			return nil, domain.ErrInternal
		}
		return nil, cerr
	}

	return location, nil
}

func (ls *LocationService) ListLocations(ctx context.Context) ([]domain.Location, domain.CError) {
	locations, cerr := ls.repo.ListLocations(ctx)
	if cerr != nil {

		logger.FromCtx(ctx).Error("Error listing location", zap.Error(cerr))
		return nil, domain.ErrInternal
	}

	return locations, nil
}

func (ls *LocationService) DeleteLocation(ctx context.Context, name string) domain.CError {
	cerr := ls.repo.DeleteLocation(ctx, name)

	if cerr != nil {
		if cerr.Code() == 500 {

			logger.FromCtx(ctx).Error("Error deleting location", zap.Error(cerr))
			return domain.ErrInternal
		}
		return cerr
	}

	return nil
}

func (ls *LocationService) GetNearestLocation(ctx context.Context, latitude, longitude float64) (*domain.NearestLocation, domain.CError) {
	nearestLocation, cerr := ls.repo.GetNearestLocation(ctx, latitude, longitude)
	if cerr != nil {
		if cerr.Code() == 404 {
			return nil, domain.NewCError(cerr.Code(), "no location found")
		}

		return nil, domain.ErrInternal
	}

	return nearestLocation, nil
}
