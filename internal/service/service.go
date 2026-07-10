package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/apperror"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/cache"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/httpserver/middleware"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/repository"
)

// PaymentService handles business logic for payments, subscriptions, and plans.
type PaymentService struct {
	repo        *repository.Repository
	cache       *cache.Client
	encKey      string
}

// NewPaymentService creates a new service.
func NewPaymentService(repo *repository.Repository, cache *cache.Client, encKey string) *PaymentService {
	return &PaymentService{repo: repo, cache: cache, encKey: encKey}
}

func (s *PaymentService) isAdmin(ctx context.Context) bool {
	return middleware.Role(ctx) == models.RoleAdmin
}

func (s *PaymentService) userID(ctx context.Context) uuid.UUID {
	id, _ := uuid.Parse(middleware.UserID(ctx))
	return id
}

func (s *PaymentService) checkOwnership(ctx context.Context, ownerID uuid.UUID) error {
	if s.isAdmin(ctx) {
		return nil
	}
	if uid := s.userID(ctx); uid != uuid.Nil && ownerID == uid {
		return nil
	}
	return apperror.Forbidden("not owner", nil)
}

func (s *PaymentService) mapRepoErr(err error) error {
	if repository.IsNotFound(err) {
		return apperror.NotFound("resource not found", err)
	}
	if repository.IsUniqueViolation(err) {
		return apperror.Conflict("resource already exists", err)
	}
	return apperror.Internal("internal server error", err)
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
