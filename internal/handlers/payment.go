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

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) parseID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, apperror.InvalidRequest("invalid payment id", err)
	}
	return id, nil
}

// List godoc
// @Summary List payments
// @Tags payments
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param subscription_id query string false "Subscription ID filter"
// @Param status query string false "Status filter"
// @Param provider query string false "Provider filter"
// @Param sort_by query string false "Sort field"
// @Param sort_order query string false "Sort order"
// @Success 200 {object} models.PaymentListResponse
// @Router /payments [get]
func (h *PaymentHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseListFilter(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	resp, err := h.svc.ListPayments(r.Context(), filter)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *PaymentHandler) parseListFilter(r *http.Request) (models.PaymentListFilter, error) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var subscriptionID uuid.UUID
	if v := q.Get("subscription_id"); v != "" {
		var err error
		subscriptionID, err = uuid.Parse(v)
		if err != nil {
			return models.PaymentListFilter{}, apperror.InvalidRequest("invalid subscription_id filter", err)
		}
	}

	return models.PaymentListFilter{
		SubscriptionID: subscriptionID,
		Status:         models.PaymentStatus(q.Get("status")),
		Provider:       q.Get("provider"),
		SortBy:         q.Get("sort_by"),
		SortOrder:      q.Get("sort_order"),
		Limit:          limit,
		Offset:         offset,
	}, nil
}

// Create godoc
// @Summary Create a payment
// @Tags payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreatePaymentRequest true "Create request"
// @Success 201 {object} models.PaymentResponse
// @Router /payments [post]
func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePaymentRequest
	if err := readJSON(r, &req); err != nil {
		apperror.WriteJSON(w, validationErr(err))
		return
	}
	resp, err := h.svc.CreatePayment(r.Context(), req)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// GetByID godoc
// @Summary Get a payment by ID
// @Tags payments
// @Security BearerAuth
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} models.PaymentResponse
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	resp, err := h.svc.GetPayment(r.Context(), id)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Update godoc
// @Summary Update a payment
// @Tags payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param request body models.UpdatePaymentRequest true "Update request"
// @Success 200 {object} models.PaymentResponse
// @Router /payments/{id} [patch]
func (h *PaymentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	var req models.UpdatePaymentRequest
	if err := readJSON(r, &req); err != nil {
		apperror.WriteJSON(w, validationErr(err))
		return
	}
	resp, err := h.svc.UpdatePayment(r.Context(), id, req)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary Delete a payment
// @Tags payments
// @Security BearerAuth
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} models.MessageResponse
// @Router /payments/{id} [delete]
func (h *PaymentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	if err := h.svc.DeletePayment(r.Context(), id); err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, models.MessageResponse{Message: "payment deleted"})
}

// UpdateStatus godoc
// @Summary Update payment status from a payment provider
// @Tags payments-service
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param request body models.UpdatePaymentStatusRequest true "Status request"
// @Success 200 {object} models.MessageResponse
// @Router /payments/{id}/status [post]
func (h *PaymentHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	var req models.UpdatePaymentStatusRequest
	if err := readJSON(r, &req); err != nil {
		apperror.WriteJSON(w, validationErr(err))
		return
	}
	if err := h.svc.UpdatePaymentStatus(r.Context(), id, req.Status); err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, models.MessageResponse{Message: "payment status updated"})
}
