package database

import (
	"context"
	"sync"

	"github.com/linkeunid/ligo"
)

var (
	registryOnce   sync.Once
	globalRegistry *DBRegistry
)

func getRegistry() *DBRegistry {
	registryOnce.Do(func() {
		globalRegistry = NewDBRegistry()
	})
	return globalRegistry
}

func Module(db DB) ligo.Module {
	return newModule("database", db, func(r *DBRegistry) { r.RegisterDefault(db) })
}

func ModuleNamed(name string, db DB) ligo.Module {
	return newModule("database:"+name, db, func(r *DBRegistry) { r.Register(name, db) })
}

func newModule(name string, db DB, regFunc func(*DBRegistry)) ligo.Module {
	reg := getRegistry()
	regFunc(reg)

	return ligo.NewModule(
		name,
		ligo.Providers(
			ligo.Export(
				ligo.Value(
					db,
					ligo.WithHooks(
						ligo.OnInit(func() error {
							return db.Ping(context.Background())
						}),
						ligo.OnDestroy(func() error {
							return db.Close()
						}),
					),
				),
			),
			ligo.Export(
				ligo.Value(reg),
			),
		),
	)
}
