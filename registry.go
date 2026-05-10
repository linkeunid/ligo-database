package database

import (
	"fmt"
	"sync"
)

type DBRegistry struct {
	mu        sync.RWMutex
	dbs       map[string]DB
	defaultDB DB
}

func NewDBRegistry() *DBRegistry {
	return &DBRegistry{dbs: make(map[string]DB)}
}

func (r *DBRegistry) Register(name string, db DB) {
	r.mu.Lock()
	r.dbs[name] = db
	r.mu.Unlock()
}

func (r *DBRegistry) RegisterDefault(db DB) {
	r.mu.Lock()
	r.defaultDB = db
	r.mu.Unlock()
}

func (r *DBRegistry) Get(name string) DB {
	r.mu.RLock()
	db, ok := r.dbs[name]
	r.mu.RUnlock()
	if !ok {
		panic(fmt.Sprintf("database: no pool registered with name %q", name))
	}
	return db
}

func (r *DBRegistry) Default() DB {
	r.mu.RLock()
	db := r.defaultDB
	r.mu.RUnlock()
	if db == nil {
		panic("database: no default pool registered")
	}
	return db
}
