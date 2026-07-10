package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
)

// PaymentRepository defines persistence for Payment records.
type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Payment, error)
	List(ctx context.Context, filter models.PaymentListFilter) ([]models.Payment, int64, error)
	Update(ctx context.Context, id uuid.UUID, req models.UpdatePaymentRequest) (*models.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.PaymentStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}

func (r *Repository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	if payment.ID == uuid.Nil {
		payment.ID = uuid.New()
	}
	now := time.Now().UTC()
	payment.CreatedAt = now
	payment.UpdatedAt = now

	const q = `insert into payments (
		id, owner_id, subscription_id, amount_cents, currency, status, provider, provider_payment_id, metadata, created_at, updated_at
	) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.db.Exec(ctx, q,
		payment.ID, payment.OwnerID, payment.SubscriptionID, payment.AmountCents, payment.Currency, payment.Status, payment.Provider, payment.ProviderPaymentID, payment.Metadata, payment.CreatedAt, payment.UpdatedAt,
	)
	return err
}

func (r *Repository) GetPaymentByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	q := fmt.Sprintf(`select %s from payments p where p.id=$1`, paymentColumns)
	row := r.db.QueryRow(ctx, q, id)
	return scanPayment(row)
}

func (r *Repository) ListPayments(ctx context.Context, filter models.PaymentListFilter) ([]models.Payment, int64, error) {
	where, args := r.buildPaymentWhere(filter)
	sort := sanitizeSort(filter.SortBy, filter.SortOrder, paymentSortWhitelist, "created_at")
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	listQ := fmt.Sprintf(`select %s from payments p %s order by %s limit $%d offset $%d`, paymentColumns, where, sort, nextArg(args), nextArg(args)+1)
	allArgs := append(args, limit, offset)

	rows, err := r.db.Query(ctx, listQ, allArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []models.Payment
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *payment)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	countQ := fmt.Sprintf(`select count(*) from payments p %s`, where)
	var count int64
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&count); err != nil {
		return nil, 0, err
	}
	return list, count, nil
}

func (r *Repository) buildPaymentWhere(filter models.PaymentListFilter) (string, []any) {
	var conds []string
	var args []any
	if filter.OwnerID != uuid.Nil {
		conds = append(conds, fmt.Sprintf("p.owner_id = $%d", nextArg(args)))
		args = append(args, filter.OwnerID)
	}
	if filter.SubscriptionID != uuid.Nil {
		conds = append(conds, fmt.Sprintf("p.subscription_id = $%d", nextArg(args)))
		args = append(args, filter.SubscriptionID)
	}
	if filter.Status != "" {
		conds = append(conds, fmt.Sprintf("p.status = $%d", nextArg(args)))
		args = append(args, filter.Status)
	}
	if filter.Provider != "" {
		conds = append(conds, fmt.Sprintf("p.provider = $%d", nextArg(args)))
		args = append(args, filter.Provider)
	}
	if len(conds) == 0 {
		return "", args
	}
	return "where " + strings.Join(conds, " and "), args
}

func (r *Repository) UpdatePayment(ctx context.Context, id uuid.UUID, req models.UpdatePaymentRequest) (*models.Payment, error) {
	payment, err := r.GetPaymentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	status := coalesceStr((*string)(req.Status), string(payment.Status))
	providerPaymentID := coalesceStr(req.ProviderPaymentID, payment.ProviderPaymentID)
	metadata := coalesceJSONMap(req.Metadata, payment.Metadata)

	const q = `update payments set status=$2, provider_payment_id=$3, metadata=$4, updated_at=now()
		where id=$1
		returning ` + paymentColumns
	row := r.db.QueryRow(ctx, q,
		id, status, providerPaymentID, metadata,
	)
	return scanPayment(row)
}

func (r *Repository) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status models.PaymentStatus) error {
	_, err := r.db.Exec(ctx, `update payments set status=$2, updated_at=now() where id=$1`, id, status)
	return err
}

func (r *Repository) DeletePayment(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `delete from payments where id=$1`, id)
	return err
}

const paymentColumns = `p.id, p.owner_id, p.subscription_id, p.amount_cents, p.currency, p.status, p.provider, p.provider_payment_id, p.metadata, p.created_at, p.updated_at`

var paymentSortWhitelist = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"status":     "status",
	"provider":   "provider",
}

func scanPayment(row scanner) (*models.Payment, error) {
	var payment models.Payment
	if err := row.Scan(
		&payment.ID, &payment.OwnerID, &payment.SubscriptionID, &payment.AmountCents, &payment.Currency, &payment.Status, &payment.Provider, &payment.ProviderPaymentID, &payment.Metadata, &payment.CreatedAt, &payment.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &payment, nil
}
