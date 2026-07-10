package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/teaspeak-v2/wt-bot-ms-payments-v1/internal/models"
)

type db interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	db db
}

func New(db db) *Repository {
	return &Repository{db: db}
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

type scanner interface {
	Scan(...any) error
}

func coalesceStr(v *string, def string) string {
	if v != nil {
		return *v
	}
	return def
}

func coalesceBool(v *bool, def bool) bool {
	if v != nil {
		return *v
	}
	return def
}

func coalesceInt64(v *int64, def int64) int64 {
	if v != nil {
		return *v
	}
	return def
}

func coalesceJSONMap(v *models.JSONMap, def models.JSONMap) models.JSONMap {
	if v != nil {
		return *v
	}
	return def
}

func sanitizeSort(sortBy, sortOrder string, whitelist map[string]string, defaultSort string) string {
	col, ok := whitelist[strings.ToLower(sortBy)]
	if !ok {
		col = defaultSort
	}
	order := "desc"
	if strings.EqualFold(sortOrder, "asc") {
		order = "asc"
	}
	return col + " " + order
}

func nextArg(args []any) int {
	return len(args) + 1
}
