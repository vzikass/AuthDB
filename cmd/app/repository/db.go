package repository

import (
	"context"
	"fmt"
	"time"
	
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
)

var(
	Dbpool *pgxpool.Pool
	TestDbpool *pgxpool.Pool
) 
	
// Initialize maindb and testdb
func InitDBConn(ctx context.Context, dbURL string) (dbpool *pgxpool.Pool, err error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		err = fmt.Errorf("failed to parse pg config: %v", err)
		return
	}

	// MaxConns is the maximum size of the pool
	cfg.MaxConns = int32(10)
	cfg.MinConns = int32(1)
	// HealthCheckPeriod is the duration between checks of the health of idle connections.
	cfg.HealthCheckPeriod = 1 * time.Minute
	// MaxConnLifetime is the duration since creation after which a connection will be automatically closed.
	cfg.MaxConnLifetime = 24 * time.Hour
	// It's like maxconnlifetime but it'll be closed by the health check.
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	
	dbpool, err = pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		err = fmt.Errorf("failed to connect config: %w", err)
		return
	}

	TestDbpool = dbpool
	Dbpool = dbpool
	return
}
