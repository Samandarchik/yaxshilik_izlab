package main

import (
	"context"
	_ "embed"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations.sql
var migrationsSQL string

func InitDB(ctx context.Context, pgURL string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(pgURL)
	if err != nil {
		log.Fatalf("db: parse config: %v", err)
	}
	cfg.MaxConns = 20
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("db: connect: %v", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatalf("db: ping: %v", err)
	}
	log.Println("db: connected")

	if _, err := pool.Exec(ctx, migrationsSQL); err != nil {
		log.Fatalf("db: migrate: %v", err)
	}
	log.Println("db: migrations applied")

	return pool
}
