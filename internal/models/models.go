package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserRole mirrors the role enum from the users microservice.
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// SubscriptionStatus represents the lifecycle of a subscription.
type SubscriptionStatus string

const (
	SubscriptionActive   SubscriptionStatus = "active"
	SubscriptionCanceled SubscriptionStatus = "canceled"
	SubscriptionExpired  SubscriptionStatus = "expired"
	SubscriptionPastDue  SubscriptionStatus = "past_due"
)

// PaymentStatus represents the status of a payment.
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

// PlanInterval describes the billing interval of a plan.
type PlanInterval string

const (
	PlanIntervalMonthly  PlanInterval = "monthly"
	PlanIntervalYearly   PlanInterval = "yearly"
	PlanIntervalOneTime  PlanInterval = "one_time"
)

// JSONMap is a generic JSON object stored in Postgres jsonb.
type JSONMap map[string]any

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value any) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("invalid json type")
	}
	if len(data) == 0 {
		*j = JSONMap{}
		return nil
	}
	return json.Unmarshal(data, j)
}

// Plan represents a subscription or product plan.
type Plan struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Currency    string    `json:"currency"`
	Interval    PlanInterval `json:"interval"`
	Features    JSONMap   `json:"features"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PlanResponse is the public view of a plan.
type PlanResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	PriceCents  int64     `json:"price_cents"`
	Currency    string    `json:"currency"`
	Interval    PlanInterval `json:"interval"`
	Features    JSONMap   `json:"features"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PlanListResponse is a paginated list of plans.
type PlanListResponse struct {
	Plans  []PlanResponse `json:"plans"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// CreatePlanRequest is the payload for creating a plan.
type CreatePlanRequest struct {
	Name        string    `json:"name" validate:"required,min=2,max=120"`
	Description string    `json:"description" validate:"max=1024"`
	PriceCents  int64     `json:"price_cents" validate:"gte=0"`
	Currency    string    `json:"currency" validate:"required,len=3"`
	Interval    PlanInterval `json:"interval" validate:"required,oneof=monthly yearly one_time"`
	Features    JSONMap   `json:"features"`
	IsActive    bool      `json:"is_active"`
}

// UpdatePlanRequest is the payload for updating a plan.
type UpdatePlanRequest struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=2,max=120"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=1024"`
	PriceCents  *int64     `json:"price_cents,omitempty" validate:"omitempty,gte=0"`
	Currency    *string    `json:"currency,omitempty" validate:"omitempty,len=3"`
	Interval    *PlanInterval `json:"interval,omitempty" validate:"omitempty,oneof=monthly yearly one_time"`
	Features    *JSONMap   `json:"features,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
}

// PlanListFilter is used for listing plans.
type PlanListFilter struct {
	Search    string
	IsActive  *bool
	Interval  PlanInterval
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}

// Subscription represents a user's subscription to a plan.
type Subscription struct {
	ID               uuid.UUID          `json:"id"`
	OwnerID          uuid.UUID          `json:"owner_id"`
	PlanID           uuid.UUID          `json:"plan_id"`
	Status           SubscriptionStatus `json:"status"`
	CurrentPeriodStart time.Time        `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time       `json:"current_period_end"`
	CanceledAt         *time.Time       `json:"canceled_at"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// SubscriptionResponse is the public view of a subscription.
type SubscriptionResponse struct {
	ID               uuid.UUID          `json:"id"`
	OwnerID          uuid.UUID          `json:"owner_id"`
	PlanID           uuid.UUID          `json:"plan_id"`
	Status           SubscriptionStatus `json:"status"`
	CurrentPeriodStart time.Time        `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time       `json:"current_period_end"`
	CanceledAt         *time.Time       `json:"canceled_at"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// SubscriptionListResponse is a paginated list of subscriptions.
type SubscriptionListResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
	Total         int64                  `json:"total"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

// CreateSubscriptionRequest is the payload for creating a subscription.
type CreateSubscriptionRequest struct {
	PlanID uuid.UUID `json:"plan_id" validate:"required"`
}

// UpdateSubscriptionRequest is the payload for updating a subscription.
type UpdateSubscriptionRequest struct {
	Status             *SubscriptionStatus `json:"status,omitempty" validate:"omitempty,oneof=active canceled expired past_due"`
	CurrentPeriodStart *time.Time          `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time          `json:"current_period_end,omitempty"`
	CanceledAt         *time.Time          `json:"canceled_at,omitempty"`
}

// SubscriptionListFilter is used for listing subscriptions.
type SubscriptionListFilter struct {
	OwnerID  uuid.UUID
	PlanID   uuid.UUID
	Status   SubscriptionStatus
	SortBy   string
	SortOrder string
	Limit    int
	Offset   int
}

// Payment represents a payment record.
type Payment struct {
	ID               uuid.UUID     `json:"id"`
	OwnerID          uuid.UUID     `json:"owner_id"`
	SubscriptionID   *uuid.UUID    `json:"subscription_id"`
	AmountCents      int64         `json:"amount_cents"`
	Currency         string        `json:"currency"`
	Status           PaymentStatus `json:"status"`
	Provider         string        `json:"provider"`
	ProviderPaymentID string       `json:"provider_payment_id"`
	Metadata         JSONMap       `json:"metadata"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// PaymentResponse is the public view of a payment.
type PaymentResponse struct {
	ID               uuid.UUID     `json:"id"`
	OwnerID          uuid.UUID     `json:"owner_id"`
	SubscriptionID   *uuid.UUID    `json:"subscription_id"`
	AmountCents      int64         `json:"amount_cents"`
	Currency         string        `json:"currency"`
	Status           PaymentStatus `json:"status"`
	Provider         string        `json:"provider"`
	ProviderPaymentID string       `json:"provider_payment_id"`
	Metadata         JSONMap       `json:"metadata"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// PaymentListResponse is a paginated list of payments.
type PaymentListResponse struct {
	Payments []PaymentResponse `json:"payments"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// CreatePaymentRequest is the payload for creating a payment.
type CreatePaymentRequest struct {
	SubscriptionID   *uuid.UUID `json:"subscription_id,omitempty"`
	AmountCents      int64      `json:"amount_cents" validate:"required,gte=0"`
	Currency         string     `json:"currency" validate:"required,len=3"`
	Provider         string     `json:"provider" validate:"required,max=64"`
	ProviderPaymentID string    `json:"provider_payment_id" validate:"max=255"`
	Metadata         JSONMap    `json:"metadata"`
}

// UpdatePaymentRequest is the payload for updating a payment.
type UpdatePaymentRequest struct {
	Status            *PaymentStatus `json:"status,omitempty" validate:"omitempty,oneof=pending completed failed refunded"`
	ProviderPaymentID *string        `json:"provider_payment_id,omitempty" validate:"omitempty,max=255"`
	Metadata          *JSONMap       `json:"metadata,omitempty"`
}

// UpdatePaymentStatusRequest is the payload for a service updating payment status.
type UpdatePaymentStatusRequest struct {
	Status PaymentStatus `json:"status" validate:"required,oneof=pending completed failed refunded"`
}

// PaymentListFilter is used for listing payments.
type PaymentListFilter struct {
	OwnerID        uuid.UUID
	SubscriptionID uuid.UUID
	Status         PaymentStatus
	Provider       string
	SortBy         string
	SortOrder      string
	Limit          int
	Offset         int
}

// MessageResponse is a generic message response.
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse is a generic error response.
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// ClaimsContextKey is the context key for JWT claims.
type ClaimsContextKey struct{}
