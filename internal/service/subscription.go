package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/apperror"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/repository"
)

func (s *PaymentService) toSubscriptionResponse(sub *models.Subscription) *models.SubscriptionResponse {
	return &models.SubscriptionResponse{
		ID:                 sub.ID,
		OwnerID:            sub.OwnerID,
		PlanID:             sub.PlanID,
		Status:             sub.Status,
		CurrentPeriodStart: sub.CurrentPeriodStart,
		CurrentPeriodEnd:   sub.CurrentPeriodEnd,
		CanceledAt:         sub.CanceledAt,
		CreatedAt:          sub.CreatedAt,
		UpdatedAt:          sub.UpdatedAt,
	}
}

// ListSubscriptions returns a paginated list of subscriptions.
func (s *PaymentService) ListSubscriptions(ctx context.Context, filter models.SubscriptionListFilter) (*models.SubscriptionListResponse, error) {
	if !s.isAdmin(ctx) {
		filter.OwnerID = s.userID(ctx)
	}
	list, total, err := s.repo.ListSubscriptions(ctx, filter)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	resp := &models.SubscriptionListResponse{
		Subscriptions: make([]models.SubscriptionResponse, 0, len(list)),
		Total:         total,
		Limit:         filter.Limit,
		Offset:        filter.Offset,
	}
	for i := range list {
		resp.Subscriptions = append(resp.Subscriptions, *s.toSubscriptionResponse(&list[i]))
	}
	return resp, nil
}

// CreateSubscription creates a subscription for the authenticated user.
func (s *PaymentService) CreateSubscription(ctx context.Context, req models.CreateSubscriptionRequest) (*models.SubscriptionResponse, error) {
	ownerID := s.userID(ctx)
	if ownerID == uuid.Nil {
		return nil, apperror.Unauthorized("missing authenticated user", nil)
	}
	plan, err := s.repo.GetPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	if !plan.IsActive {
		return nil, apperror.InvalidRequest("plan is not active", nil)
	}

	now := time.Now().UTC()
	var periodEnd *time.Time
	if plan.Interval != models.PlanIntervalOneTime {
		t := now.AddDate(0, 1, 0)
		if plan.Interval == models.PlanIntervalYearly {
			t = now.AddDate(1, 0, 0)
		}
		periodEnd = &t
	}

	sub := &models.Subscription{
		ID:                 uuid.New(),
		OwnerID:            ownerID,
		PlanID:             plan.ID,
		Status:             models.SubscriptionActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, s.mapRepoErr(err)
	}
	if s.cache != nil {
		_ = s.cache.DeleteSubscription(ctx, sub.ID)
	}
	return s.toSubscriptionResponse(sub), nil
}

// GetSubscription returns a subscription by ID.
func (s *PaymentService) GetSubscription(ctx context.Context, id uuid.UUID) (*models.SubscriptionResponse, error) {
	if s.cache != nil {
		if cached, err := s.cache.GetSubscription(ctx, id); err == nil {
			if err := s.checkOwnership(ctx, cached.OwnerID); err != nil {
				return nil, err
			}
			return s.toSubscriptionResponse(cached), nil
		}
	}
	sub, err := s.repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	if err := s.checkOwnership(ctx, sub.OwnerID); err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.SetSubscription(ctx, sub, 5*time.Minute)
	}
	return s.toSubscriptionResponse(sub), nil
}

// UpdateSubscription updates a subscription.
func (s *PaymentService) UpdateSubscription(ctx context.Context, id uuid.UUID, req models.UpdateSubscriptionRequest) (*models.SubscriptionResponse, error) {
	sub, err := s.repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	if !s.isAdmin(ctx) {
		if uid := s.userID(ctx); uid == uuid.Nil || sub.OwnerID != uid {
			return nil, apperror.Forbidden("not owner", nil)
		}
		// Non-admin users can only cancel.
		if req.Status != nil && *req.Status != models.SubscriptionCanceled {
			return nil, apperror.Forbidden("users can only cancel subscriptions", nil)
		}
		req.CanceledAt = ptrTime(time.Now().UTC())
	}

	updated, err := s.repo.UpdateSubscription(ctx, id, req)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	if s.cache != nil {
		_ = s.cache.DeleteSubscription(ctx, id)
	}
	return s.toSubscriptionResponse(updated), nil
}

// DeleteSubscription removes a subscription.
func (s *PaymentService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	if !s.isAdmin(ctx) {
		return apperror.Forbidden("only admins can delete subscriptions", nil)
	}
	if err := s.repo.DeleteSubscription(ctx, id); err != nil {
		return s.mapRepoErr(err)
	}
	if s.cache != nil {
		_ = s.cache.DeleteSubscription(ctx, id)
	}
	return nil
}

// GetOwnerEntitlement reports whether the given owner currently has an active
// subscription. It is intended for service-to-service calls (no ownership
// check) and never fails when the owner simply has no subscription.
func (s *PaymentService) GetOwnerEntitlement(ctx context.Context, ownerID uuid.UUID) (*models.EntitlementResponse, error) {
	if ownerID == uuid.Nil {
		return nil, apperror.InvalidRequest("owner_id is required", nil)
	}
	resp := &models.EntitlementResponse{OwnerID: ownerID}
	sub, err := s.repo.GetActiveSubscriptionByOwner(ctx, ownerID)
	if err != nil {
		if repository.IsNotFound(err) {
			return resp, nil
		}
		return nil, s.mapRepoErr(err)
	}
	resp.Active = true
	resp.Subscription = s.toSubscriptionResponse(sub)
	planID := sub.PlanID
	resp.PlanID = &planID
	if plan, err := s.repo.GetPlanByID(ctx, sub.PlanID); err == nil {
		resp.Features = plan.Features
	}
	return resp, nil
}

// GetSubscriptionStatus returns the status of a subscription for service-to-service validation.
func (s *PaymentService) GetSubscriptionStatus(ctx context.Context, id uuid.UUID) (models.SubscriptionStatus, error) {
	status, err := s.repo.GetSubscriptionStatusByID(ctx, id)
	if err != nil {
		return "", s.mapRepoErr(err)
	}
	return status, nil
}

