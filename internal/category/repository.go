package category

import (
	"context"
	"fmt"

	"squirrel/ent"
	"squirrel/ent/category"
)

type categoryRepository struct {
	client *ent.Client
}

// NewRepository creates a new category repository.
func NewRepository(client *ent.Client) *categoryRepository {
	return &categoryRepository{client: client}
}

func (r *categoryRepository) Create(ctx context.Context, c Category) (Category, error) {
	builder := r.client.Category.
		Create().
		SetAppID(c.AppID).
		SetUserID(c.UserID).
		SetName(c.Name).
		SetStatus(c.Status)

	if c.DivisionID != nil {
		builder = builder.SetDivisionID(*c.DivisionID)
	}

	entCat, err := builder.Save(ctx)
	if err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}

	return r.mapToModel(entCat), nil
}

func (r *categoryRepository) List(ctx context.Context, appID int, divisionID *int, name string, limit, offset int) ([]Category, error) {
	query := r.client.Category.
		Query().
		Where(category.AppID(appID))

	if divisionID != nil {
		query = query.Where(category.DivisionID(*divisionID))
	}

	if name != "" {
		query = query.Where(category.NameContains(name))
	}

	entCats, err := query.
		Order(ent.Asc(category.FieldName), ent.Asc(category.FieldID)).
		Limit(limit).
		Offset(offset).
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

func (r *categoryRepository) GetByID(ctx context.Context, appID, id int) (Category, error) {
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

func (r *categoryRepository) Update(ctx context.Context, appID, id int, c Category) (Category, error) {
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

	return r.GetByID(ctx, appID, id)
}

func (r *categoryRepository) Delete(ctx context.Context, appID, id int) error {
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

func (r *categoryRepository) mapToModel(entCat *ent.Category) Category {
	return Category{
		ID:         entCat.ID,
		AppID:      entCat.AppID,
		UserID:     entCat.UserID,
		DivisionID: entCat.DivisionID,
		Name:       entCat.Name,
		Status:     entCat.Status,
		CreatedAt:  entCat.CreatedAt,
		UpdatedAt:  entCat.UpdatedAt,
	}
}
