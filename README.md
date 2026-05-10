# ligo-database

Driver-agnostic database module for [Ligo](https://github.com/linkeunid/ligo), inspired by [@nestjs/typeorm](https://docs.nestjs.com/techniques/database) + TypeORM — but raw SQL, no ORM.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue)](https://go.dev/dl)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-30%20passing-brightgreen)](https://github.com/linkeunid/ligo-database)
[![Coverage](https://img.shields.io/badge/coverage-0%25-yellow)](https://github.com/linkeunid/ligo-database)

## Install

```bash
go get github.com/linkeunid/ligo-database
```

## Quick Start

```go
package main

import (
    database "github.com/linkeunid/ligo-database"
    dbpgx "github.com/linkeunid/ligo-database/pgx"
    "github.com/linkeunid/ligo"
)

func AppModule() ligo.Module {
    dbModule, err := dbpgx.PostgresModule(dbpgx.Config{
        DSN:      "postgres://user:pass@localhost:5432/mydb",
        MaxConns: 25,
        MinConns: 5,
    })
    if err != nil {
        log.Fatal(err)
    }

    return ligo.NewModule("app",
        ligo.Imports(dbModule),
        ligo.Providers(
            ligo.Factory[*UserService](NewUserService),
        ),
    )
}
```

## Three Layers

### Layer 1 — Module Setup

Register a database pool as a global injectable:

```go
dbModule, _ := dbpgx.PostgresModule(dbpgx.Config{
    DSN: "postgres://user:pass@localhost/mydb",
})
```

### Layer 2 — Inject database.DB Directly

Write raw SQL with full control:

```go
type UserService struct {
    db database.DB
}

func NewUserService(db database.DB) *UserService {
    return &UserService{db: db}
}

func (s *UserService) FindByID(ctx context.Context, id int) (*User, error) {
    row := s.db.QueryRow(ctx, "SELECT id, name, email FROM users WHERE id = $1", id)
    var u User
    return &u, row.Scan(&u.ID, &u.Name, &u.Email)
}
```

### Layer 3 — Repository Pattern with ForFeature

Typed repository with automatic transaction awareness:

```go
type UserRepository struct {
    *database.BaseRepository
}

func NewUserRepository(db database.DB) *UserRepository {
    return &UserRepository{BaseRepository: database.NewBaseRepository(db)}
}

func (r *UserRepository) FindAll(ctx context.Context) ([]*User, error) {
    rows, err := r.Query(ctx, "SELECT id, name, email FROM users")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var users []*User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
            return nil, err
        }
        users = append(users, &u)
    }
    return users, rows.Err()
}

func UserModule() ligo.Module {
    return ligo.NewModule("user",
        ligo.Providers(
            database.ForFeature[*UserRepository](NewUserRepository),
            ligo.Factory[*UserService](NewUserService),
        ),
    )
}
```

## Transactions

Context-scoped transactions — repos automatically use the active tx:

```go
func (s *UserService) TransferPoints(ctx context.Context, from, to int, pts int) error {
    return database.RunInTx(ctx, s.db, func(ctx context.Context) error {
        if err := s.userRepo.DeductPoints(ctx, from, pts); err != nil {
            return err // auto-rollback
        }
        return s.userRepo.AddPoints(ctx, to, pts) // auto-commit if nil
    })
}
```

`BaseRepository.DB(ctx)` returns the active transaction from context, or the pool if no tx — repos don't need to change.

## Multi-Database

Named pools for multi-database setups:

```go
app.Register(
    dbpgx.PostgresModule(dbpgx.Config{DSN: "postgres://maindb"}),
    dbpgx.PostgresModuleNamed("analytics", dbpgx.Config{DSN: "postgres://analytics"}),
)

// Inject the registry
type AnalyticsService struct {
    registry *database.DBRegistry
}

func (s *AnalyticsService) DoWork(ctx context.Context) {
    db := s.registry.Get("analytics")
}
```

## Driver-Specific Features

Need pgx-native features? Use `Unwrap()`:

```go
pool := s.db.Unwrap().(*pgxpool.Pool)
rows, _ := pool.Query(ctx, "SELECT * FROM users")
users, _ := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[User])
```

## Health Check

```go
checker := &database.HealthChecker{DB: db}
if err := checker.Check(ctx); err != nil {
    // database unhealthy
}
```

## License

MIT
