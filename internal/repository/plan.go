package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
)

// PlanRepository defines persistence for Plan records.
type PlanRepository interface {
	Create(ctx context.Context, plan *models.Plan) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Plan, error)
	List(ctx context.Context, filter models.PlanListFilter) ([]models.Plan, int64, error)
	Update(ctx context.Context, id uuid.UUID, req models.UpdatePlanRequest) (*models.Plan, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

func (r *Repository) CreatePlan(ctx context.Context, plan *models.Plan) error {
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	now := time.Now().UTC()
	plan.CreatedAt = now
	plan.UpdatedAt = now

	const q = `insert into plans (
		id, name, description, price_cents, currency, interval, features, is_active, created_at, updated_at
	) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.Exec(ctx, q,
		plan.ID, plan.Name, plan.Description, plan.PriceCents, plan.Currency, plan.Interval, plan.Features, plan.IsActive, plan.CreatedAt, plan.UpdatedAt,
	)
	return err
}

func (r *Repository) GetPlanByID(ctx context.Context, id uuid.UUID) (*models.Plan, error) {
	q := fmt.Sprintf(`select %s from plans p where p.id=$1`, planColumns)
	row := r.db.QueryRow(ctx, q, id)
	return scanPlan(row)
}

func (r *Repository) ListPlans(ctx context.Context, filter models.PlanListFilter) ([]models.Plan, int64, error) {
	where, args := r.buildPlanWhere(filter)
	sort := sanitizeSort(filter.SortBy, filter.SortOrder, planSortWhitelist, "created_at")
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	listQ := fmt.Sprintf(`select %s from plans p %s order by %s limit $%d offset $%d`, planColumns, where, sort, nextArg(args), nextArg(args)+1)
	allArgs := append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQ, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []models.Plan
	for rows.Next() {
		plan, err := scanPlan(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *plan)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countQ := fmt.Sprintf(`select count(*) from plans p %s`, where)
	var count int64
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&count); err != nil {
		return nil, 0, err
	}
	return list, count, nil
}

func (r *Repository) buildPlanWhere(filter models.PlanListFilter) (string, []any) {
	var conds []string
	var args []any
	if filter.Search != "" {
		idx := nextArg(args)
		conds = append(conds, fmt.Sprintf(`(p.name ilike $%d or p.description ilike $%d)`, idx, idx))
		args = append(args, "%"+filter.Search+"%")
	}
	if filter.IsActive != nil {
		conds = append(conds, fmt.Sprintf("p.is_active = $%d", nextArg(args)))
		args = append(args, *filter.IsActive)
	}
	if filter.Interval != "" {
		conds = append(conds, fmt.Sprintf("p.interval = $%d", nextArg(args)))
		args = append(args, filter.Interval)
	}
	if len(conds) == 0 {
		return "", args
	}
	return "where " + strings.Join(conds, " and "), args
}

func (r *Repository) UpdatePlan(ctx context.Context, id uuid.UUID, req models.UpdatePlanRequest) (*models.Plan, error) {
	plan, err := r.GetPlanByID(ctx, id)
	if err != nil {
		return nil, err
	}
	name := coalesceStr(req.Name, plan.Name)
	description := coalesceStr(req.Description, plan.Description)
	priceCents := coalesceInt64(req.PriceCents, plan.PriceCents)
	currency := coalesceStr(req.Currency, plan.Currency)
	interval := coalesceStr((*string)(req.Interval), string(plan.Interval))
	features := coalesceJSONMap(req.Features, plan.Features)
	isActive := coalesceBool(req.IsActive, plan.IsActive)

	const q = `update plans set name=$2, description=$3, price_cents=$4, currency=$5, interval=$6, features=$7, is_active=$8, updated_at=now()
		where id=$1
		returning ` + planColumns
	row := r.db.QueryRow(ctx, q,
		id, name, description, priceCents, currency, interval, features, isActive,
	)
	return scanPlan(row)
}

func (r *Repository) DeletePlan(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `delete from plans where id=$1`, id)
	return err
}

const planColumns = `p.id, p.name, p.description, p.price_cents, p.currency, p.interval, p.features, p.is_active, p.created_at, p.updated_at`

var planSortWhitelist = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"name":       "name",
	"price_cents": "price_cents",
	"is_active":  "is_active",
}

func scanPlan(row scanner) (*models.Plan, error) {
	var plan models.Plan
	if err := row.Scan(
		&plan.ID, &plan.Name, &plan.Description, &plan.PriceCents, &plan.Currency, &plan.Interval, &plan.Features, &plan.IsActive, &plan.CreatedAt, &plan.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &plan, nil
}
