package category

import (
	"context"
	"testing"

	"squirrel/ent/enttest"

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
		cat, err := svc.Create(ctx, 1, 1, req)
		assert.NoError(t, err)
		assert.Equal(t, "Test Category", cat.Name)
		assert.Equal(t, 1, cat.AppID)
		assert.Equal(t, 1, cat.UserID)
		assert.Equal(t, int8(1), cat.Status)
	})

	t.Run("List Categories", func(t *testing.T) {
		cats, err := svc.List(ctx, 1, 1, "")
		assert.NoError(t, err)
		assert.Len(t, cats, 1)
	})

	t.Run("Get Category By ID", func(t *testing.T) {
		cat, err := svc.GetByID(ctx, 1, 1, 1)
		assert.NoError(t, err)
		assert.Equal(t, "Test Category", cat.Name)
	})

	t.Run("Update Category", func(t *testing.T) {
		req := UpdateCategoryRequest{
			Name:   "Updated Category",
			Status: 0,
		}
		cat, err := svc.Update(ctx, 1, 1, 1, req)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Category", cat.Name)
		assert.Equal(t, int8(0), cat.Status)
	})

	t.Run("Multi-user and Multi-app Isolation", func(t *testing.T) {
		// App 1, User 1 creates a category
		req1 := CreateCategoryRequest{Name: "App1 User1 Cat", Status: 1}
		cat1, err := svc.Create(ctx, 1, 1, req1)
		assert.NoError(t, err)

		// App 1, User 2 should be able to see App 1, User 1's category
		cats, err := svc.List(ctx, 1, 2, "")
		assert.NoError(t, err)
		found := false
		for _, c := range cats {
			if c.ID == cat1.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "App 1 User 2 should see category created by User 1")

		// App 1, User 2 should be able to get App 1, User 1's category by ID
		catById, err := svc.GetByID(ctx, 1, 2, cat1.ID)
		assert.NoError(t, err)
		assert.Equal(t, cat1.ID, catById.ID)

		// App 2, User 3 should NOT be able to see App 1's category
		catsApp2, err := svc.List(ctx, 2, 3, "")
		assert.NoError(t, err)
		for _, c := range catsApp2 {
			assert.NotEqual(t, cat1.ID, c.ID)
		}

		// App 2, User 3 should NOT be able to get App 1's category by ID
		_, err = svc.GetByID(ctx, 2, 3, cat1.ID)
		assert.ErrorIs(t, err, ErrCategoryNotFound)

		// App 1, User 2 should be able to update App 1, User 1's category
		updateReq := UpdateCategoryRequest{Name: "Shared Update", Status: 1}
		updated, err := svc.Update(ctx, 1, 2, cat1.ID, updateReq)
		assert.NoError(t, err)
		assert.Equal(t, "Shared Update", updated.Name)

		// App 1, User 2 should be able to delete App 1, User 1's category
		err = svc.Delete(ctx, 1, 2, cat1.ID)
		assert.NoError(t, err)
	})
}
