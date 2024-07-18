package repository

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
)

var Dbpool *pgxpool.Pool

func InitDBConn(ctx context.Context) (dbpool *pgxpool.Pool, err error) {
	url := "host=127.0.0.1 port=5432 user=postgres password=193566 dbname=authdb sslmode=disable"
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		err = fmt.Errorf("failed to parse pg config: %w", err)
		return
	}
	cfg.MaxConns = int32(5)
	cfg.MinConns = int32(1)
	cfg.HealthCheckPeriod = 1 * time.Minute
	cfg.MaxConnLifetime = 24 * time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 1 * time.Second
	cfg.ConnConfig.DialFunc = (&net.Dialer{
		KeepAlive: cfg.HealthCheckPeriod,
		Timeout:   cfg.ConnConfig.ConnectTimeout,
	}).DialContext
	dbpool, err = pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		err = fmt.Errorf("failed to connect config: %w", err)
		return
	}
	Dbpool = dbpool
	return
}
