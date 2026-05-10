package pgx

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	database "github.com/linkeunid/ligo-database"
	"github.com/stretchr/testify/assert"
)

func TestPoolWrapper_ImplementsDB(t *testing.T) {
	var _ database.DB = (*poolWrapper)(nil)
}

func TestTxAdapter_ImplementsTx(t *testing.T) {
	var _ database.Tx = (*txAdapter)(nil)
}

func TestConfig_Parsing(t *testing.T) {
	cfg := Config{
		DSN: "postgres://user:pass@localhost:5432/mydb",
	}
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	assert.NoError(t, err)
	assert.NotNil(t, poolCfg)
}

func TestPostgresModule_ReturnsErrorOnInvalidDSN(t *testing.T) {
	_, err := PostgresModule(Config{DSN: "not-a-valid-url"})
	assert.Error(t, err)
}

func TestPostgresModuleNamed_ReturnsErrorOnInvalidDSN(t *testing.T) {
	_, err := PostgresModuleNamed("test", Config{DSN: "postgres://nonexistent:99999/db"})
	assert.Error(t, err)
}
