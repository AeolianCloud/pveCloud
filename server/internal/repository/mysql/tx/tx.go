package tx

import (
	"context"

	"gorm.io/gorm"
)

type Handle = *gorm.DB

type Manager struct {
	db *gorm.DB
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) Within(fn func(Handle) error) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (m *Manager) WithinContext(ctx context.Context, fn func(Handle) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (m *Manager) DB() *gorm.DB {
	return m.db
}
