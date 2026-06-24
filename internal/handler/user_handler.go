package handler

import (
	"net/http"

	"github.com/kick/sigma-connected/internal/dto"
	"github.com/kick/sigma-connected/internal/service"
	"github.com/kick/sigma-connected/internal/validator"
	"github.com/kick/sigma-connected/pkg/response"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := parseJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(validator.Get()); err != nil {
		response.ValidationError(w, err)
		return
	}

	result, err := h.userService.Register(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := parseJSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(validator.Get()); err != nil {
		response.ValidationError(w, err)
		return
	}

	result, err := h.userService.Login(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, result)
}
