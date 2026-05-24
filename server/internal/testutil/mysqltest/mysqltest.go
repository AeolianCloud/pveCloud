package mysqltest

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	mysqlconfig "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const envDSN = "PVECLOUD_TEST_MYSQL_DSN"

// Open creates an isolated temporary MySQL database for integration tests.
// Tests are skipped unless PVECLOUD_TEST_MYSQL_DSN is set, so the default
// unit-test path never mutates a developer's local pveCloud database.
func Open(t *testing.T) *gorm.DB {
	t.Helper()
	rawDSN := strings.TrimSpace(getenv(envDSN))
	if rawDSN == "" {
		t.Skipf("set %s to run MySQL integration tests", envDSN)
	}

	cfg, err := mysqlconfig.ParseDSN(rawDSN)
	if err != nil {
		t.Fatalf("parse %s: %v", envDSN, err)
	}
	cfg.ParseTime = true

	serverCfg := cfg.Clone()
	serverCfg.DBName = ""
	serverDB, err := sql.Open("mysql", serverCfg.FormatDSN())
	if err != nil {
		t.Fatalf("open MySQL server connection: %v", err)
	}
	t.Cleanup(func() { _ = serverDB.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := serverDB.PingContext(ctx); err != nil {
		t.Fatalf("ping MySQL server: %v", err)
	}

	dbName := fmt.Sprintf("pvecloud_it_%d", time.Now().UnixNano())
	if _, err := serverDB.ExecContext(ctx, "CREATE DATABASE `"+dbName+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
		t.Fatalf("create temporary database: %v", err)
	}
	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer dropCancel()
		_, _ = serverDB.ExecContext(dropCtx, "DROP DATABASE IF EXISTS `"+dbName+"`")
	})

	cfg.DBName = dbName
	db, err := gorm.Open(mysql.Open(cfg.FormatDSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("open temporary database: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

// Exec executes schema setup statements in order.
func Exec(t *testing.T, db *gorm.DB, statements ...string) {
	t.Helper()
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("exec schema statement %q: %v", statement, err)
		}
	}
}

func getenv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
