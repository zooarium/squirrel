package category

import (
	"context"
	"fmt"
	"squirrel/ent"
	"squirrel/ent/category"
)

// Repository handles database operations for categories.
type Repository struct {
	client *ent.Client
}

// NewRepository creates a new category repository.
func NewRepository(client *ent.Client) *Repository {
	return &Repository{client: client}
}

// Create creates a new category.
func (r *Repository) Create(ctx context.Context, c Category) (Category, error) {
	entCat, err := r.client.Category.
		Create().
		SetAppID(c.AppID).
		SetUserID(c.UserID).
		SetName(c.Name).
		SetStatus(c.Status).
		Save(ctx)
	if err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}

	return r.mapToModel(entCat), nil
}

// List returns all categories for an app with optional name filter.
func (r *Repository) List(ctx context.Context, appID, userID int, name string) ([]Category, error) {
	query := r.client.Category.
		Query().
		Where(category.AppID(appID))

	if name != "" {
		query = query.Where(category.NameContains(name))
	}

	entCats, err := query.
		Order(ent.Asc(category.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	cats := make([]Category, len(entCats))
	for i, entCat := range entCats {
		cats[i] = r.mapToModel(entCat)
	}
	return cats, nil
}

// GetByID returns a category by its ID and app ID.
func (r *Repository) GetByID(ctx context.Context, appID, userID, id int) (Category, error) {
	entCat, err := r.client.Category.
		Query().
		Where(category.ID(id), category.AppID(appID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, fmt.Errorf("get category by id: %w", err)
	}

	return r.mapToModel(entCat), nil
}

// Update updates a category.
func (r *Repository) Update(ctx context.Context, appID, userID, id int, c Category) (Category, error) {
	// Use Update() with predicate to ensure app ownership
	count, err := r.client.Category.
		Update().
		Where(category.ID(id), category.AppID(appID)).
		SetName(c.Name).
		SetStatus(c.Status).
		Save(ctx)

	if err != nil {
		return Category{}, fmt.Errorf("update category: %w", err)
	}
	if count == 0 {
		return Category{}, ErrCategoryNotFound
	}

	// Fetch updated entity to return
	return r.GetByID(ctx, appID, userID, id)
}

// Delete deletes a category.
func (r *Repository) Delete(ctx context.Context, appID, userID, id int) error {
	count, err := r.client.Category.
		Delete().
		Where(category.ID(id), category.AppID(appID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if count == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *Repository) mapToModel(entCat *ent.Category) Category {
	return Category{
		ID:        entCat.ID,
		AppID:     entCat.AppID,
		UserID:    entCat.UserID,
		Name:      entCat.Name,
		Status:    entCat.Status,
		CreatedAt: entCat.CreatedAt,
		UpdatedAt: entCat.UpdatedAt,
	}
}
