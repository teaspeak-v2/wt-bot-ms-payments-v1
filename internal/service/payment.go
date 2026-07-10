package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/apperror"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
)

func (s *PaymentService) toPaymentResponse(payment *models.Payment) *models.PaymentResponse {
	return &models.PaymentResponse{
		ID:                payment.ID,
		OwnerID:           payment.OwnerID,
		SubscriptionID:    payment.SubscriptionID,
		AmountCents:       payment.AmountCents,
		Currency:          payment.Currency,
		Status:            payment.Status,
		Provider:          payment.Provider,
		ProviderPaymentID: payment.ProviderPaymentID,
		Metadata:          payment.Metadata,
		CreatedAt:         payment.CreatedAt,
		UpdatedAt:         payment.UpdatedAt,
	}
}

// ListPayments returns a paginated list of payments.
func (s *PaymentService) ListPayments(ctx context.Context, filter models.PaymentListFilter) (*models.PaymentListResponse, error) {
	if !s.isAdmin(ctx) {
		filter.OwnerID = s.userID(ctx)
	}
	list, total, err := s.repo.ListPayments(ctx, filter)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	resp := &models.PaymentListResponse{
		Payments: make([]models.PaymentResponse, 0, len(list)),
		Total:    total,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
	}
	for i := range list {
		resp.Payments = append(resp.Payments, *s.toPaymentResponse(&list[i]))
	}
	return resp, nil
}

// CreatePayment creates a payment record.
func (s *PaymentService) CreatePayment(ctx context.Context, req models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	ownerID := s.userID(ctx)
	if ownerID == uuid.Nil {
		return nil, apperror.Unauthorized("missing authenticated user", nil)
	}

	if req.SubscriptionID != nil && *req.SubscriptionID != uuid.Nil {
		sub, err := s.repo.GetSubscriptionByID(ctx, *req.SubscriptionID)
		if err != nil {
			return nil, s.mapRepoErr(err)
		}
		if err := s.checkOwnership(ctx, sub.OwnerID); err != nil {
			return nil, err
		}
	}

	payment := &models.Payment{
		ID:                uuid.New(),
		OwnerID:           ownerID,
		SubscriptionID:    req.SubscriptionID,
		AmountCents:       req.AmountCents,
		Currency:          req.Currency,
		Status:            models.PaymentPending,
		Provider:          req.Provider,
		ProviderPaymentID: req.ProviderPaymentID,
		Metadata:          req.Metadata,
	}

	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, s.mapRepoErr(err)
	}
	return s.toPaymentResponse(payment), nil
}

// GetPayment returns a payment by ID.
func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*models.PaymentResponse, error) {
	payment, err := s.repo.GetPaymentByID(ctx, id)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	if err := s.checkOwnership(ctx, payment.OwnerID); err != nil {
		return nil, err
	}
	return s.toPaymentResponse(payment), nil
}

// UpdatePayment updates a payment.
func (s *PaymentService) UpdatePayment(ctx context.Context, id uuid.UUID, req models.UpdatePaymentRequest) (*models.PaymentResponse, error) {
	payment, err := s.repo.GetPaymentByID(ctx, id)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	if !s.isAdmin(ctx) {
		if uid := s.userID(ctx); uid == uuid.Nil || payment.OwnerID != uid {
			return nil, apperror.Forbidden("not owner", nil)
		}
	}
	updated, err := s.repo.UpdatePayment(ctx, id, req)
	if err != nil {
		return nil, s.mapRepoErr(err)
	}
	return s.toPaymentResponse(updated), nil
}

// UpdatePaymentStatus updates payment status from a payment provider.
func (s *PaymentService) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status models.PaymentStatus) error {
	if err := s.repo.UpdatePaymentStatus(ctx, id, status); err != nil {
		return s.mapRepoErr(err)
	}
	return nil
}

// DeletePayment removes a payment.
func (s *PaymentService) DeletePayment(ctx context.Context, id uuid.UUID) error {
	if !s.isAdmin(ctx) {
		return apperror.Forbidden("only admins can delete payments", nil)
	}
	if err := s.repo.DeletePayment(ctx, id); err != nil {
		return s.mapRepoErr(err)
	}
	return nil
}

