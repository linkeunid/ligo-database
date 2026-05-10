package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBRegistry_Get(t *testing.T) {
	db1 := &dbStub{}
	db2 := &dbStub{}
	r := NewDBRegistry()
	r.Register("primary", db1)
	r.Register("analytics", db2)

	assert.Equal(t, db1, r.Get("primary"))
	assert.Equal(t, db2, r.Get("analytics"))
}

func TestDBRegistry_Get_PanicsOnUnknown(t *testing.T) {
	r := NewDBRegistry()
	assert.Panics(t, func() { r.Get("unknown") })
}

func TestDBRegistry_Default(t *testing.T) {
	db := &dbStub{}
	r := NewDBRegistry()
	r.RegisterDefault(db)

	assert.Equal(t, db, r.Default())
}

func TestDBRegistry_Default_PanicsWhenNotSet(t *testing.T) {
	r := NewDBRegistry()
	assert.Panics(t, func() { r.Default() })
}
