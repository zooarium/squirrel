package transaction

import (
	"context"
	"fmt"
	"vyaya/ent"
	"vyaya/ent/transaction"
)

// Repository handles database operations for transactions.
type Repository struct {
	client *ent.Client
}

// NewRepository creates a new transaction repository.
func NewRepository(client *ent.Client) *Repository {
	return &Repository{client: client}
}

// Create creates a new transaction.
func (r *Repository) Create(ctx context.Context, t Transaction) (Transaction, error) {
	builder := r.client.Transaction.
		Create().
		SetUserID(t.UserID).
		SetAmount(t.Amount).
		SetType(transaction.Type(t.Type))

	if t.CategoryID != nil {
		builder.SetCategoryID(*t.CategoryID)
	}

	entTx, err := builder.Save(ctx)
	if err != nil {
		return Transaction{}, fmt.Errorf("create transaction: %w", err)
	}

	return r.mapToModel(entTx), nil
}

// List returns all transactions.
func (r *Repository) List(ctx context.Context) ([]Transaction, error) {
	entTxs, err := r.client.Transaction.
		Query().
		Order(ent.Desc(transaction.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}

	txs := make([]Transaction, len(entTxs))
	for i, entTx := range entTxs {
		txs[i] = r.mapToModel(entTx)
	}
	return txs, nil
}

// GetByID returns a transaction by its ID.
func (r *Repository) GetByID(ctx context.Context, id int) (Transaction, error) {
	entTx, err := r.client.Transaction.
		Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return Transaction{}, ErrTransactionNotFound
		}
		return Transaction{}, fmt.Errorf("get transaction by id: %w", err)
	}

	return r.mapToModel(entTx), nil
}

// Update updates a transaction.
func (r *Repository) Update(ctx context.Context, id int, t Transaction) (Transaction, error) {
	builder := r.client.Transaction.
		UpdateOneID(id).
		SetAmount(t.Amount).
		SetType(transaction.Type(t.Type))

	if t.CategoryID != nil {
		builder.SetCategoryID(*t.CategoryID)
	} else {
		builder.ClearCategoryID()
	}

	entTx, err := builder.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return Transaction{}, ErrTransactionNotFound
		}
		return Transaction{}, fmt.Errorf("update transaction: %w", err)
	}

	return r.mapToModel(entTx), nil
}

// Delete deletes a transaction.
func (r *Repository) Delete(ctx context.Context, id int) error {
	err := r.client.Transaction.
		DeleteOneID(id).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrTransactionNotFound
		}
		return fmt.Errorf("delete transaction: %w", err)
	}
	return nil
}

func (r *Repository) mapToModel(entTx *ent.Transaction) Transaction {
	return Transaction{
		ID:         entTx.ID,
		UserID:     entTx.UserID,
		Amount:     entTx.Amount,
		Type:       string(entTx.Type),
		CategoryID: entTx.CategoryID,
		CreatedAt:  entTx.CreatedAt,
		UpdatedAt:  entTx.UpdatedAt,
	}
}
