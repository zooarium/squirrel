package category

import (
	"time"
)

// Category represents a category in the system.
type Category struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name" validate:"required"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCategoryRequest defines the request body for creating a category.
type CreateCategoryRequest struct {
	UserID int    `json:"user_id" validate:"required"`
	Name   string `json:"name" validate:"required"`
	Status int8   `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateCategoryRequest defines the request body for updating a category.
type UpdateCategoryRequest struct {
	Name   string `json:"name" validate:"required"`
	Status int8   `json:"status" validate:"omitempty,oneof=0 1"`
}
