package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type pingErrorDB struct {
	mockDB
	err error
}

func (m *pingErrorDB) Ping(context.Context) error { return m.err }

func TestHealthChecker_Check_Success(t *testing.T) {
	db := &pingErrorDB{}
	h := &HealthChecker{DB: db}
	assert.NoError(t, h.Check(context.Background()))
}

func TestHealthChecker_Check_Error(t *testing.T) {
	db := &pingErrorDB{err: errors.New("connection refused")}
	h := &HealthChecker{DB: db}
	assert.Error(t, h.Check(context.Background()))
}
