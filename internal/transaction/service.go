package transaction

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	// ErrTransactionNotFound is returned when a transaction is not found.
	ErrTransactionNotFound = errors.New("transaction not found")
)

// Repository defines the data access contract for transactions.
type Repository interface {
	Create(ctx context.Context, t Transaction) (Transaction, error)
	List(ctx context.Context, appID, userID int, filter TransactionFilter) ([]Transaction, error)
	GetStats(ctx context.Context, appID, userID int, filter TransactionFilter) (TransactionStats, error)
	GetByID(ctx context.Context, appID, userID, id int) (Transaction, error)
	Update(ctx context.Context, appID, userID, id int, t Transaction) (Transaction, error)
	Delete(ctx context.Context, appID, userID, id int) error
}

// Service defines the business logic for transactions.
type Service interface {
	Create(ctx context.Context, appID, userID, divisionID int, req CreateTransactionRequest) (Transaction, error)
	List(ctx context.Context, appID, userID int, filter TransactionFilter) (TransactionListResponse, error)
	Stats(ctx context.Context, appID, userID int, filter TransactionFilter) (TransactionStats, error)
	GetByID(ctx context.Context, appID, userID, id int) (Transaction, error)
	Update(ctx context.Context, appID, userID, id int, req UpdateTransactionRequest) (Transaction, error)
	Delete(ctx context.Context, appID, userID, id int) error
}

// StatsCache is the subset of an in-memory cache the service relies on for
// caching transaction stats. A nil StatsCache disables caching.
type StatsCache interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Purge()
}

type service struct {
	repo       Repository
	validate   *validator.Validate
	statsCache StatsCache
}

// NewService creates a new transaction service.
//
// statsCache may be nil, in which case stats are always computed from the
// repository (no caching).
func NewService(repo Repository, statsCache StatsCache) Service {
	return &service{
		repo:       repo,
		validate:   validator.New(),
		statsCache: statsCache,
	}
}

// statsCacheKey builds a deterministic cache key from app_id and all stats
// filter parameters. Pagination (Limit/Offset) is intentionally excluded
// because stats are computed over the full filtered set.
func statsCacheKey(appID, userID int, filter TransactionFilter) string {
	var b strings.Builder
	b.WriteString("stats:app=")
	b.WriteString(strconv.Itoa(appID))
	b.WriteString("|user=")
	b.WriteString(strconv.Itoa(userID))
	b.WriteString("|cat=")
	if filter.CategoryID != nil {
		b.WriteString(strconv.Itoa(*filter.CategoryID))
	}
	b.WriteString("|type=")
	b.WriteString(filter.Type)
	b.WriteString("|recurring=")
	if filter.Recurring != nil {
		b.WriteString(strconv.Itoa(int(*filter.Recurring)))
	}
	b.WriteString("|dated=")
	b.WriteString(filter.Dated)
	b.WriteString("|from=")
	if filter.From != nil {
		b.WriteString(filter.From.UTC().Format(time.RFC3339))
	}
	b.WriteString("|to=")
	if filter.To != nil {
		b.WriteString(filter.To.UTC().Format(time.RFC3339))
	}
	return b.String()
}

// getStats returns transaction stats, using the cache when available.
func (s *service) getStats(ctx context.Context, appID, userID int, filter TransactionFilter) (TransactionStats, error) {
	if s.statsCache != nil {
		key := statsCacheKey(appID, userID, filter)
		if v, ok := s.statsCache.Get(key); ok {
			if stats, ok := v.(TransactionStats); ok {
				return stats, nil
			}
		}
		stats, err := s.repo.GetStats(ctx, appID, userID, filter)
		if err != nil {
			return TransactionStats{}, err
		}
		s.statsCache.Set(key, stats)
		return stats, nil
	}

	return s.repo.GetStats(ctx, appID, userID, filter)
}

// invalidateStats purges cached stats after a write. Purging the whole cache
// is the simplest correct strategy and keeps stale aggregates from being served.
func (s *service) invalidateStats() {
	if s.statsCache != nil {
		s.statsCache.Purge()
	}
}

// Create creates a new transaction.
func (s *service) Create(ctx context.Context, appID, userID, divisionID int, req CreateTransactionRequest) (Transaction, error) {
	if err := s.validate.Struct(req); err != nil {
		return Transaction{}, fmt.Errorf("validate request: %w", err)
	}

	tx := Transaction{
		AppID:      appID,
		UserID:     userID,
		DivisionID: &divisionID,
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
		slog.Error("failed to create transaction", "error", err, "app_id", appID, "user_id", userID, "division_id", divisionID)
		return Transaction{}, err
	}

	s.invalidateStats()

	slog.Info("transaction created", "id", created.ID, "app_id", appID, "user_id", created.UserID, "division_id", divisionID)
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

	statsFilter := filter
	statsFilter.Type = "expense"
	statsFilter.Limit = 0
	statsFilter.Offset = 0
	stats, err := s.getStats(ctx, appID, userID, statsFilter)
	if err != nil {
		slog.Error("failed to get transaction stats", "error", err, "app_id", appID, "user_id", userID)
		return TransactionListResponse{}, err
	}

	return TransactionListResponse{
		Transactions: txs,
		Stats:        stats,
	}, nil
}

// Stats returns transaction statistics.
func (s *service) Stats(ctx context.Context, appID, userID int, filter TransactionFilter) (TransactionStats, error) {
	filter.Type = "expense"
	filter.Limit = 0
	filter.Offset = 0
	stats, err := s.getStats(ctx, appID, userID, filter)
	if err != nil {
		slog.Error("failed to get transaction stats", "error", err, "app_id", appID, "user_id", userID)
		return TransactionStats{}, err
	}
	return stats, nil
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

	s.invalidateStats()

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

	s.invalidateStats()

	slog.Info("transaction deleted", "id", id, "app_id", appID, "user_id", userID)
	return nil
}
