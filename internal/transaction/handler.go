package transaction

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"vyaya/internal/platform/render"

	"dvarapala/pkg/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Handler handles HTTP requests for transactions.
type Handler struct {
	svc      Service
	validate *validator.Validate
}

// NewHandler creates a new transaction handler.
func NewHandler(svc Service) *Handler {
	return &Handler{
		svc:      svc,
		validate: validator.New(),
	}
}

// Routes returns the chi router for transaction endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetByID)
		r.Post("/", h.Update)
		r.Delete("/", h.Delete)
	})

	return r
}

func (h *Handler) getUserID(r *http.Request) (int, error) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		return 0, errors.New("user not authenticated")
	}
	return claims.UserID, nil
}

// Create handles transaction creation.
// @Summary Create a new transaction
// @Description Create a new transaction with the provided amount, type
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body CreateTransactionRequest true "Transaction object"
// @Success 201 {object} render.Response{data=Transaction}
// @Failure 400 {object} render.Response
// @Failure 401 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /transactions [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusCreated, tx)
}

// List handles listing all transactions.
// @Summary List all transactions
// @Description Get a list of all transactions for the authenticated user
// @Tags transactions
// @Produce json
// @Success 200 {object} render.Response{data=[]Transaction}
// @Failure 401 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /transactions [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	txs, err := h.svc.List(r.Context(), userID)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusOK, txs)
}

// GetByID handles getting a transaction by ID.
// @Summary Get transaction by ID
// @Description Get a single transaction by its ID if it belongs to the user
// @Tags transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} render.Response{data=Transaction}
// @Failure 400 {object} render.Response
// @Failure 401 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /transactions/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	tx, err := h.svc.GetByID(r.Context(), userID, id)
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

// Update handles updating a transaction.
// @Summary Update transaction by ID
// @Description Update an existing transaction if it belongs to the user
// @Tags transactions
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param transaction body UpdateTransactionRequest true "Transaction object"
// @Success 200 {object} render.Response{data=Transaction}
// @Failure 400 {object} render.Response
// @Failure 401 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /transactions/{id} [post]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

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

	tx, err := h.svc.Update(r.Context(), userID, id, req)
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

// Delete handles deleting a transaction.
// @Summary Delete transaction by ID
// @Description Delete a transaction by its ID if it belongs to the user
// @Tags transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 204 "No Content"
// @Failure 400 {object} render.Response
// @Failure 401 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /transactions/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	err = h.svc.Delete(r.Context(), userID, id)
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

func (h *Handler) getIDParam(r *http.Request) (int, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.Atoi(idStr)
}
