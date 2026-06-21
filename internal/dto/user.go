package dto

import "github.com/go-playground/validator/v10"

type RegisterRequest struct {
    TenantSlug string `json:"tenant_slug" validate:"required,min=2,max=50"`
    Name       string `json:"name" validate:"required,min=2,max=100"`
    Email      string `json:"email" validate:"required,email"`
    Password   string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
    TenantSlug string `json:"tenant_slug" validate:"required,min=2,max=50"`
    Email      string `json:"email" validate:"required,email"`
    Password   string `json:"password" validate:"required"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  UserResponse `json:"user"`
}

type UserResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Role  string `json:"role"`
}

func (r RegisterRequest) Validate(validate *validator.Validate) error {
    return validate.Struct(r)
}

func (r LoginRequest) Validate(validate *validator.Validate) error {
    return validate.Struct(r)
}
