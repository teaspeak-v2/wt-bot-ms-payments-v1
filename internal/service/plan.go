package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/apperror"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
)

func (s *PaymentService) toPlanResponse(plan *models.Plan) *models.PlanResponse {
	return &models.PlanResponse{
		ID:          plan.ID,
		Name:        plan.Name,
		Description: plan.Description,
		PriceCents:  plan.PriceCents,
		Currency:    plan.Currency,
		Interval:    plan.Interval,
		Features:    plan.Features,
		IsActive:    plan.IsActive,
		CreatedAt:   plan.CreatedAt,
		UpdatedAt:   plan.UpdatedAt,
	}
}

// ListPlans returns a paginated list of plans.
func (s *PaymentService) ListPlans(ctx context.Context, filter models.PlanListFilter) (*models.PlanListResponse, error) {
	list, total, err := s.repo.ListPlans(ctx, filter)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	resp := &models.PlanListResponse{
		Plans:  make([]models.PlanResponse, 0, len(list)),
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}
	for i := range list {
		resp.Plans = append(resp.Plans, *s.toPlanResponse(&list[i]))
	}
	return resp, nil
}

// CreatePlan creates a new plan.
func (s *PaymentService) CreatePlan(ctx context.Context, req models.CreatePlanRequest) (*models.PlanResponse, error) {
	if !s.isAdmin(ctx) {
		return nil, apperror.Forbidden("only admins can create plans", nil)
	}
	plan := &models.Plan{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		Currency:    req.Currency,
		Interval:    req.Interval,
		Features:    req.Features,
		IsActive:    req.IsActive,
	}
	if err := s.repo.CreatePlan(ctx, plan); err != nil {
		return nil, s.mapRepoErr(err)
	}
	return s.toPlanResponse(plan), nil
}

// GetPlan returns a plan by ID.
func (s *PaymentService) GetPlan(ctx context.Context, id uuid.UUID) (*models.PlanResponse, error) {
	plan, err := s.repo.GetPlanByID(ctx, id)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	return s.toPlanResponse(plan), nil
}

// UpdatePlan updates a plan.
func (s *PaymentService) UpdatePlan(ctx context.Context, id uuid.UUID, req models.UpdatePlanRequest) (*models.PlanResponse, error) {
	if !s.isAdmin(ctx) {
		return nil, apperror.Forbidden("only admins can update plans", nil)
	}
	plan, err := s.repo.UpdatePlan(ctx, id, req)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	return s.toPlanResponse(plan), nil
}

// DeletePlan deletes a plan.
func (s *PaymentService) DeletePlan(ctx context.Context, id uuid.UUID) error {
	if !s.isAdmin(ctx) {
		return apperror.Forbidden("only admins can delete plans", nil)
	}
	if err := s.repo.DeletePlan(ctx, id); err != nil {
		return s.mapRepoErr(err)
	}
	return nil
}
