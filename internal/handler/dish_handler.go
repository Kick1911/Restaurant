package handler

import (
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "github.com/kick/sigma-connected/internal/auth"
    "github.com/kick/sigma-connected/internal/dto"
    "github.com/kick/sigma-connected/internal/service"
    "github.com/kick/sigma-connected/internal/validator"
    "github.com/kick/sigma-connected/pkg/response"
)

type DishHandler struct {
    dishService *service.DishService
}

func NewDishHandler(dishService *service.DishService) *DishHandler {
    return &DishHandler{dishService: dishService}
}

func (h *DishHandler) Create(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        response.Error(w, http.StatusUnauthorized, "authentication required")
        return
    }

    var req dto.CreateDishRequest
    if err := parseJSON(r, &req); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := req.Validate(validator.Get()); err != nil {
        response.ValidationError(w, err)
        return
    }

    tenantID, err := uuid.Parse(claims.TenantID)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "invalid tenant")
        return
    }

    result, err := h.dishService.Create(r.Context(), tenantID, req)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.JSON(w, http.StatusCreated, result)
}

func (h *DishHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        response.Error(w, http.StatusUnauthorized, "authentication required")
        return
    }

    dishIDStr := chi.URLParam(r, "id")
    dishID, err := uuid.Parse(dishIDStr)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "invalid dish ID")
        return
    }

    tenantID, err := uuid.Parse(claims.TenantID)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "invalid tenant")
        return
    }

    result, err := h.dishService.GetByID(r.Context(), tenantID, dishID)
    if err != nil {
        response.Error(w, http.StatusNotFound, err.Error())
        return
    }

    response.JSON(w, http.StatusOK, result)
}

func (h *DishHandler) Update(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        response.Error(w, http.StatusUnauthorized, "authentication required")
        return
    }

    dishIDStr := chi.URLParam(r, "id")
    dishID, err := uuid.Parse(dishIDStr)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "invalid dish ID")
        return
    }

    var req dto.UpdateDishRequest
    if err := parseJSON(r, &req); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := req.Validate(validator.Get()); err != nil {
        response.ValidationError(w, err)
        return
    }

    tenantID, err := uuid.Parse(claims.TenantID)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "invalid tenant")
        return
    }

    result, err := h.dishService.Update(r.Context(), tenantID, dishID, req)
    if err != nil {
        response.Error(w, http.StatusNotFound, err.Error())
        return
    }

    response.JSON(w, http.StatusOK, result)
}

func (h *DishHandler) Delete(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        response.Error(w, http.StatusUnauthorized, "authentication required")
        return
    }

    dishIDStr := chi.URLParam(r, "id")
    dishID, err := uuid.Parse(dishIDStr)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "invalid dish ID")
        return
    }

    tenantID, err := uuid.Parse(claims.TenantID)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "invalid tenant")
        return
    }

    if err := h.dishService.Delete(r.Context(), tenantID, dishID); err != nil {
        response.Error(w, http.StatusNotFound, err.Error())
        return
    }

    response.JSON(w, http.StatusOK, map[string]string{"message": "dish deleted"})
}

func (h *DishHandler) Search(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        response.Error(w, http.StatusUnauthorized, "authentication required")
        return
    }

    tenantID, err := uuid.Parse(claims.TenantID)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "invalid tenant")
        return
    }

    query := r.URL.Query().Get("q")
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }

    dishes, total, err := h.dishService.Search(r.Context(), tenantID, query, page, limit)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.Paginated(w, dishes, page, limit, total)
}
