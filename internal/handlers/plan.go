package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/apperror"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/service"
)

type PlanHandler struct {
	svc *service.PaymentService
}

func NewPlanHandler(svc *service.PaymentService) *PlanHandler {
	return &PlanHandler{svc: svc}
}

func (h *PlanHandler) parseID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, apperror.InvalidRequest("invalid plan id", err)
	}
	return id, nil
}

// List godoc
// @Summary List plans
// @Tags plans
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search term"
// @Param is_active query bool false "Active filter"
// @Param interval query string false "Interval filter"
// @Param sort_by query string false "Sort field"
// @Param sort_order query string false "Sort order"
// @Success 200 {object} models.PlanListResponse
// @Router /plans [get]
func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseListFilter(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	resp, err := h.svc.ListPlans(r.Context(), filter)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *PlanHandler) parseListFilter(r *http.Request) (models.PlanListFilter, error) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	isActive, err := parseBoolPtr(q.Get("is_active"))
	if err != nil {
		return models.PlanListFilter{}, apperror.InvalidRequest("invalid is_active filter", err)
	}

	return models.PlanListFilter{
		Search:    strings.TrimSpace(q.Get("search")),
		IsActive:  isActive,
		Interval:  models.PlanInterval(q.Get("interval")),
		SortBy:    q.Get("sort_by"),
		SortOrder: q.Get("sort_order"),
		Limit:     limit,
		Offset:    offset,
	}, nil
}

// Create godoc
// @Summary Create a plan
// @Tags plans
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreatePlanRequest true "Create request"
// @Success 201 {object} models.PlanResponse
// @Router /plans [post]
func (h *PlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePlanRequest
	if err := readJSON(r, &req); err != nil {
		apperror.WriteJSON(w, validationErr(err))
		return
	}
	resp, err := h.svc.CreatePlan(r.Context(), req)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// GetByID godoc
// @Summary Get a plan by ID
// @Tags plans
// @Security BearerAuth
// @Produce json
// @Param id path string true "Plan ID"
// @Success 200 {object} models.PlanResponse
// @Router /plans/{id} [get]
func (h *PlanHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	resp, err := h.svc.GetPlan(r.Context(), id)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Update godoc
// @Summary Update a plan
// @Tags plans
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Plan ID"
// @Param request body models.UpdatePlanRequest true "Update request"
// @Success 200 {object} models.PlanResponse
// @Router /plans/{id} [patch]
func (h *PlanHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	var req models.UpdatePlanRequest
	if err := readJSON(r, &req); err != nil {
		apperror.WriteJSON(w, validationErr(err))
		return
	}
	resp, err := h.svc.UpdatePlan(r.Context(), id, req)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Delete godoc
// @Summary Delete a plan
// @Tags plans
// @Security BearerAuth
// @Produce json
// @Param id path string true "Plan ID"
// @Success 200 {object} models.MessageResponse
// @Router /plans/{id} [delete]
func (h *PlanHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	if err := h.svc.DeletePlan(r.Context(), id); err != nil {
		apperror.WriteJSON(w, err)
		return
	}
	writeJSON(w, http.StatusOK, models.MessageResponse{Message: "plan deleted"})
}
