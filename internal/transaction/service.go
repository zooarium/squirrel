package transaction

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

var (
	// ErrTransactionNotFound is returned when a transaction is not found.
	ErrTransactionNotFound = errors.New("transaction not found")
)

// Service handles business logic for transactions.
type Service struct {
	repo     *Repository
	validate *validator.Validate
}

// NewService creates a new transaction service.
func NewService(repo *Repository) *Service {
	return &Service{
		repo:     repo,
		validate: validator.New(),
	}
}

// Create creates a new transaction.
func (s *Service) Create(ctx context.Context, req CreateTransactionRequest) (Transaction, error) {
	if err := s.validate.Struct(req); err != nil {
		return Transaction{}, fmt.Errorf("validate request: %w", err)
	}

	tx := Transaction{
		UserID:     req.UserID,
		Amount:     req.Amount,
		Type:       req.Type,
		CategoryID: req.CategoryID,
	}

	created, err := s.repo.Create(ctx, tx)
	if err != nil {
		slog.Error("failed to create transaction", "error", err, "user_id", req.UserID)
		return Transaction{}, err
	}

	slog.Info("transaction created", "id", created.ID, "user_id", created.UserID)
	return created, nil
}

// List returns all transactions.
func (s *Service) List(ctx context.Context) ([]Transaction, error) {
	txs, err := s.repo.List(ctx)
	if err != nil {
		slog.Error("failed to list transactions", "error", err)
		return nil, err
	}
	return txs, nil
}

// GetByID returns a transaction by its ID.
func (s *Service) GetByID(ctx context.Context, id int) (Transaction, error) {
	tx, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if !errors.Is(err, ErrTransactionNotFound) {
			slog.Error("failed to get transaction by id", "error", err, "id", id)
		}
		return Transaction{}, err
	}
	return tx, nil
}

// Update updates an existing transaction.
func (s *Service) Update(ctx context.Context, id int, req UpdateTransactionRequest) (Transaction, error) {
	if err := s.validate.Struct(req); err != nil {
		return Transaction{}, fmt.Errorf("validate request: %w", err)
	}

	tx := Transaction{
		Amount:     req.Amount,
		Type:       req.Type,
		CategoryID: req.CategoryID,
	}

	updated, err := s.repo.Update(ctx, id, tx)
	if err != nil {
		if !errors.Is(err, ErrTransactionNotFound) {
			slog.Error("failed to update transaction", "error", err, "id", id)
		}
		return Transaction{}, err
	}

	slog.Info("transaction updated", "id", updated.ID)
	return updated, nil
}

// Delete deletes a transaction by its ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if !errors.Is(err, ErrTransactionNotFound) {
			slog.Error("failed to delete transaction", "error", err, "id", id)
		}
		return err
	}

	slog.Info("transaction deleted", "id", id)
	return nil
}
