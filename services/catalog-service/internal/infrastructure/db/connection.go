package db

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// ConnectionPool manages database connection pool settings
type ConnectionPool struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConnectionPool returns default connection pool settings
func DefaultConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}
}

// ConfigureConnectionPool configures the connection pool
func ConfigureConnectionPool(db *gorm.DB, pool *ConnectionPool) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(pool.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(pool.ConnMaxIdleTime)

	return nil
}

// HealthCheck checks database connection health
func HealthCheck(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// GetStats returns database connection statistics
func GetStats(db *gorm.DB) (*sql.DBStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	stats := sqlDB.Stats()
	return &stats, nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Transaction executes a function within a transaction
func Transaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}

