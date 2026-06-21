package handler

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "github.com/kick/sigma-connected/internal/auth"
    "github.com/kick/sigma-connected/internal/dto"
    "github.com/kick/sigma-connected/internal/service"
    "github.com/kick/sigma-connected/internal/validator"
    "github.com/kick/sigma-connected/pkg/response"
)

type RatingHandler struct {
    ratingService *service.RatingService
}

func NewRatingHandler(ratingService *service.RatingService) *RatingHandler {
    return &RatingHandler{ratingService: ratingService}
}

func (h *RatingHandler) Create(w http.ResponseWriter, r *http.Request) {
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

    var req dto.CreateRatingRequest
    if err := parseJSON(r, &req); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := req.Validate(validator.Get()); err != nil {
        response.ValidationError(w, err)
        return
    }

    userID, err := uuid.Parse(claims.UserID)
    if err != nil {
        response.Error(w, http.StatusBadRequest, "invalid user")
        return
    }

    result, err := h.ratingService.Create(r.Context(), dishID, userID, req)
    if err != nil {
        response.Error(w, http.StatusBadRequest, err.Error())
        return
    }

    response.JSON(w, http.StatusCreated, result)
}

func (h *RatingHandler) GetByDishID(w http.ResponseWriter, r *http.Request) {
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

    ratings, err := h.ratingService.GetByDishID(r.Context(), dishID)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, err.Error())
        return
    }

    response.JSON(w, http.StatusOK, ratings)
}
