package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Enum for user roles
type UserRole string

const (
	RAdmin UserRole = "admin"
	RUser  UserRole = "user"
)

var StringToUserRole = map[string]UserRole{
	"admin": RAdmin,
	"user":  RUser,
}

func (e *UserRole) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = UserRole(s)
	case string:
		*e = UserRole(s)
	default:
		return fmt.Errorf("unsupported scan type for UsersRoleEnum: %T", src)
	}
	return nil
}

func (ur UserRole) String() string {
	return string(ur)
}

// User represents a row in the "users" table
type User struct {
	ID        uuid.UUID  `json:"id"`
	FirstName string     `json:"first_name" validate:"required"`
	LastName  string     `json:"last_name" validate:"required"`
	Email     string     `json:"email" validate:"required"`
	Password  string     `json:"password,omitempty" validate:"required"`
	Role      UserRole   `json:"role"`
	IsActive  bool       `json:"is_active" validate:"required"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Role      string `json:"-" swaggerignore:"true"`
}

type CreateUserRequest struct {
	Email     string   `json:"email" validate:"required"`
	Password  string   `json:"password" validate:"required"`
	FirstName string   `json:"first_name" validate:"required"`
	LastName  string   `json:"last_name" validate:"required"`
	Role      UserRole `json:"-"`
}
