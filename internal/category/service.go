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

// Repository defines the data access contract for categories.
type Repository interface {
	Create(ctx context.Context, c Category) (Category, error)
	List(ctx context.Context, appID int, divisionID *int, name string, limit, offset int) ([]Category, error)
	GetByID(ctx context.Context, appID, id int) (Category, error)
	Update(ctx context.Context, appID, id int, c Category) (Category, error)
	Delete(ctx context.Context, appID, id int) error
}

// Service defines the business logic for categories.
type Service interface {
	Create(ctx context.Context, appID, userID, divisionID int, req CreateCategoryRequest) (Category, error)
	List(ctx context.Context, appID int, divisionID *int, name string, limit, offset int) ([]Category, error)
	GetByID(ctx context.Context, appID, id int) (Category, error)
	Update(ctx context.Context, appID, id int, req UpdateCategoryRequest) (Category, error)
	Delete(ctx context.Context, appID, id int) error
}

type service struct {
	repo     Repository
	validate *validator.Validate
}

// NewService creates a new category service.
func NewService(repo Repository) Service {
	return &service{
		repo:     repo,
		validate: validator.New(),
	}
}

// Create creates a new category.
func (s *service) Create(ctx context.Context, appID, userID, divisionID int, req CreateCategoryRequest) (Category, error) {
	if err := s.validate.Struct(req); err != nil {
		return Category{}, fmt.Errorf("validate request: %w", err)
	}

	if req.Status == 0 {
		req.Status = 1
	}

	cat := Category{
		AppID:      appID,
		UserID:     userID,
		DivisionID: &divisionID,
		Name:       req.Name,
		Status:     req.Status,
	}

	created, err := s.repo.Create(ctx, cat)
	if err != nil {
		slog.Error("failed to create category", "error", err, "name", req.Name, "app_id", appID, "user_id", userID, "division_id", divisionID)
		return Category{}, err
	}

	slog.Info("category created", "id", created.ID, "name", created.Name, "app_id", appID, "user_id", userID, "division_id", divisionID)
	return created, nil
}

// List returns all categories for an app, optionally filtered by division.
func (s *service) List(ctx context.Context, appID int, divisionID *int, name string, limit, offset int) ([]Category, error) {
	cats, err := s.repo.List(ctx, appID, divisionID, name, limit, offset)
	if err != nil {
		slog.Error("failed to list categories", "error", err, "app_id", appID, "division_id", divisionID, "name", name)
		return nil, err
	}
	return cats, nil
}

// GetByID returns a category by its ID.
func (s *service) GetByID(ctx context.Context, appID, id int) (Category, error) {
	cat, err := s.repo.GetByID(ctx, appID, id)
	if err != nil {
		if !errors.Is(err, ErrCategoryNotFound) {
			slog.Error("failed to get category by id", "error", err, "id", id, "app_id", appID)
		}
		return Category{}, err
	}
	return cat, nil
}

// Update updates an existing category.
func (s *service) Update(ctx context.Context, appID, id int, req UpdateCategoryRequest) (Category, error) {
	if err := s.validate.Struct(req); err != nil {
		return Category{}, fmt.Errorf("validate request: %w", err)
	}

	cat := Category{
		Name:   req.Name,
		Status: req.Status,
	}

	updated, err := s.repo.Update(ctx, appID, id, cat)
	if err != nil {
		if !errors.Is(err, ErrCategoryNotFound) {
			slog.Error("failed to update category", "error", err, "id", id, "app_id", appID)
		}
		return Category{}, err
	}

	slog.Info("category updated", "id", updated.ID, "name", updated.Name, "app_id", appID)
	return updated, nil
}

// Delete deletes a category by its ID.
func (s *service) Delete(ctx context.Context, appID, id int) error {
	err := s.repo.Delete(ctx, appID, id)
	if err != nil {
		if !errors.Is(err, ErrCategoryNotFound) {
			slog.Error("failed to delete category", "error", err, "id", id, "app_id", appID)
		}
		return err
	}

	slog.Info("category deleted", "id", id, "app_id", appID)
	return nil
}
