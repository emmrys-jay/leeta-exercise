package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"leeta/internal/adapter/config"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"

	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// migrationsFS is a filesystem that embeds the migrations folder
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

/**
 * DB is a wrapper for PostgreSQL database connection
 * that uses pgxpool as database driver.
 * It also holds a reference to squirrel.StatementBuilderType
 * which is used to build SQL queries that compatible with PostgreSQL syntax
 */
type DB struct {
	*pgxpool.Pool
	QueryBuilder *squirrel.StatementBuilderType
	url          string
}

func dsn(config *config.DatabaseConfiguration) string {
	url := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable",
		config.Protocol,
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Name,
	)

	return url
}

// New creates a new PostgreSQL database instance
func New(ctx context.Context, config *config.DatabaseConfiguration) (*DB, error) {
	url := dsn(config)
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}

	err = db.Ping(ctx)
	if err != nil {
		return nil, err
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return &DB{
		db,
		&psql,
		url,
	}, nil
}

// Migrate runs the database migration
func (db *DB) Migrate() error {
	driver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("Error connecting to db: %w", err)
	}
	defer driver.Close()

	migrations, err := migrate.NewWithSourceInstance("iofs", driver, db.url)
	if err != nil {
		return err
	}

	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Error connecting to db: %w", err)
	}

	return nil
}

// ErrorCode returns the error code of the given error
func (db *DB) ErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return "0000"
}

// Close closes the database connection
func (db *DB) Close() {
	db.Pool.Close()
}
