package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kick/sigma-connected/internal/auth"
	"github.com/kick/sigma-connected/internal/domain"
	"github.com/kick/sigma-connected/internal/dto"
	"github.com/kick/sigma-connected/internal/middleware"
	"github.com/kick/sigma-connected/internal/repository"
)

type UserService struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
	bfProtector *middleware.BruteForceProtector
	jwtSecret  string
	jwtExpiry  time.Duration
	db         *sqlx.DB
}

func NewUserService(
	userRepo *repository.UserRepository,
	tenantRepo *repository.TenantRepository,
	bfProtector *middleware.BruteForceProtector,
	jwtSecret string,
	jwtExpiry time.Duration,
	db *sqlx.DB,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		tenantRepo:  tenantRepo,
		bfProtector: bfProtector,
		jwtSecret:   jwtSecret,
		jwtExpiry:   jwtExpiry,
		db:          db,
	}
}

func (s *UserService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	tenant, err := s.tenantRepo.FindBySlug(ctx, req.TenantSlug)
	if err != nil {
		tenant = &domain.Tenant{
			ID:   uuid.New(),
			Name: req.TenantSlug,
			Slug: req.TenantSlug,
		}
		if err := s.tenantRepo.Create(ctx, tenant); err != nil {
			return nil, errors.New("failed to create tenant")
		}
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		ID:           uuid.New(),
		TenantID:     tenant.ID,
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         domain.RoleCustomer,
	}

	if req.TenantSlug == "admin" {
		user.Role = domain.RoleAdmin
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.New("email already registered for this tenant")
	}

	token, err := auth.GenerateToken(s.jwtSecret, user.ID, tenant.ID, user.Email, string(user.Role), s.jwtExpiry)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
			Role:  string(user.Role),
		},
	}, nil
}

func (s *UserService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	locked, err := s.bfProtector.IsLocked(ctx, req.Email)
	if err != nil {
		return nil, errors.New("internal error")
	}
	if locked {
		return nil, errors.New("account is temporarily locked due to too many failed attempts. Try again later")
	}

	tenant, err := s.tenantRepo.FindBySlug(ctx, req.TenantSlug)
	if err != nil {
		return nil, errors.New("invalid tenant or credentials")
	}

	user, err := s.userRepo.FindByEmail(ctx, tenant.ID, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		s.bfProtector.RecordFailedAttempt(ctx, req.Email)
		return nil, errors.New("invalid email or password")
	}

	s.bfProtector.ResetAttempts(ctx, req.Email)

	token, err := auth.GenerateToken(s.jwtSecret, user.ID, tenant.ID, user.Email, string(user.Role), s.jwtExpiry)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
			Role:  string(user.Role),
		},
	}, nil
}
