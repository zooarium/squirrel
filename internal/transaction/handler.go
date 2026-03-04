package transaction

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"squirrel/internal/platform/render"

	"keeper/pkg/auth"

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
	r.Get("/stats", h.Stats)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetByID)
		r.Put("/", h.Update)
		r.Delete("/", h.Delete)
	})

	return r
}

func (h *Handler) getClaims(r *http.Request) (*auth.UserClaims, error) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		return nil, errors.New("user not authenticated")
	}
	return claims, nil
}

// Create handles transaction creation.
// @Summary Create a new transaction
// @Description Create a new transaction with the provided amount, type, category, recurring status and dated
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
	claims, err := h.getClaims(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := h.svc.Create(r.Context(), claims.AppID, claims.UserID, req)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusCreated, tx)
}

// List handles listing all transactions.
// @Summary List all transactions
// @Description Get a list of all transactions for the authenticated app with optional filtering. Stats in response only include expenses.
// @Tags transactions
// @Produce json
// @Param category_id query int false "Filter by category ID"
// @Param type query string false "Filter by type (income, expense)"
// @Param recurring query int false "Filter by recurring status (0 or 1, default 0)"
// @Param dated query string false "Filter by predefined date ranges (today, yesterday, this month, last month, this year, last year)"
// @Param from query string false "Filter from date (YYYY-MM-DD)"
// @Param to query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} render.Response{data=TransactionListResponse}
// @Failure 401 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /transactions [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, err := h.getClaims(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	filter := h.parseFilter(r)

	txs, err := h.svc.List(r.Context(), claims.AppID, claims.UserID, filter)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusOK, txs)
}

// Stats handles transaction statistics.
// @Summary Get transaction statistics
// @Description Get statistics for transactions for the authenticated app with optional filtering. Only expenses are included.
// @Tags transactions
// @Produce json
// @Param category_id query int false "Filter by category ID"
// @Param recurring query int false "Filter by recurring status (0 or 1, default 0)"
// @Param dated query string false "Filter by predefined date ranges (today, yesterday, this month, last month, this year, last year)"
// @Param from query string false "Filter from date (YYYY-MM-DD)"
// @Param to query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} render.Response{data=TransactionStats}
// @Failure 401 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /transactions/stats [get]
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	claims, err := h.getClaims(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	filter := h.parseFilter(r)

	stats, err := h.svc.Stats(r.Context(), claims.AppID, claims.UserID, filter)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusOK, stats)
}

func (h *Handler) parseFilter(r *http.Request) TransactionFilter {
	filter := TransactionFilter{}

	if val := r.URL.Query().Get("category_id"); val != "" {
		if id, err := strconv.Atoi(val); err == nil {
			filter.CategoryID = &id
		}
	}

	filter.Type = r.URL.Query().Get("type")

	if val := r.URL.Query().Get("recurring"); val != "" {
		if i, err := strconv.ParseInt(val, 10, 8); err == nil {
			i8 := int8(i)
			filter.Recurring = &i8
		}
	}

	filter.Dated = r.URL.Query().Get("dated")

	if val := r.URL.Query().Get("from"); val != "" {
		if t, err := time.Parse("2006-01-02", val); err == nil {
			filter.From = &t
		}
	}

	if val := r.URL.Query().Get("to"); val != "" {
		if t, err := time.Parse("2006-01-02", val); err == nil {
			filter.To = &t
		}
	}

	return filter
}

// GetByID handles getting a transaction by ID.
// @Summary Get transaction by ID
// @Description Get a single transaction by its ID if it belongs to the app
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
	claims, err := h.getClaims(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	tx, err := h.svc.GetByID(r.Context(), claims.AppID, claims.UserID, id)
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
// @Description Update an existing transaction including amount, type, category, recurring status and dated if it belongs to the app
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
// @Router /transactions/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims, err := h.getClaims(r)
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

	tx, err := h.svc.Update(r.Context(), claims.AppID, claims.UserID, id, req)
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
// @Description Delete a transaction by its ID if it belongs to the app
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
	claims, err := h.getClaims(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid transaction ID")
		return
	}

	err = h.svc.Delete(r.Context(), claims.AppID, claims.UserID, id)
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
