package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"videocall/internal/infrastructure/config"

	_ "github.com/go-sql-driver/mysql"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(cfg *config.Storage) (*Database, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxConns)
	db.SetMaxIdleConns(cfg.DBIdleConns)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Connected to MariaDB")

	return &Database{db: db}, nil
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}
