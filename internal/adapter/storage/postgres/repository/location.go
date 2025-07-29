package repository

import (
	"context"
	"time"

	"savely/internal/adapter/storage/postgres"
	"savely/internal/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

/**
 * UserRepository implements port.UserRepository interface
 * and provides an access to the postgres database
 */
type UserRepository struct {
	db *postgres.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *postgres.DB) *UserRepository {
	return &UserRepository{
		db,
	}
}

func (ur *UserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, domain.CError) {
	query := ur.db.QueryBuilder.Insert("users").
		Columns("first_name", "last_name", "email", "password", "role", "is_active").
		Values(user.FirstName, user.LastName, user.Email, user.Password, user.Role, user.IsActive).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, domain.NewInternalCError(err.Error())
	}

	err = ur.db.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		// 23505 is the error code for a unique conflict error
		if errCode := ur.db.ErrorCode(err); errCode == "23505" {
			return nil, domain.ErrConflictingData
		}
		return nil, domain.NewInternalCError(err.Error())
	}

	return user, nil
}

// GetUserByID gets a user by ID from the database
func (ur *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, domain.CError) {
	var user domain.User

	query := ur.db.QueryBuilder.Select("*").
		From("users").
		Where(sq.Eq{"id": id, "deleted_at": nil}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, domain.NewInternalCError(err.Error())
	}

	err = ur.db.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, domain.NewInternalCError(err.Error())
	}

	return &user, nil
}

// GetUserByEmail gets a user by email from the database
func (ur *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, domain.CError) {
	var user domain.User

	query := ur.db.QueryBuilder.Select("*").
		From("users").
		Where(sq.Eq{"email": email, "deleted_at": nil}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, domain.NewInternalCError(err.Error())
	}

	err = ur.db.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, domain.NewInternalCError(err.Error())
	}

	return &user, nil
}

// ListUsers lists all users from the database
func (ur *UserRepository) ListUsers(ctx context.Context) ([]domain.User, domain.CError) {
	var user domain.User
	var users []domain.User

	query := ur.db.QueryBuilder.Select("*").
		From("users").
		Where(sq.Eq{"deleted_at": nil}).
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
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, domain.NewInternalCError(err.Error())
		}

		users = append(users, user)
	}

	return users, nil
}

// UpdateUser updates a user by ID in the database
func (ur *UserRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, domain.CError) {
	query := ur.db.QueryBuilder.Update("users").
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": user.ID, "deleted_at": nil}).
		Suffix("RETURNING *")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, domain.NewInternalCError(err.Error())
	}

	err = ur.db.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		if errCode := ur.db.ErrorCode(err); errCode == "23505" {
			return nil, domain.ErrConflictingData
		}
		return nil, domain.NewInternalCError(err.Error())
	}

	return user, nil
}

// DeleteUser deletes a user by ID from the database
func (ur *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) domain.CError {
	query := ur.db.QueryBuilder.Update("users").
		Set("deleted_at", time.Now()).
		Set("is_active", false).
		Where(sq.Eq{"id": id})

	sql, args, err := query.ToSql()
	if err != nil {
		return domain.NewInternalCError(err.Error())
	}

	_, err = ur.db.Exec(ctx, sql, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.ErrDataNotFound
		}
		return domain.NewInternalCError(err.Error())
	}

	return nil
}
