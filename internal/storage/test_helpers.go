package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-redis/redis/v8"
)

// getTestMySQLDSN returns MySQL DSN for testing
// Uses environment variable or defaults to a test database
func getTestMySQLDSN() string {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		// Default test DSN - adjust as needed
		dsn = "root:password@tcp(localhost:3306)/favorites_test?charset=utf8mb4&parseTime=True&loc=Local"
	}
	return dsn
}

// getTestRedisAddr returns Redis address for testing
func getTestRedisAddr() string {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return addr
}

// getTestRedisPassword returns Redis password for testing
func getTestRedisPassword() string {
	return os.Getenv("TEST_REDIS_PASSWORD")
}

// createTestUser creates a test user in the database
func createTestUser(db *sql.DB, userReference, email, name string) error {
	query := `INSERT INTO users (reference, email, name) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE email = VALUES(email), name = VALUES(name)`
	_, err := db.Exec(query, userReference, email, name)
	if err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}
	return nil
}

// skipIfMySQLUnavailable skips the test if MySQL is not available
func skipIfMySQLUnavailable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping MySQL integration test in short mode")
	}

	dsn := getTestMySQLDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skipf("Skipping MySQL test: failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Skipf("Skipping MySQL test: MySQL is not available: %v", err)
	}
}

// skipIfRedisUnavailable skips the test if Redis is not available
func skipIfRedisUnavailable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	addr := getTestRedisAddr()
	password := getTestRedisPassword()
	
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping Redis test: Redis is not available: %v", err)
	}
}



