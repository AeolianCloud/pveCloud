package testutil

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	mysqlx "github.com/go-sql-driver/mysql"
)

func RunMigrations(db *sql.DB) error {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	entries, err := os.ReadDir(filepath.Join(root, "migrations"))
	if err != nil {
		return err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(root, "migrations", name))
		if err != nil {
			return err
		}
		for _, stmt := range splitSQLStatements(string(data)) {
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil && !isAlreadyExistsError(err) {
				return err
			}
		}
	}
	return nil
}

func ResetTables(db *sql.DB) error {
	tables := []string{
		"async_task_logs",
		"async_tasks",
		"instance_actions",
		"instance_services",
		"instances",
		"payment_callback_logs",
		"payment_orders",
		"billing_records",
		"order_items",
		"orders",
		"resource_reservations",
		"sku_region_node_bindings",
		"resource_nodes",
		"regions",
		"product_skus",
		"products",
		"admins",
		"users",
	}

	if _, err := db.Exec(`SET FOREIGN_KEY_CHECKS = 0`); err != nil {
		return err
	}
	defer func() {
		_, _ = db.Exec(`SET FOREIGN_KEY_CHECKS = 1`)
	}()

	for _, table := range tables {
		if _, err := db.Exec(`TRUNCATE TABLE ` + table); err != nil {
			return err
		}
	}
	return nil
}

func splitSQLStatements(sqlText string) []string {
	parts := strings.Split(sqlText, ";")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		items = append(items, stmt)
	}
	return items
}

func isAlreadyExistsError(err error) bool {
	mysqlErr, ok := err.(*mysqlx.MySQLError)
	if !ok {
		return false
	}
	return mysqlErr.Number == 1050
}
