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

// CategoryService handles business logic for categories.
type CategoryService struct {
	repo     *CategoryRepository
	validate *validator.Validate
}

// NewCategoryService creates a new category service.
func NewCategoryService(repo *CategoryRepository) *CategoryService {
	return &CategoryService{
		repo:     repo,
		validate: validator.New(),
	}
}

// Create creates a new category.
func (s *CategoryService) Create(ctx context.Context, req CreateCategoryRequest) (Category, error) {
	if err := s.validate.Struct(req); err != nil {
		return Category{}, fmt.Errorf("validate request: %w", err)
	}

	if req.Status == 0 {
		req.Status = 1 // Default to active if not provided or 0 is passed as empty?
		// Actually validator oneof=0 1 will allow 0.
		// If req.Status is not provided, it will be 0.
		// Let's assume if it's 0 it's valid if they really want inactive.
	}

	cat := Category{
		UserID: req.UserID,
		Name:   req.Name,
		Status: req.Status,
	}

	created, err := s.repo.Create(ctx, cat)
	if err != nil {
		slog.Error("failed to create category", "error", err, "name", req.Name, "user_id", req.UserID)
		return Category{}, err
	}

	slog.Info("category created", "id", created.ID, "name", created.Name, "user_id", created.UserID)
	return created, nil
}

// List returns all categories.
func (s *CategoryService) List(ctx context.Context) ([]Category, error) {
	cats, err := s.repo.List(ctx)
	if err != nil {
		slog.Error("failed to list categories", "error", err)
		return nil, err
	}
	return cats, nil
}

// GetByID returns a category by its ID.
func (s *CategoryService) GetByID(ctx context.Context, id int) (Category, error) {
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
func (s *CategoryService) Update(ctx context.Context, id int, req UpdateCategoryRequest) (Category, error) {
	if err := s.validate.Struct(req); err != nil {
		return Category{}, fmt.Errorf("validate request: %w", err)
	}

	cat := Category{
		Name:   req.Name,
		Status: req.Status,
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
func (s *CategoryService) Delete(ctx context.Context, id int) error {
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
