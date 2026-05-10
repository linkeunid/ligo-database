package database

import (
	"context"

	"github.com/linkeunid/ligo"
)

type BaseRepository struct {
	db DB
}

func NewBaseRepository(db DB) *BaseRepository {
	return &BaseRepository{db: db}
}

func (r *BaseRepository) DB(ctx context.Context) Querier {
	if tx, ok := TxFromCtx(ctx); ok {
		return tx
	}
	return r.db
}

func (r *BaseRepository) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	return r.DB(ctx).Exec(ctx, query, args...)
}

func (r *BaseRepository) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	return r.DB(ctx).Query(ctx, query, args...)
}

func (r *BaseRepository) QueryRow(ctx context.Context, query string, args ...any) Row {
	return r.DB(ctx).QueryRow(ctx, query, args...)
}

func ForFeature[R any](fn func(DB) R, opts ...ligo.ProviderOption) ligo.Provider {
	return ligo.Factory[R](fn, opts...)
}
