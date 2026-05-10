package database

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockQuerier struct {
	mockDB
	execCalled     bool
	queryCalled    bool
	queryRowCalled bool
}

func (m *mockQuerier) Exec(context.Context, string, ...any) (Result, error) {
	m.execCalled = true
	return nil, nil
}

func (m *mockQuerier) Query(context.Context, string, ...any) (Rows, error) {
	m.queryCalled = true
	return nil, nil
}

func (m *mockQuerier) QueryRow(context.Context, string, ...any) Row {
	m.queryRowCalled = true
	return nil
}

func TestBaseRepository_DB_ReturnsPoolWhenNoTx(t *testing.T) {
	db := &mockQuerier{}
	repo := NewBaseRepository(db)
	got := repo.DB(context.Background())
	assert.Equal(t, db, got)
}

func TestBaseRepository_DB_ReturnsTxWhenInContext(t *testing.T) {
	db := &mockQuerier{}
	tx := &mockTx{}
	ctx := WithTx(context.Background(), tx)
	repo := NewBaseRepository(db)
	got := repo.DB(ctx)
	assert.Equal(t, tx, got)
}

func TestBaseRepository_Exec(t *testing.T) {
	q := &mockQuerier{}
	repo := NewBaseRepository(q)
	repo.Exec(context.Background(), "DELETE FROM users WHERE id = $1", 1)
	assert.True(t, q.execCalled)
}

func TestBaseRepository_Query(t *testing.T) {
	q := &mockQuerier{}
	repo := NewBaseRepository(q)
	repo.Query(context.Background(), "SELECT * FROM users")
	assert.True(t, q.queryCalled)
}

func TestBaseRepository_QueryRow(t *testing.T) {
	q := &mockQuerier{}
	repo := NewBaseRepository(q)
	repo.QueryRow(context.Background(), "SELECT 1")
	assert.True(t, q.queryRowCalled)
}

func TestForFeature_ReturnsLigoProvider(t *testing.T) {
	p := ForFeature(func(db DB) *mockQuerier {
		return &mockQuerier{}
	})
	assert.NotNil(t, p)
	assert.Equal(t, reflect.TypeOf(&mockQuerier{}), p.Type())
}

func TestForFeature_ReturnsFactoryType(t *testing.T) {
	p := ForFeature(func(db DB) *mockQuerier {
		return &mockQuerier{}
	})
	assert.False(t, p.IsTransient())
	assert.Nil(t, p.Eager())
}
