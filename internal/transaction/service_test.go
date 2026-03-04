package transaction

import (
	"context"
	"testing"

	"squirrel/ent/enttest"
	"squirrel/internal/category"

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

	t.Run("Stats Only Expense", func(t *testing.T) {
		// Currently we have one transaction with ID 1 which is "income" (200.75)
		// Let's create an expense
		req := CreateTransactionRequest{
			Amount: 50.0,
			Type:   "expense",
		}
		_, err := svc.Create(ctx, 1, 1, req)
		assert.NoError(t, err)

		// Stats should only show the expense (50.0), not the income (200.75)
		stats, err := svc.Stats(ctx, 1, 1, TransactionFilter{})
		assert.NoError(t, err)

		totalSum := 0.0
		for _, s := range stats.CategoryWiseAmountSum {
			totalSum += s.TotalSum
		}
		assert.Equal(t, 50.0, totalSum)
		assert.Len(t, stats.Top10ByAmount, 1)
		assert.Equal(t, 50.0, stats.Top10ByAmount[0].Amount)
	})

	t.Run("Multi-user and Multi-app Isolation", func(t *testing.T) {
		// App 1, User 1 creates a transaction
		req1 := CreateTransactionRequest{Amount: 50.0, Type: "expense"}
		tx1, err := svc.Create(ctx, 1, 1, req1)
		assert.NoError(t, err)

		// App 1, User 2 should be able to see App 1, User 1's transaction
		resp, err := svc.List(ctx, 1, 2, TransactionFilter{})
		assert.NoError(t, err)
		found := false
		for _, tx := range resp.Transactions {
			if tx.ID == tx1.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "App 1 User 2 should see transaction created by User 1")

		// App 2, User 3 should NOT be able to see App 1's transaction
		respApp2, err := svc.List(ctx, 2, 3, TransactionFilter{})
		assert.NoError(t, err)
		for _, tx := range respApp2.Transactions {
			assert.NotEqual(t, tx1.ID, tx.ID)
		}

		// App 2, User 3 should NOT be able to get App 1's transaction by ID
		_, err = svc.GetByID(ctx, 2, 3, tx1.ID)
		assert.ErrorIs(t, err, ErrTransactionNotFound)

		// App 1, User 2 should be able to update App 1, User 1's transaction
		updateReq := UpdateTransactionRequest{Amount: 75.0, Type: "expense"}
		updated, err := svc.Update(ctx, 1, 2, tx1.ID, updateReq)
		assert.NoError(t, err)
		assert.Equal(t, 75.0, updated.Amount)

		// App 1, User 2 should be able to delete App 1, User 1's transaction
		err = svc.Delete(ctx, 1, 2, tx1.ID)
		assert.NoError(t, err)
	})
}
