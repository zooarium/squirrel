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
	Recurring  int8      `json:"recurring"`
	Dated      time.Time `json:"dated"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateTransactionRequest defines the request body for creating a transaction.
type CreateTransactionRequest struct {
	Amount     float64    `json:"amount" validate:"required"`
	Type       string     `json:"type" validate:"required,oneof=income expense"`
	CategoryID *int       `json:"category_id,omitempty"`
	Recurring  *int8      `json:"recurring"`
	Dated      *time.Time `json:"dated"`
}

// UpdateTransactionRequest defines the request body for updating a transaction.
type UpdateTransactionRequest struct {
	Amount     float64    `json:"amount" validate:"required"`
	Type       string     `json:"type" validate:"required,oneof=income expense"`
	CategoryID *int       `json:"category_id,omitempty"`
	Recurring  *int8      `json:"recurring"`
	Dated      *time.Time `json:"dated"`
}

// TransactionFilter defines the filters for listing transactions.
type TransactionFilter struct {
	CategoryID *int       `json:"category_id"`
	Type       string     `json:"type"`
	Recurring  *int8      `json:"recurring"`
	Dated      string     `json:"dated"`
	From       *time.Time `json:"from"`
	To         *time.Time `json:"to"`
}

// CategoryAmountSum represents the sum of amounts for a category.
type CategoryAmountSum struct {
	CategoryID int     `json:"category_id"`
	TotalSum   float64 `json:"total_sum"`
}

// TransactionStats represents the statistics for transactions.
type TransactionStats struct {
	CategoryWiseAmountSum    []CategoryAmountSum `json:"category_wise_amount_sum"`
	CategoryTop10ByAmountSum []CategoryAmountSum `json:"category_top_10_by_amount_sum"`
	Top10ByAmount            []Transaction       `json:"top_10_by_amount"`
}

// TransactionListResponse defines the response for listing transactions.
type TransactionListResponse struct {
	Transactions []Transaction    `json:"transactions"`
	Stats        TransactionStats `json:"stats"`
}
