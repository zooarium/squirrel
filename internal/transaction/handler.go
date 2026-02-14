package transaction

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"vyaya/internal/platform/render"

	"github.com/go-chi/chi/v5"
)

// TransactionHandler handles HTTP requests for transactions.
type TransactionHandler struct {
	svc *TransactionService
}

// NewTransactionHandler creates a new transaction handler.
func NewTransactionHandler(svc *TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

// Routes returns the chi router for transaction endpoints.
func (h *TransactionHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.CreateTransaction)
	r.Get("/", h.ListTransactions)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetTransactionByID)
		r.Post("/", h.UpdateTransaction)
		r.Delete("/", h.DeleteTransaction)
	})

	return r
}

// CreateTransaction handles transaction creation.
// @Summary Create a new transaction
// @Description Create a new transaction with the provided amount, type, and user ID
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body CreateTransactionRequest true "Transaction object"
// @Success 201 {object} render.Response{data=Transaction}
// @Failure 400 {object} render.Response
// @Failure 500 {object} render.Response
// @Router /transactions [post]
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.svc.Create(r.Context(), req)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusCreated, tx)
}

// ListTransactions handles listing all transactions.
// @Summary List all transactions
// @Description Get a list of all transactions
// @Tags transactions
// @Produce json
// @Success 200 {object} render.Response{data=[]Transaction}
// @Failure 500 {object} render.Response
// @Router /transactions [get]
func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	txs, err := h.svc.List(r.Context())
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusOK, txs)
}

// GetTransactionByID handles getting a transaction by ID.
// @Summary Get transaction by ID
// @Description Get a single transaction by its ID
// @Tags transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} render.Response{data=Transaction}
// @Failure 400 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Router /transactions/{id} [get]
func (h *TransactionHandler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	tx, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			render.Error(w, http.StatusNotFound, "transaction not found")
			return
		}
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusOK, tx)
}

// UpdateTransaction handles updating a transaction.
// @Summary Update transaction by ID
// @Description Update an existing transaction with the provided amount and type
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param transaction body UpdateTransactionRequest true "Transaction object"
// @Success 200 {object} render.Response{data=Transaction}
// @Failure 400 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Router /transactions/{id} [post]
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			render.Error(w, http.StatusNotFound, "transaction not found")
			return
		}
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusOK, tx)
}

// DeleteTransaction handles deleting a transaction.
// @Summary Delete transaction by ID
// @Description Delete a transaction by its ID
// @Tags transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 204 "No Content"
// @Failure 400 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Router /transactions/{id} [delete]
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			render.Error(w, http.StatusNotFound, "transaction not found")
			return
		}
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TransactionHandler) getIDParam(r *http.Request) (int, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.Atoi(idStr)
}
