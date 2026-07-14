package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/apperror"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/service"
)

type SubscriptionHandler struct {
	svc *service.PaymentService
}

func NewSubscriptionHandler(svc *service.PaymentService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

func (h *SubscriptionHandler) parseID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, apperror.InvalidRequest("invalid subscription id", err)
	}
	return id, nil
}

// List godoc
// @Summary List subscriptions
// @Tags subscriptions
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param owner_id query string false "Owner ID filter"
// @Param plan_id query string false "Plan ID filter"
// @Param status query string false "Status filter"
// @Param sort_by query string false "Sort field"
// @Param sort_order query string false "Sort order"
// @Success 200 {object} models.SubscriptionListResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseListFilter(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	resp, err := h.svc.ListSubscriptions(r.Context(), filter)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *SubscriptionHandler) parseListFilter(r *http.Request) (models.SubscriptionListFilter, error) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var ownerID uuid.UUID
	if v := q.Get("owner_id"); v != "" {
		var err error
		ownerID, err = uuid.Parse(v)
		if err != nil {
			return models.SubscriptionListFilter{}, apperror.InvalidRequest("invalid owner_id filter", err)
		}
	}
	var planID uuid.UUID
	if v := q.Get("plan_id"); v != "" {
		var err error
		planID, err = uuid.Parse(v)
		if err != nil {
			return models.SubscriptionListFilter{}, apperror.InvalidRequest("invalid plan_id filter", err)
		}
	}

	return models.SubscriptionListFilter{
		OwnerID:   ownerID,
		PlanID:    planID,
		Status:    models.SubscriptionStatus(q.Get("status")),
		SortBy:    q.Get("sort_by"),
		SortOrder: q.Get("sort_order"),
		Limit:     limit,
		Offset:    offset,
	}, nil
}

// Create godoc
// @Summary Create a subscription
// @Tags subscriptions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateSubscriptionRequest true "Create request"
// @Success 201 {object} models.SubscriptionResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSubscriptionRequest
	if err := readJSON(r, &req); err != nil {
		apperror.WriteJSON(w, validationErr(err))
		return
	}
	resp, err := h.svc.CreateSubscription(r.Context(), req)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// GetByID godoc
// @Summary Get a subscription by ID
// @Tags subscriptions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.SubscriptionResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	resp, err := h.svc.GetSubscription(r.Context(), id)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Update godoc
// @Summary Update a subscription
// @Tags subscriptions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param request body models.UpdateSubscriptionRequest true "Update request"
// @Success 200 {object} models.SubscriptionResponse
// @Router /subscriptions/{id} [patch]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	var req models.UpdateSubscriptionRequest
	if err := readJSON(r, &req); err != nil {
		apperror.WriteJSON(w, validationErr(err))
		return
	}
	resp, err := h.svc.UpdateSubscription(r.Context(), id, req)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary Delete a subscription
// @Tags subscriptions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.MessageResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	if err := h.svc.DeleteSubscription(r.Context(), id); err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, models.MessageResponse{Message: "subscription deleted"})
}

// Entitlement godoc
// @Summary Check whether an owner has an active subscription (service only)
// @Tags subscriptions-service
// @Produce json
// @Param owner_id path string true "Owner ID"
// @Success 200 {object} models.EntitlementResponse
// @Router /entitlements/{owner_id} [get]
func (h *SubscriptionHandler) Entitlement(w http.ResponseWriter, r *http.Request) {
	ownerID, err := uuid.Parse(chi.URLParam(r, "owner_id"))
	if err != nil {
		apperror.WriteJSON(w, apperror.InvalidRequest("invalid owner_id", err))
		return
	}
	resp, err := h.svc.GetOwnerEntitlement(r.Context(), ownerID)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Status godoc
// @Summary Get subscription status for service validation
// @Tags subscriptions-service
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.SubscriptionResponse
// @Router /subscriptions/{id}/status [get]
func (h *SubscriptionHandler) Status(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	status, err := h.svc.GetSubscriptionStatus(r.Context(), id)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(status)})
}
