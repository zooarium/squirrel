package transaction

import (
	"context"
	"testing"

	"vyaya/ent/enttest"
	"vyaya/internal/category"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestTransactionService(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer func() {
		_ = client.Close()
	}()

	// Create a category first
	catRepo := category.NewRepository(client)
	cat, _ := catRepo.Create(context.Background(), category.Category{AppID: 1, UserID: 1, Name: "Food", Status: 1})

	repo := NewRepository(client)
	svc := NewService(repo)
	ctx := context.Background()

	t.Run("Create Transaction", func(t *testing.T) {
		recurring := int8(1)
		req := CreateTransactionRequest{
			Amount:     100.50,
			Type:       "expense",
			CategoryID: &cat.ID,
			Recurring:  &recurring,
		}
		tx, err := svc.Create(ctx, 1, 1, req)
		assert.NoError(t, err)
		assert.Equal(t, 100.50, tx.Amount)
		assert.Equal(t, "expense", tx.Type)
		assert.Equal(t, &cat.ID, tx.CategoryID)
		assert.Equal(t, int8(1), tx.Recurring)
		assert.Equal(t, 1, tx.AppID)
	})

	t.Run("List Transactions", func(t *testing.T) {
		// Default recurring=0, so should return 0 since we only created a recurring=1 tx
		resp, err := svc.List(ctx, 1, 1, TransactionFilter{})
		assert.NoError(t, err)
		assert.Len(t, resp.Transactions, 0)

		recurring := int8(1)
		resp, err = svc.List(ctx, 1, 1, TransactionFilter{Recurring: &recurring})
		assert.NoError(t, err)
		assert.Len(t, resp.Transactions, 1)
		assert.NotEmpty(t, resp.Stats.CategoryWiseAmountSum)
		assert.Equal(t, 100.50, resp.Stats.CategoryWiseAmountSum[0].TotalSum)
		assert.Len(t, resp.Stats.Top10ByAmount, 1)

		notRecurring := int8(0)
		resp, err = svc.List(ctx, 1, 1, TransactionFilter{Recurring: &notRecurring})
		assert.NoError(t, err)
		assert.Len(t, resp.Transactions, 0)
	})

	t.Run("Get Transaction By ID", func(t *testing.T) {
		tx, err := svc.GetByID(ctx, 1, 1, 1)
		assert.NoError(t, err)
		assert.Equal(t, 100.50, tx.Amount)
	})

	t.Run("Update Transaction", func(t *testing.T) {
		recurring := int8(0)
		req := UpdateTransactionRequest{
			Amount:    200.75,
			Type:      "income",
			Recurring: &recurring,
		}
		tx, err := svc.Update(ctx, 1, 1, 1, req)
		assert.NoError(t, err)
		assert.Equal(t, 200.75, tx.Amount)
		assert.Equal(t, "income", tx.Type)
		assert.Equal(t, int8(0), tx.Recurring)
		assert.Nil(t, tx.CategoryID)
	})

	t.Run("Delete Transaction", func(t *testing.T) {
		err := svc.Delete(ctx, 1, 1, 1)
		assert.NoError(t, err)

		_, err = svc.GetByID(ctx, 1, 1, 1)
		assert.ErrorIs(t, err, ErrTransactionNotFound)
	})
}
