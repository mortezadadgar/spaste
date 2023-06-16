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
func (s *SQLiteStore) Create(p *modules.Paste) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.DB.ExecContext(ctx,
		"INSERT INTO pastes(text, lang, line_count, addr, created_at) values(?, ?, ?, ?, ?)",
		p.Text,
		p.Lang,
		p.LineCount,
		p.Address,
		p.TimeStamp,
	)
	if err != nil {
		return fmt.Errorf("failed to insert to sqlite table: %v", err)
	}

	return nil
}

// Get gets paste by its address from sqlite store.
func (s *SQLiteStore) Get(addr string) (*modules.Paste, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p modules.Paste
	err := s.DB.QueryRowContext(ctx,
		"SELECT * FROM pastes WHERE addr = ?", addr).Scan(
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
