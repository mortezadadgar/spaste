package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mortezadadgar/spaste/internal/config"
	"github.com/mortezadadgar/spaste/internal/paste"

	// sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	DB *sql.DB
}

// New returns a instance of SQLiteStore.
func New(c config.Config) (*SQLite, error) {
	store := &SQLite{}
	if err := store.connect(c); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *SQLite) Close() error {
	return s.DB.Close()
}

func (s *SQLite) connect(config config.Config) error {
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
func (s *SQLite) Create(ctx context.Context, m paste.Module) error {
	result, err := s.DB.ExecContext(ctx,
		"INSERT INTO pastes(text, lang, line_count, addr, created_at) values(?, ?, ?, ?, ?)",
		&m.Text,
		&m.Lang,
		&m.LineCount,
		&m.Address,
		&m.TimeStamp,
	)
	if err != nil {
		return fmt.Errorf("failed to insert to sqlite table: %v", err)
	}

	nrows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if nrows != 1 {
		return fmt.Errorf("expected to only one row affected")
	}

	return nil
}

// Get gets paste by its address from sqlite store.
func (s *SQLite) Get(ctx context.Context, address string) (paste.Module, error) {
	var p paste.Module
	err := s.DB.QueryRowContext(ctx,
		"SELECT * FROM pastes WHERE addr = ?", address).Scan(
		&p.ID,
		&p.Text,
		&p.Lang,
		&p.LineCount,
		&p.Address,
		&p.TimeStamp,
	)
	if err == sql.ErrNoRows {
		return paste.Module{}, paste.ErrNoPasteFound
	} else if err != nil {
		return paste.Module{}, fmt.Errorf("failed to select from sqlite table: %v", err)
	}

	return p, nil
}
