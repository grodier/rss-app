package pgsql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
)

// DBTX abstracts query methods shared by *sql.DB and *sql.Tx.
// This enables services to work with either and facilitates testing.
type DBTX interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Verify DB implements DBTX at compile time.
var _ DBTX = (*DB)(nil)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type DB struct {
	dsn string
	db  *sql.DB

	MaxOpenConnections int
	MaxIdleConnections int
	MaxIdleTime        time.Duration
}

func NewDB(dsn string) *DB {
	return &DB{dsn: dsn}
}

func (pg *DB) Open() error {
	db, err := sql.Open("postgres", pg.dsn)
	if err != nil {
		return err
	}
	pg.db = db

	pg.db.SetMaxOpenConns(pg.MaxOpenConnections)
	pg.db.SetMaxIdleConns(pg.MaxIdleConnections)
	pg.db.SetConnMaxIdleTime(pg.MaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pg.db.PingContext(ctx)
	if err != nil {
		pg.db.Close()
		return err
	}

	return nil
}

func (pg *DB) Close() error {
	if pg.db != nil {
		return pg.db.Close()
	}
	return nil
}

// DBTX interface implementation

func (pg *DB) Exec(query string, args ...any) (sql.Result, error) {
	return pg.db.Exec(query, args...)
}

func (pg *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return pg.db.ExecContext(ctx, query, args...)
}

func (pg *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return pg.db.Query(query, args...)
}

func (pg *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return pg.db.QueryContext(ctx, query, args...)
}

func (pg *DB) QueryRow(query string, args ...any) *sql.Row {
	return pg.db.QueryRow(query, args...)
}

func (pg *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return pg.db.QueryRowContext(ctx, query, args...)
}
