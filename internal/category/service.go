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

// Service handles business logic for categories.
type Service struct {
	repo     *Repository
	validate *validator.Validate
}

// NewService creates a new category service.
func NewService(repo *Repository) *Service {
	return &Service{
		repo:     repo,
		validate: validator.New(),
	}
}

// Create creates a new category.
func (s *Service) Create(ctx context.Context, req CreateCategoryRequest) (Category, error) {
	if err := s.validate.Struct(req); err != nil {
		return Category{}, fmt.Errorf("validate request: %w", err)
	}

	cat := Category{
		Name: req.Name,
	}

	created, err := s.repo.Create(ctx, cat)
	if err != nil {
		slog.Error("failed to create category", "error", err, "name", req.Name)
		return Category{}, err
	}

	slog.Info("category created", "id", created.ID, "name", created.Name)
	return created, nil
}

// List returns all categories.
func (s *Service) List(ctx context.Context) ([]Category, error) {
	cats, err := s.repo.List(ctx)
	if err != nil {
		slog.Error("failed to list categories", "error", err)
		return nil, err
	}
	return cats, nil
}

// GetByID returns a category by its ID.
func (s *Service) GetByID(ctx context.Context, id int) (Category, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if !errors.Is(err, ErrCategoryNotFound) {
			slog.Error("failed to get category by id", "error", err, "id", id)
		}
		return Category{}, err
	}
	return cat, nil
}

// Update updates an existing category.
func (s *Service) Update(ctx context.Context, id int, req UpdateCategoryRequest) (Category, error) {
	if err := s.validate.Struct(req); err != nil {
		return Category{}, fmt.Errorf("validate request: %w", err)
	}

	cat := Category{
		Name: req.Name,
	}

	updated, err := s.repo.Update(ctx, id, cat)
	if err != nil {
		if !errors.Is(err, ErrCategoryNotFound) {
			slog.Error("failed to update category", "error", err, "id", id)
		}
		return Category{}, err
	}

	slog.Info("category updated", "id", updated.ID, "name", updated.Name)
	return updated, nil
}

// Delete deletes a category by its ID.
func (s *Service) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if !errors.Is(err, ErrCategoryNotFound) {
			slog.Error("failed to delete category", "error", err, "id", id)
		}
		return err
	}

	slog.Info("category deleted", "id", id)
	return nil
}
