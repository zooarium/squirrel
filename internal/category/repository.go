package category

import (
	"context"
	"fmt"
	"vyaya/ent"
	"vyaya/ent/category"
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
		SetName(c.Name).
		Save(ctx)
	if err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}

	return r.mapToModel(entCat), nil
}

// List returns all categories.
func (r *Repository) List(ctx context.Context) ([]Category, error) {
	entCats, err := r.client.Category.
		Query().
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

// GetByID returns a category by its ID.
func (r *Repository) GetByID(ctx context.Context, id int) (Category, error) {
	entCat, err := r.client.Category.
		Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, fmt.Errorf("get category by id: %w", err)
	}

	return r.mapToModel(entCat), nil
}

// Update updates a category.
func (r *Repository) Update(ctx context.Context, id int, c Category) (Category, error) {
	entCat, err := r.client.Category.
		UpdateOneID(id).
		SetName(c.Name).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, fmt.Errorf("update category: %w", err)
	}

	return r.mapToModel(entCat), nil
}

// Delete deletes a category.
func (r *Repository) Delete(ctx context.Context, id int) error {
	err := r.client.Category.
		DeleteOneID(id).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}

func (r *Repository) mapToModel(entCat *ent.Category) Category {
	return Category{
		ID:        entCat.ID,
		Name:      entCat.Name,
		CreatedAt: entCat.CreatedAt,
		UpdatedAt: entCat.UpdatedAt,
	}
}
