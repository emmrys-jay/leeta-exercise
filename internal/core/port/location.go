package port

import (
	"context"

	"leeta/internal/core/domain"
)

// LocationRepository is an interface for interacting with Location-related data
type LocationRepository interface {
	// CreateLocation inserts a new location into the database
	CreateLocation(ctx context.Context, location *domain.Location) (*domain.Location, domain.CError)
	// GetLocationByID fetches a new location from the database using it's id
	GetLocationByID(ctx context.Context, id string) (*domain.Location, domain.CError)
	// GetLocationByName fetches a new location from the database using it's name
	GetLocationByName(ctx context.Context, name string) (*domain.Location, domain.CError)
	// ListLocations fetches and returns all locations in the database
	ListLocations(ctx context.Context) ([]domain.Location, domain.CError)
	// DeleteLocation performs a soft delete on a location specified by its name or slug
	DeleteLocation(ctx context.Context, name string) domain.CError
	// GetNearestLocation fetches the nearest location to the longitude and latitude from the database
	GetNearestLocation(ctx context.Context, latitude, longitude float64) (*domain.NearestLocation, domain.CError)
}

// LocationService is an interface for interacting with Location-related business logic
type LocationService interface {
	// RegisterLocation is used to register a new location. It returns the new location after saving it
	RegisterLocation(ctx context.Context, location *domain.RegisterLocationRequest) (*domain.Location, domain.CError)
	// GetLocation returns a location specified by its id
	GetLocation(ctx context.Context, id string) (*domain.Location, domain.CError)
	// ListLocations returns all locations in the system
	ListLocations(ctx context.Context) ([]domain.Location, domain.CError)
	// DeleteLocation deletes a location specified by id
	DeleteLocation(ctx context.Context, id string) domain.CError
	// GetNearestLocation returns the nearest location to the longitude and latitude
	GetNearestLocation(ctx context.Context, latitude, longitude float64) (*domain.NearestLocation, domain.CError)
}
