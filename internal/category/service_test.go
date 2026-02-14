package category

import (
	"context"
	"testing"

	"vyaya/ent/enttest"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestCategoryService(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() {
		_ = client.Close()
	}()

	repo := NewRepository(client)
	svc := NewService(repo)
	ctx := context.Background()

	t.Run("Create Category", func(t *testing.T) {
		req := CreateCategoryRequest{
			Name:   "Test Category",
			Status: 1,
		}
		cat, err := svc.Create(ctx, 1, req)
		assert.NoError(t, err)
		assert.Equal(t, "Test Category", cat.Name)
		assert.Equal(t, 1, cat.UserID)
		assert.Equal(t, int8(1), cat.Status)
	})

	t.Run("List Categories", func(t *testing.T) {
		cats, err := svc.List(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, cats, 1)
	})

	t.Run("Get Category By ID", func(t *testing.T) {
		cat, err := svc.GetByID(ctx, 1, 1)
		assert.NoError(t, err)
		assert.Equal(t, "Test Category", cat.Name)
	})

	t.Run("Update Category", func(t *testing.T) {
		req := UpdateCategoryRequest{
			Name:   "Updated Category",
			Status: 0,
		}
		cat, err := svc.Update(ctx, 1, 1, req)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Category", cat.Name)
		assert.Equal(t, int8(0), cat.Status)
	})

	t.Run("Delete Category", func(t *testing.T) {
		err := svc.Delete(ctx, 1, 1)
		assert.NoError(t, err)

		_, err = svc.GetByID(ctx, 1, 1)
		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})
}
