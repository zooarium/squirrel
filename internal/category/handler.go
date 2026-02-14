package category

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"vyaya/internal/platform/render"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Handler handles HTTP requests for categories.
type Handler struct {
	svc      Service
	validate *validator.Validate
}

// NewHandler creates a new category handler.
func NewHandler(svc Service) *Handler {
	return &Handler{
		svc:      svc,
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
		r.Post("/", h.Update)
		r.Delete("/", h.Delete)
	})

	return r
}

// Create handles category creation.
// @Summary Create a new category
// @Description Create a new category with the provided name and user ID
// @Tags categories
// @Accept json
// @Produce json
// @Param category body CreateCategoryRequest true "Category object"
// @Success 201 {object} render.Response{data=Category}
// @Failure 400 {object} render.Response
// @Failure 500 {object} render.Response
// @Router /categories [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.svc.Create(r.Context(), req)
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusCreated, cat)
}

// List handles listing all categories.
// @Summary List all categories
// @Description Get a list of all categories
// @Tags categories
// @Produce json
// @Success 200 {object} render.Response{data=[]Category}
// @Failure 500 {object} render.Response
// @Router /categories [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.List(r.Context())
	if err != nil {
		render.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	render.JSON(w, http.StatusOK, cats)
}

// GetByID handles getting a category by ID.
// @Summary Get category by ID
// @Description Get a single category by its ID
// @Tags categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} render.Response{data=Category}
// @Failure 400 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Router /categories/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	cat, err := h.svc.GetByID(r.Context(), id)
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
// @Description Update an existing category with the provided name and status
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body UpdateCategoryRequest true "Category object"
// @Success 200 {object} render.Response{data=Category}
// @Failure 400 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Router /categories/{id} [post]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat, err := h.svc.Update(r.Context(), id, req)
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
// @Description Delete a category by its ID
// @Tags categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 204 "No Content"
// @Failure 400 {object} render.Response
// @Failure 404 {object} render.Response
// @Failure 500 {object} render.Response
// @Router /categories/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDParam(r)
	if err != nil {
		render.Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	err = h.svc.Delete(r.Context(), id)
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
