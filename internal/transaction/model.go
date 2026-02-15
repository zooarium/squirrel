package transaction

import (
	"time"
)

// Transaction represents a financial transaction.
type Transaction struct {
	ID         int       `json:"id"`
	AppID      int       `json:"app_id"`
	UserID     int       `json:"user_id"`
	Amount     float64   `json:"amount"`
	Type       string    `json:"type"` // income, expense
	CategoryID *int      `json:"category_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateTransactionRequest defines the request body for creating a transaction.
type CreateTransactionRequest struct {
	Amount     float64 `json:"amount" validate:"required"`
	Type       string  `json:"type" validate:"required,oneof=income expense"`
	CategoryID *int    `json:"category_id,omitempty"`
}

// UpdateTransactionRequest defines the request body for updating a transaction.
type UpdateTransactionRequest struct {
	Amount     float64 `json:"amount" validate:"required"`
	Type       string  `json:"type" validate:"required,oneof=income expense"`
	CategoryID *int    `json:"category_id,omitempty"`
}
