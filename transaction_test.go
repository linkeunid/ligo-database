package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTx struct {
	committed  bool
	rolledBack bool
}

func (t *mockTx) Exec(context.Context, string, ...any) (Result, error) { return nil, nil }
func (t *mockTx) Query(context.Context, string, ...any) (Rows, error)  { return nil, nil }
func (t *mockTx) QueryRow(context.Context, string, ...any) Row         { return nil }
func (t *mockTx) Commit(context.Context) error { t.committed = true; return nil }

func (t *mockTx) Rollback(context.Context) error { t.rolledBack = true; return nil }

type mockBeginDB struct {
	mockDB
	tx *mockTx
}

func (m *mockBeginDB) Begin(ctx context.Context) (Tx, error) {
	return m.tx, nil
}

func TestRunInTx_CommitsOnSuccess(t *testing.T) {
	tx := &mockTx{}
	db := &mockBeginDB{tx: tx}
	called := false

	err := RunInTx(context.Background(), db, func(ctx context.Context) error {
		called = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, called)
	assert.True(t, tx.committed)
	assert.False(t, tx.rolledBack)
}

func TestRunInTx_RollbacksOnError(t *testing.T) {
	tx := &mockTx{}
	db := &mockBeginDB{tx: tx}

	err := RunInTx(context.Background(), db, func(ctx context.Context) error {
		return errors.New("boom")
	})

	assert.Error(t, err)
	assert.True(t, tx.rolledBack)
	assert.False(t, tx.committed)
}

func TestWithTxAndTxFromCtx(t *testing.T) {
	tx := &mockTx{}
	ctx := WithTx(context.Background(), tx)

	got, ok := TxFromCtx(ctx)
	assert.True(t, ok)
	assert.Equal(t, tx, got)
}

func TestTxFromCtx_ReturnsFalseWhenNoTx(t *testing.T) {
	_, ok := TxFromCtx(context.Background())
	assert.False(t, ok)
}

func TestRunInTx_PropagatesTxInContext(t *testing.T) {
	tx := &mockTx{}
	db := &mockBeginDB{tx: tx}

	err := RunInTx(context.Background(), db, func(ctx context.Context) error {
		got, ok := TxFromCtx(ctx)
		assert.True(t, ok)
		assert.Equal(t, tx, got)
		return nil
	})

	assert.NoError(t, err)
}
