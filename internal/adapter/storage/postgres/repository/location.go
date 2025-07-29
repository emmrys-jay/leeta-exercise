package repository

import (
	"context"

	"leeta/internal/adapter/storage/postgres"
	"leeta/internal/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
)

/**
 * UserRepository implements port.UserRepository interface
 * and provides an access to the postgres database
 */
type LocationRepository struct {
	db *postgres.DB
}

// NewLocationRepository creates a new location repository instance
func NewLocationRepository(db *postgres.DB) *LocationRepository {
	return &LocationRepository{
		db,
	}
}

func (ur *LocationRepository) CreateLocation(ctx context.Context, location *domain.Location) (*domain.Location, domain.CError) {

	query := `
		INSERT INTO locations (name, slug, latitude, longitude, geo) 
		VALUES ($1, $2, $3, $4, ST_MakePoint($4, $3)::geography) 
		RETURNING id, name, slug, latitude, longitude, created_at
	`

	err := ur.db.QueryRow(
		ctx, query, location.Name, slug.Make(location.Name), location.Latitude, location.Longitude,
	).Scan(
		&location.ID, &location.Name, &location.Slug, &location.Latitude, &location.Longitude,
		&location.CreatedAt,
	)

	if err != nil {
		// 23505 is the error code for a unique conflict error
		if errCode := ur.db.ErrorCode(err); errCode == "23505" {
			return nil, domain.ErrConflictingData
		}

		return nil, domain.NewInternalCError(err.Error())
	}

	return location, nil
}

// GetUserByID gets a user by ID from the database
func (ur *LocationRepository) GetLocationByID(ctx context.Context, id string) (*domain.Location, domain.CError) {
	var location domain.Location

	query := ur.db.QueryBuilder.Select("id", "name", "slug", "latitude", "longitude", "created_at").
		From("locations").
		Where(sq.Eq{"id": id}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, domain.NewInternalCError(err.Error())
	}

	err = ur.db.QueryRow(ctx, sql, args...).Scan(
		&location.ID,
		&location.Name,
		&location.Slug,
		&location.Latitude,
		&location.Longitude,
		&location.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, domain.NewInternalCError(err.Error())
	}

	return &location, nil
}

// GetLocationBySlug gets a location by slug from the database
func (ur *LocationRepository) GetLocationByName(ctx context.Context, name string) (*domain.Location, domain.CError) {
	var location domain.Location

	query := ur.db.QueryBuilder.Select("id", "name", "slug", "latitude", "longitude", "created_at").
		From("locations").
		Where(sq.Or{sq.Eq{"name": name}, sq.Eq{"slug": slug.Make(name)}}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, domain.NewInternalCError(err.Error())
	}

	err = ur.db.QueryRow(ctx, sql, args...).Scan(
		&location.ID,
		&location.Name,
		&location.Slug,
		&location.Latitude,
		&location.Longitude,
		&location.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, domain.NewInternalCError(err.Error())
	}

	return &location, nil
}

// ListLocations lists all locations from the database
func (ur *LocationRepository) ListLocations(ctx context.Context) ([]domain.Location, domain.CError) {
	var location domain.Location
	var locations []domain.Location

	query := ur.db.QueryBuilder.Select("id", "name", "slug", "latitude", "longitude", "created_at").
		From("locations").
		OrderBy("created_at DESC")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, domain.NewInternalCError(err.Error())
	}

	rows, err := ur.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, domain.NewInternalCError(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(
			&location.ID,
			&location.Name,
			&location.Slug,
			&location.Latitude,
			&location.Longitude,
			&location.CreatedAt,
		)
		if err != nil {
			return nil, domain.NewInternalCError(err.Error())
		}

		locations = append(locations, location)
	}

	return locations, nil
}

// DeleteLocation deletes a location by name or slug from the database
func (ur *LocationRepository) DeleteLocation(ctx context.Context, name string) domain.CError {
	query := ur.db.QueryBuilder.Delete("locations").
		Where(sq.Or{sq.Eq{"name": name}, sq.Eq{"slug": slug.Make(name)}})

	sql, args, err := query.ToSql()
	if err != nil {
		return domain.NewInternalCError(err.Error())
	}

	_, err = ur.db.Exec(ctx, sql, args...)
	if err != nil {
		return domain.NewInternalCError(err.Error())
	}

	return nil
}

// GetNearestLocation gets the nearest location from the database
func (ur *LocationRepository) GetNearestLocation(ctx context.Context, latitude, longitude float64) (*domain.NearestLocation, domain.CError) {
	var location domain.NearestLocation

	query := `
		SELECT id, name, slug, latitude, longitude, created_at,
		ST_Distance(geo, ST_MakePoint($1, $2)::geography) AS distance_meters
		FROM locations
		ORDER BY distance_meters
		LIMIT 1
	`

	err := ur.db.QueryRow(ctx, query, longitude, latitude).Scan(
		&location.ID, &location.Name, &location.Slug, &location.Latitude,
		&location.Longitude, &location.CreatedAt, &location.Distance,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, domain.NewInternalCError(err.Error())
	}

	return &location, nil
}
