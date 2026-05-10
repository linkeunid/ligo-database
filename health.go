package database

import "context"

type HealthChecker struct {
	DB DB
}

func (h *HealthChecker) Check(ctx context.Context) error {
	return h.DB.Ping(ctx)
}
