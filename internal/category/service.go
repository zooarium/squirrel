package category

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
)

var (
	// ErrCategoryNotFound is returned when a category is not found.
	ErrCategoryNotFound = errors.New("category not found")
)

// Service defines the business logic for categories.
type Service interface {
	Create(ctx context.Context, userID int, req CreateCategoryRequest) (Category, error)
	List(ctx context.Context, userID int) ([]Category, error)
	GetByID(ctx context.Context, userID, id int) (Category, error)
	Update(ctx context.Context, userID, id int, req UpdateCategoryRequest) (Category, error)
	Delete(ctx context.Context, userID, id int) error
}

type service struct {
	repo     *Repository
	validate *validator.Validate
}

// NewService creates a new category service.
func NewService(repo *Repository) Service {
	return &service{
		repo:     repo,
		validate: validator.New(),
	}
}

// Create creates a new category.
func (s *service) Create(ctx context.Context, userID int, req CreateCategoryRequest) (Category, error) {
	if err := s.validate.Struct(req); err != nil {
		return Category{}, fmt.Errorf("validate request: %w", err)
	}

	if req.Status == 0 {
		req.Status = 1
	}

	cat := Category{
		UserID: userID,
		Name:   req.Name,
		Status: req.Status,
	}

	created, err := s.repo.Create(ctx, cat)
	if err != nil {
		slog.Error("failed to create category", "error", err, "name", req.Name, "user_id", userID)
		return Category{}, err
	}

	slog.Info("category created", "id", created.ID, "name", created.Name, "user_id", userID)
	return created, nil
}

// List returns all categories for a user.
func (s *service) List(ctx context.Context, userID int) ([]Category, error) {
	cats, err := s.repo.List(ctx, userID)
	if err != nil {
		slog.Error("failed to list categories", "error", err, "user_id", userID)
		return nil, err
	}
	return cats, nil
}

// GetByID returns a category by its ID and user ID.
func (s *service) GetByID(ctx context.Context, userID, id int) (Category, error) {
	cat, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		if !errors.Is(err, ErrCategoryNotFound) {
			slog.Error("failed to get category by id", "error", err, "id", id, "user_id", userID)
		}
		return Category{}, err
	}
	return cat, nil
}

// Update updates an existing category.
func (s *service) Update(ctx context.Context, userID, id int, req UpdateCategoryRequest) (Category, error) {
	if err := s.validate.Struct(req); err != nil {
		return Category{}, fmt.Errorf("validate request: %w", err)
	}

	cat := Category{
		Name:   req.Name,
		Status: req.Status,
	}

	updated, err := s.repo.Update(ctx, userID, id, cat)
	if err != nil {
		if !errors.Is(err, ErrCategoryNotFound) {
			slog.Error("failed to update category", "error", err, "id", id, "user_id", userID)
		}
		return Category{}, err
	}

	slog.Info("category updated", "id", updated.ID, "name", updated.Name)
	return updated, nil
}

// Delete deletes a category by its ID.
func (s *service) Delete(ctx context.Context, userID, id int) error {
	err := s.repo.Delete(ctx, userID, id)
	if err != nil {
		if !errors.Is(err, ErrCategoryNotFound) {
			slog.Error("failed to delete category", "error", err, "id", id, "user_id", userID)
		}
		return err
	}

	slog.Info("category deleted", "id", id)
	return nil
}
