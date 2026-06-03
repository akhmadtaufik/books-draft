package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a PostgreSQL connection pool using environment variables.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	host := mustGetenv("DB_HOST")
	port := mustGetenv("DB_PORT")
	user := mustGetenv("DB_USER")
	password := mustGetenv("DB_PASSWORD")
	dbName := mustGetenv("DB_NAME")
	
	// SSL mode is optionally overridable, default to disable for local
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbName, sslMode,
	)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return pool, nil
}

// SeedDefaultUser inserts a default user if none exists and returns the user ID.
func SeedDefaultUser(ctx context.Context, pool *pgxpool.Pool) (uuid.UUID, error) {
	const email = "author@novel.app"
	var userID uuid.UUID

	// Try to find existing user first.
	err := pool.QueryRow(ctx,
		"SELECT id FROM users WHERE email = $1", email,
	).Scan(&userID)

	if err == nil {
		log.Printf("📖 Default user found: %s", userID)
		return userID, nil
	}

	// User doesn't exist — insert one.
	err = pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash)
		 VALUES ($1, $2)
		 RETURNING id`,
		email, "not-a-real-hash",
	).Scan(&userID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("seed default user: %w", err)
	}

	log.Printf("🌱 Seeded default user: %s (%s)", email, userID)
	return userID, nil
}

func mustGetenv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		log.Fatalf("Environment variable %s is missing", key)
	}
	return val
}
