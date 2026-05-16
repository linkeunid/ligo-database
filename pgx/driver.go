package pgx

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	database "github.com/linkeunid/ligo-database"
	"github.com/linkeunid/ligo"
)

// --- Config ---

type Config struct {
	DSN               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// connectTimeout bounds the pre-flight ping so a wrong host doesn't hang.
const connectTimeout = 5 * time.Second

func newPool(cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("database/pgx: invalid DSN (check scheme, host, port, user, password, dbname, sslmode): %w", err)
	}
	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolCfg.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.HealthCheckPeriod > 0 {
		poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database/pgx: create pool: %w", err)
	}

	// Pre-flight ping: surface auth / connectivity errors at construction
	// time with a clear message, instead of being swallowed by a lifecycle
	// hook later as a generic "hook execution failed".
	pingCtx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf(
			"database/pgx: cannot connect to %s (check credentials, host/port, and that the database exists): %w",
			redactDSN(poolCfg.ConnConfig), err,
		)
	}
	return pool, nil
}

// redactDSN renders a safe, log-friendly form of the connection target with
// the password masked.
func redactDSN(c *pgx.ConnConfig) string {
	if c == nil {
		return "<unknown>"
	}
	db := c.Database
	if db == "" {
		db = "?"
	}
	return fmt.Sprintf("postgres://%s:***@%s:%d/%s", c.User, c.Host, c.Port, db)
}

// --- Module constructors ---

func PostgresModule(cfg Config) (ligo.Module, error) {
	w, err := wrapPool(cfg)
	if err != nil {
		return ligo.Module{}, err
	}
	return database.Module(w), nil
}

func PostgresModuleNamed(name string, cfg Config) (ligo.Module, error) {
	w, err := wrapPool(cfg)
	if err != nil {
		return ligo.Module{}, err
	}
	return database.ModuleNamed(name, w), nil
}

func wrapPool(cfg Config) (*poolWrapper, error) {
	pool, err := newPool(cfg)
	if err != nil {
		return nil, err
	}
	return &poolWrapper{querierAdapter: querierAdapter{querier: pool}, pool: pool}, nil
}

// --- Shared querier adapter ---

type pgxQuerier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type querierAdapter struct {
	querier pgxQuerier
}

func (a querierAdapter) Exec(ctx context.Context, query string, args ...any) (database.Result, error) {
	tag, err := a.querier.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &resultAdapter{tag: tag}, nil
}

func (a querierAdapter) Query(ctx context.Context, query string, args ...any) (database.Rows, error) {
	rows, err := a.querier.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &rowsAdapter{rows: rows}, nil
}

func (a querierAdapter) QueryRow(ctx context.Context, query string, args ...any) database.Row {
	row := a.querier.QueryRow(ctx, query, args...)
	return &rowAdapter{row: row}
}

// --- Pool wrapper (database.DB) ---

type poolWrapper struct {
	querierAdapter
	pool *pgxpool.Pool
}

func (w *poolWrapper) Ping(ctx context.Context) error { return w.pool.Ping(ctx) }
func (w *poolWrapper) Close() error                   { w.pool.Close(); return nil }
func (w *poolWrapper) Unwrap() any                    { return w.pool }

func (w *poolWrapper) Begin(ctx context.Context) (database.Tx, error) {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &txAdapter{querierAdapter: querierAdapter{querier: tx}, tx: tx}, nil
}

// --- Tx wrapper (database.Tx) ---

type txAdapter struct {
	querierAdapter
	tx pgx.Tx
}

func (t *txAdapter) Commit(ctx context.Context) error   { return t.tx.Commit(ctx) }
func (t *txAdapter) Rollback(ctx context.Context) error { return t.tx.Rollback(ctx) }

// --- Adapters ---

type resultAdapter struct {
	tag pgconn.CommandTag
}

func (r *resultAdapter) RowsAffected() (int64, error) { return r.tag.RowsAffected(), nil }

type rowsAdapter struct {
	rows pgx.Rows
}

func (r *rowsAdapter) Next() bool            { return r.rows.Next() }
func (r *rowsAdapter) Scan(dest ...any) error { return r.rows.Scan(dest...) }
func (r *rowsAdapter) Close()                { r.rows.Close() }
func (r *rowsAdapter) Err() error            { return r.rows.Err() }

type rowAdapter struct {
	row pgx.Row
}

func (r *rowAdapter) Scan(dest ...any) error { return r.row.Scan(dest...) }
