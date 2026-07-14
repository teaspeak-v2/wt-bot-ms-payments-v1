package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
)

// SubscriptionRepository defines persistence for Subscription records.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	List(ctx context.Context, filter models.SubscriptionListFilter) ([]models.Subscription, int64, error)
	Update(ctx context.Context, id uuid.UUID, req models.UpdateSubscriptionRequest) (*models.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetStatusByID(ctx context.Context, id uuid.UUID) (models.SubscriptionStatus, error)
}

func (r *Repository) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	now := time.Now().UTC()
	sub.CreatedAt = now
	sub.UpdatedAt = now

	const q = `insert into subscriptions (
		id, owner_id, plan_id, status, current_period_start, current_period_end, canceled_at, created_at, updated_at
	) values ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.db.Exec(ctx, q,
		sub.ID, sub.OwnerID, sub.PlanID, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.CanceledAt, sub.CreatedAt, sub.UpdatedAt,
	)
	return err
}

func (r *Repository) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	q := fmt.Sprintf(`select %s from subscriptions s where s.id=$1`, subscriptionColumns)
	row := r.db.QueryRow(ctx, q, id)
	return scanSubscription(row)
}

// GetActiveSubscriptionByOwner returns the owner's current active subscription,
// if any. A subscription counts as active when its status is "active" and it has
// not passed its current period end. Returns pgx.ErrNoRows when none exists.
func (r *Repository) GetActiveSubscriptionByOwner(ctx context.Context, ownerID uuid.UUID) (*models.Subscription, error) {
	q := fmt.Sprintf(`select %s from subscriptions s
		where s.owner_id=$1 and s.status=$2
		and (s.current_period_end is null or s.current_period_end > now())
		order by s.current_period_end desc nulls first, s.created_at desc
		limit 1`, subscriptionColumns)
	row := r.db.QueryRow(ctx, q, ownerID, models.SubscriptionActive)
	return scanSubscription(row)
}

func (r *Repository) GetSubscriptionStatusByID(ctx context.Context, id uuid.UUID) (models.SubscriptionStatus, error) {
	var status string
	err := r.db.QueryRow(ctx, `select status from subscriptions where id=$1`, id).Scan(&status)
	if err != nil {
		return "", err
	}
	return models.SubscriptionStatus(status), nil
}

func (r *Repository) ListSubscriptions(ctx context.Context, filter models.SubscriptionListFilter) ([]models.Subscription, int64, error) {
	where, args := r.buildSubscriptionWhere(filter)
	sort := sanitizeSort(filter.SortBy, filter.SortOrder, subscriptionSortWhitelist, "created_at")
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	listQ := fmt.Sprintf(`select %s from subscriptions s %s order by %s limit $%d offset $%d`, subscriptionColumns, where, sort, nextArg(args), nextArg(args)+1)
	allArgs := append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQ, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []models.Subscription
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *sub)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countQ := fmt.Sprintf(`select count(*) from subscriptions s %s`, where)
	var count int64
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&count); err != nil {
		return nil, 0, err
	}
	return list, count, nil
}

func (r *Repository) buildSubscriptionWhere(filter models.SubscriptionListFilter) (string, []any) {
	var conds []string
	var args []any
	if filter.OwnerID != uuid.Nil {
		conds = append(conds, fmt.Sprintf("s.owner_id = $%d", nextArg(args)))
		args = append(args, filter.OwnerID)
	}
	if filter.PlanID != uuid.Nil {
		conds = append(conds, fmt.Sprintf("s.plan_id = $%d", nextArg(args)))
		args = append(args, filter.PlanID)
	}
	if filter.Status != "" {
		conds = append(conds, fmt.Sprintf("s.status = $%d", nextArg(args)))
		args = append(args, filter.Status)
	}
	if len(conds) == 0 {
		return "", args
	}
	return "where " + strings.Join(conds, " and "), args
}

func (r *Repository) UpdateSubscription(ctx context.Context, id uuid.UUID, req models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	sub, err := r.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	status := coalesceStr((*string)(req.Status), string(sub.Status))
	currentPeriodStart := coalesceTime(req.CurrentPeriodStart, sub.CurrentPeriodStart)
	currentPeriodEnd := coalesceTimePtr(sub.CurrentPeriodEnd, req.CurrentPeriodEnd)
	canceledAt := coalesceTimePtr(sub.CanceledAt, req.CanceledAt)

	const q = `update subscriptions set status=$2, current_period_start=$3, current_period_end=$4, canceled_at=$5, updated_at=now()
		where id=$1
		returning ` + subscriptionColumns
	row := r.db.QueryRow(ctx, q,
		id, status, currentPeriodStart, currentPeriodEnd, canceledAt,
	)
	return scanSubscription(row)
}

func (r *Repository) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `delete from subscriptions where id=$1`, id)
	return err
}

func coalesceTime(v *time.Time, def time.Time) time.Time {
	if v != nil {
		return *v
	}
	return def
}

func coalesceTimePtr(def *time.Time, v *time.Time) *time.Time {
	if v != nil {
		return v
	}
	return def
}

const subscriptionColumns = `s.id, s.owner_id, s.plan_id, s.status, s.current_period_start, s.current_period_end, s.canceled_at, s.created_at, s.updated_at`

var subscriptionSortWhitelist = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"status":     "status",
	"plan_id":    "plan_id",
}

func scanSubscription(row scanner) (*models.Subscription, error) {
	var sub models.Subscription
	if err := row.Scan(
		&sub.ID, &sub.OwnerID, &sub.PlanID, &sub.Status, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CanceledAt, &sub.CreatedAt, &sub.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &sub, nil
}
