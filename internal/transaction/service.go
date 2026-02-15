package transaction

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	// ErrTransactionNotFound is returned when a transaction is not found.
	ErrTransactionNotFound = errors.New("transaction not found")
)

// Service defines the business logic for transactions.
type Service interface {
	Create(ctx context.Context, appID, userID int, req CreateTransactionRequest) (Transaction, error)
	List(ctx context.Context, appID, userID int, filter TransactionFilter) (TransactionListResponse, error)
	GetByID(ctx context.Context, appID, userID, id int) (Transaction, error)
	Update(ctx context.Context, appID, userID, id int, req UpdateTransactionRequest) (Transaction, error)
	Delete(ctx context.Context, appID, userID, id int) error
}

type service struct {
	repo     *Repository
	validate *validator.Validate
}

// NewService creates a new transaction service.
func NewService(repo *Repository) Service {
	return &service{
		repo:     repo,
		validate: validator.New(),
	}
}

// Create creates a new transaction.
func (s *service) Create(ctx context.Context, appID, userID int, req CreateTransactionRequest) (Transaction, error) {
	if err := s.validate.Struct(req); err != nil {
		return Transaction{}, fmt.Errorf("validate request: %w", err)
	}

	tx := Transaction{
		AppID:      appID,
		UserID:     userID,
		Amount:     req.Amount,
		Type:       req.Type,
		CategoryID: req.CategoryID,
	}

	if req.Recurring != nil {
		tx.Recurring = *req.Recurring
	}

	if req.Dated != nil {
		tx.Dated = *req.Dated
	} else {
		tx.Dated = time.Now()
	}

	created, err := s.repo.Create(ctx, tx)
	if err != nil {
		slog.Error("failed to create transaction", "error", err, "app_id", appID, "user_id", userID)
		return Transaction{}, err
	}

	slog.Info("transaction created", "id", created.ID, "app_id", appID, "user_id", created.UserID)
	return created, nil
}

// List returns all transactions for a user.
func (s *service) List(ctx context.Context, appID, userID int, filter TransactionFilter) (TransactionListResponse, error) {
	if filter.Recurring == nil {
		var zero int8 = 0
		filter.Recurring = &zero
	}

	txs, err := s.repo.List(ctx, appID, userID, filter)
	if err != nil {
		slog.Error("failed to list transactions", "error", err, "app_id", appID, "user_id", userID)
		return TransactionListResponse{}, err
	}

	stats, err := s.repo.GetStats(ctx, appID, userID, filter)
	if err != nil {
		slog.Error("failed to get transaction stats", "error", err, "app_id", appID, "user_id", userID)
		return TransactionListResponse{}, err
	}

	return TransactionListResponse{
		Transactions: txs,
		Stats:        stats,
	}, nil
}

// GetByID returns a transaction by its ID.
func (s *service) GetByID(ctx context.Context, appID, userID, id int) (Transaction, error) {
	tx, err := s.repo.GetByID(ctx, appID, userID, id)
	if err != nil {
		if !errors.Is(err, ErrTransactionNotFound) {
			slog.Error("failed to get transaction by id", "error", err, "id", id, "app_id", appID, "user_id", userID)
		}
		return Transaction{}, err
	}
	return tx, nil
}

// Update updates an existing transaction.
func (s *service) Update(ctx context.Context, appID, userID, id int, req UpdateTransactionRequest) (Transaction, error) {
	if err := s.validate.Struct(req); err != nil {
		return Transaction{}, fmt.Errorf("validate request: %w", err)
	}

	tx := Transaction{
		Amount:     req.Amount,
		Type:       req.Type,
		CategoryID: req.CategoryID,
	}

	if req.Recurring != nil {
		tx.Recurring = *req.Recurring
	}

	if req.Dated != nil {
		tx.Dated = *req.Dated
	}

	updated, err := s.repo.Update(ctx, appID, userID, id, tx)
	if err != nil {
		if !errors.Is(err, ErrTransactionNotFound) {
			slog.Error("failed to update transaction", "error", err, "id", id, "app_id", appID, "user_id", userID)
		}
		return Transaction{}, err
	}

	slog.Info("transaction updated", "id", updated.ID, "app_id", appID, "user_id", userID)
	return updated, nil
}

// Delete deletes a transaction by its ID.
func (s *service) Delete(ctx context.Context, appID, userID, id int) error {
	err := s.repo.Delete(ctx, appID, userID, id)
	if err != nil {
		if !errors.Is(err, ErrTransactionNotFound) {
			slog.Error("failed to delete transaction", "error", err, "id", id, "app_id", appID, "user_id", userID)
		}
		return err
	}

	slog.Info("transaction deleted", "id", id, "app_id", appID, "user_id", userID)
	return nil
}
