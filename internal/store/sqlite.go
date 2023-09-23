package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mortezadadgar/spaste/internal/config"
	"github.com/mortezadadgar/spaste/internal/modules"

	// sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	DB *sql.DB
}

// NewSQLiteStore returns a instance of SQLiteStore.
func NewSQLiteStore(c config.Config) (*SQLiteStore, error) {
	store := &SQLiteStore{}
	if err := store.init(c); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) init(config config.Config) error {
	DB, err := sql.Open("sqlite3", config.ConnectionString)
	if err != nil {
		return fmt.Errorf("falied to open sqlite database %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = DB.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("falied to ping sqlite database %v", err)
	}

	s.DB = DB

	return nil
}

// Create inserts paste to sqlite store.
func (s *SQLiteStore) Create(ctx context.Context, m *modules.Paste) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		"INSERT INTO pastes(text, lang, line_count, addr, created_at) values(?, ?, ?, ?, ?)",
		m.Text,
		m.Lang,
		m.LineCount,
		m.Address,
		m.TimeStamp,
	)
	if err != nil {
		return fmt.Errorf("failed to insert to sqlite table: %v", err)
	}

	r, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if r != 1 {
		return fmt.Errorf("expected to only one row affected")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction")
	}

	return nil
}

// Get gets paste by its address from sqlite store.
func (s *SQLiteStore) Get(ctx context.Context, address string) (*modules.Paste, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	var p modules.Paste
	err = tx.QueryRowContext(ctx,
		"SELECT * FROM pastes WHERE addr = ?", address).Scan(
		&p.ID,
		&p.Text,
		&p.Lang,
		&p.LineCount,
		&p.Address,
		&p.TimeStamp,
	)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed to select from sqlite table: %v", err)
	}

	return &p, nil
}
