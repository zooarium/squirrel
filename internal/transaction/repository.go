package transaction

import (
	"context"
	"fmt"
	"sort"
	"time"

	"squirrel/ent"
	"squirrel/ent/transaction"
)

type transactionRepository struct {
	client *ent.Client
}

// NewRepository creates a new transaction repository.
func NewRepository(client *ent.Client) *transactionRepository {
	return &transactionRepository{client: client}
}

func (r *transactionRepository) Create(ctx context.Context, t Transaction) (Transaction, error) {
	builder := r.client.Transaction.
		Create().
		SetAppID(t.AppID).
		SetUserID(t.UserID).
		SetAmount(t.Amount).
		SetType(transaction.Type(t.Type)).
		SetRecurring(t.Recurring).
		SetDated(t.Dated)

	if t.CategoryID != nil {
		builder.SetCategoryID(*t.CategoryID)
	}

	entTx, err := builder.Save(ctx)
	if err != nil {
		return Transaction{}, fmt.Errorf("create transaction: %w", err)
	}

	return r.mapToModel(entTx), nil
}

func (r *transactionRepository) List(ctx context.Context, appID, userID int, filter TransactionFilter) ([]Transaction, error) {
	query := r.buildFilteredQuery(appID, userID, filter)

	entTxs, err := query.
		Order(ent.Desc(transaction.FieldDated)).
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

func (r *transactionRepository) GetStats(ctx context.Context, appID, userID int, filter TransactionFilter) (TransactionStats, error) {
	stats := TransactionStats{}

	var v []struct {
		CategoryID *int    `json:"category_id"`
		Sum        float64 `json:"sum"`
	}
	err := r.buildFilteredQuery(appID, userID, filter).
		GroupBy(transaction.FieldCategoryID).
		Aggregate(ent.Sum(transaction.FieldAmount)).
		Scan(ctx, &v)
	if err != nil {
		return stats, fmt.Errorf("category wise amount sum: %w", err)
	}

	stats.CategoryWiseAmountSum = make([]CategoryAmountSum, len(v))
	for i, item := range v {
		catID := 0
		if item.CategoryID != nil {
			catID = *item.CategoryID
		}
		stats.CategoryWiseAmountSum[i] = CategoryAmountSum{
			CategoryID: catID,
			TotalSum:   item.Sum,
		}
	}

	entTopTxs, err := r.buildFilteredQuery(appID, userID, filter).
		Order(ent.Desc(transaction.FieldAmount)).
		Limit(10).
		All(ctx)
	if err != nil {
		return stats, fmt.Errorf("top 10 by amount: %w", err)
	}

	stats.Top10ByAmount = make([]Transaction, len(entTopTxs))
	for i, entTx := range entTopTxs {
		stats.Top10ByAmount[i] = r.mapToModel(entTx)
	}

	top10 := make([]CategoryAmountSum, len(stats.CategoryWiseAmountSum))
	copy(top10, stats.CategoryWiseAmountSum)

	sort.Slice(top10, func(i, j int) bool {
		return top10[i].TotalSum > top10[j].TotalSum
	})

	if len(top10) > 10 {
		top10 = top10[:10]
	}
	stats.CategoryTop10ByAmountSum = top10

	return stats, nil
}

func (r *transactionRepository) buildFilteredQuery(appID, userID int, filter TransactionFilter) *ent.TransactionQuery {
	query := r.client.Transaction.
		Query().
		Where(transaction.AppID(appID))

	if filter.CategoryID != nil {
		query = query.Where(transaction.CategoryID(*filter.CategoryID))
	}

	if filter.Type != "" {
		query = query.Where(transaction.TypeEQ(transaction.Type(filter.Type)))
	}

	if filter.Recurring != nil {
		query = query.Where(transaction.Recurring(*filter.Recurring))
	}

	if filter.Dated != "" {
		now := time.Now()
		var start, end time.Time
		switch filter.Dated {
		case "today":
			start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			end = start.AddDate(0, 0, 1)
		case "yesterday":
			end = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			start = end.AddDate(0, 0, -1)
		case "this month":
			start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			end = start.AddDate(0, 1, 0)
		case "last month":
			start = time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
			end = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		case "this year":
			start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
			end = start.AddDate(1, 0, 0)
		case "last year":
			start = time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location())
			end = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		}
		if !start.IsZero() {
			query = query.Where(transaction.DatedGTE(start), transaction.DatedLT(end))
		}
	}

	if filter.From != nil {
		query = query.Where(transaction.DatedGTE(*filter.From))
	}
	if filter.To != nil {
		query = query.Where(transaction.DatedLTE(*filter.To))
	}

	return query
}

func (r *transactionRepository) GetByID(ctx context.Context, appID, userID, id int) (Transaction, error) {
	entTx, err := r.client.Transaction.
		Query().
		Where(transaction.ID(id), transaction.AppID(appID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return Transaction{}, ErrTransactionNotFound
		}
		return Transaction{}, fmt.Errorf("get transaction by id: %w", err)
	}

	return r.mapToModel(entTx), nil
}

func (r *transactionRepository) Update(ctx context.Context, appID, userID, id int, t Transaction) (Transaction, error) {
	builder := r.client.Transaction.
		Update().
		Where(transaction.ID(id), transaction.AppID(appID)).
		SetAmount(t.Amount).
		SetType(transaction.Type(t.Type)).
		SetRecurring(t.Recurring).
		SetDated(t.Dated)

	if t.CategoryID != nil {
		builder.SetCategoryID(*t.CategoryID)
	} else {
		builder.ClearCategoryID()
	}

	count, err := builder.Save(ctx)
	if err != nil {
		return Transaction{}, fmt.Errorf("update transaction: %w", err)
	}
	if count == 0 {
		return Transaction{}, ErrTransactionNotFound
	}

	return r.GetByID(ctx, appID, userID, id)
}

func (r *transactionRepository) Delete(ctx context.Context, appID, userID, id int) error {
	count, err := r.client.Transaction.
		Delete().
		Where(transaction.ID(id), transaction.AppID(appID)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	if count == 0 {
		return ErrTransactionNotFound
	}
	return nil
}

func (r *transactionRepository) mapToModel(entTx *ent.Transaction) Transaction {
	return Transaction{
		ID:         entTx.ID,
		AppID:      entTx.AppID,
		UserID:     entTx.UserID,
		Amount:     entTx.Amount,
		Type:       string(entTx.Type),
		CategoryID: entTx.CategoryID,
		Recurring:  entTx.Recurring,
		Dated:      entTx.Dated,
		CreatedAt:  entTx.CreatedAt,
		UpdatedAt:  entTx.UpdatedAt,
	}
}
