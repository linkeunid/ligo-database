package database

import (
	"context"
	"testing"
)

func TestQuerierInterface(t *testing.T) {
	var _ Querier = DB(nil)
	var _ Querier = Tx(nil)
}

func TestDBInterface(t *testing.T) {
	var _ DB = (*dbStub)(nil)
}

func TestTxInterface(t *testing.T) {
	var _ Tx = (*txStub)(nil)
}

type dbStub struct{}

func (d *dbStub) Exec(context.Context, string, ...any) (Result, error) { return nil, nil }
func (d *dbStub) Query(context.Context, string, ...any) (Rows, error)  { return nil, nil }
func (d *dbStub) QueryRow(context.Context, string, ...any) Row         { return nil }
func (d *dbStub) Ping(context.Context) error                           { return nil }
func (d *dbStub) Close() error                                         { return nil }
func (d *dbStub) Begin(context.Context) (Tx, error)                    { return nil, nil }
func (d *dbStub) Unwrap() any                                          { return nil }

type txStub struct{}

func (t *txStub) Exec(context.Context, string, ...any) (Result, error) { return nil, nil }
func (t *txStub) Query(context.Context, string, ...any) (Rows, error)  { return nil, nil }
func (t *txStub) QueryRow(context.Context, string, ...any) Row         { return nil }
func (t *txStub) Commit(context.Context) error                         { return nil }
func (t *txStub) Rollback(context.Context) error                       { return nil }
