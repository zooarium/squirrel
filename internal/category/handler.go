package category

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"squirrel/internal/platform/pagination"
	"squirrel/internal/platform/render"
	"squirrel/internal/policy"

	"keeper/pkg/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Handler handles HTTP requests for categories.
type Handler struct {
	svc      Service
	policy   *policy.Store
	validate *validator.Validate
}

// NewHandler creates a new category handler.
func NewHandler(svc Service, policyStore *policy.Store) *Handler {
	return &Handler{
		svc:      svc,
		policy:   policyStore,
		validate: validator.New(),
	}
}

// Routes returns the chi router for category endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
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

// Create handles category creation.
// @Summary Create a new category
// @Description Create a new category with the provided name
// @Tags categories
// @Accept json
// @Produce json
// @Param category body CreateCategoryRequest true "Category object"
// @Success 201 {object} render.Response{data=Category}
// @Failure 400 {object} render.Response
// @Failure 401 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /categories [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, err := h.getClaims(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	if !policy.Can(r.Context(), h.policy, claims, claims.AppID, "category", "create", "") {
		slog.Warn("create category rejected: caller lacks category.create permission", "app_id", claims.AppID, "user_id", claims.UserID)
		render.Error(w, http.StatusForbidden, "access denied")
		return
	}

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.svc.Create(r.Context(), claims.AppID, claims.UserID, claims.DivisionID, req)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusCreated, cat)
}

// List handles listing all categories.
// @Summary List all categories
// @Description Get a list of all categories for the authenticated app
// @Tags categories
// @Produce json
// @Param name query string false "Filter by category name (wildcard)"
// @Param division_id query int false "Filter by division ID"
// @Param limit query int false "Max number of results (default 50, max 500)"
// @Param offset query int false "Number of results to skip (default 0)"
// @Success 200 {object} render.Response{data=[]Category}
// @Failure 401 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /categories [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, err := h.getClaims(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	name := r.URL.Query().Get("name")

	var divisionID *int
	if didStr := r.URL.Query().Get("division_id"); didStr != "" {
		did, err := strconv.Atoi(didStr)
		if err != nil {
			render.Error(w, http.StatusBadRequest, "invalid division_id")
			return
		}
		divisionID = &did
	}

	page := pagination.Parse(r)

	cats, err := h.svc.List(r.Context(), claims.AppID, divisionID, name, page.Limit, page.Offset)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusOK, cats)
}

// GetByID handles getting a category by ID.
// @Summary Get category by ID
// @Description Get a single category by its ID if it belongs to the app
// @Tags categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} render.Response{data=Category}
// @Failure 400 {object} render.Response
// @Failure 401 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /categories/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, err := h.getClaims(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	cat, err := h.svc.GetByID(r.Context(), claims.AppID, id)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			render.Error(w, http.StatusNotFound, "category not found")
			return
		}
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusOK, cat)
}

// Update handles updating a category.
// @Summary Update category by ID
// @Description Update an existing category if it belongs to the app
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body UpdateCategoryRequest true "Category object"
// @Success 200 {object} render.Response{data=Category}
// @Failure 400 {object} render.Response
// @Failure 401 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /categories/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims, err := h.getClaims(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	if !policy.Can(r.Context(), h.policy, claims, claims.AppID, "category", "update", "") {
		slog.Warn("update category rejected: caller lacks category.update permission", "id", id, "user_id", claims.UserID)
		render.Error(w, http.StatusForbidden, "access denied")
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.svc.Update(r.Context(), claims.AppID, id, req)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			render.Error(w, http.StatusNotFound, "category not found")
			return
		}
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusOK, cat)
}

// Delete handles deleting a category.
// @Summary Delete category by ID
// @Description Delete a category by its ID if it belongs to the app
// @Tags categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 204 "No Content"
// @Failure 400 {object} render.Response
// @Failure 401 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Security Bearer
// @Router /categories/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, err := h.getClaims(r)
	if err != nil {
		render.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	if !policy.Can(r.Context(), h.policy, claims, claims.AppID, "category", "delete", "") {
		slog.Warn("delete category rejected: caller lacks category.delete permission", "id", id, "user_id", claims.UserID)
		render.Error(w, http.StatusForbidden, "access denied")
		return
	}

	err = h.svc.Delete(r.Context(), claims.AppID, id)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			render.Error(w, http.StatusNotFound, "category not found")
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
