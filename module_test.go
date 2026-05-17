package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockDB struct {
	pinged bool
	closed bool
}

func (m *mockDB) Exec(context.Context, string, ...any) (Result, error) { return nil, nil }
func (m *mockDB) Query(context.Context, string, ...any) (Rows, error)  { return nil, nil }
func (m *mockDB) QueryRow(context.Context, string, ...any) Row         { return nil }
func (m *mockDB) Ping(context.Context) error                           { m.pinged = true; return nil }

func (m *mockDB) Close() error                      { m.closed = true; return nil }
func (m *mockDB) Begin(context.Context) (Tx, error) { return nil, nil }
func (m *mockDB) Unwrap() any                       { return nil }

func TestModuleName(t *testing.T) {
	m := Module(&mockDB{})
	assert.Equal(t, "database", m.Name)
}

func TestModuleRegistersDBProvider(t *testing.T) {
	db := &mockDB{}
	m := Module(db)
	assert.NotEmpty(t, m.Providers)
}

func TestModuleRegistersRegistry(t *testing.T) {
	db := &mockDB{}
	m := Module(db)
	assert.GreaterOrEqual(t, len(m.Providers), 2)
}

func TestModuleNamed_HasQualifiedName(t *testing.T) {
	m := ModuleNamed("analytics", &mockDB{})
	assert.Equal(t, "database:analytics", m.Name)
}
