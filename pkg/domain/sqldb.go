package domain

import (
	"context"
	"database/sql"
)

// SQLDB encapsulates "database/sql" from the stdlib with methods for initializing
// a database connection and running SQL scripts against the database
type SQLDB interface {
	// Init creates the database and schema if either or both do not already exist
	Init(ctx context.Context, host string, port string, username string, password string, dbname string) error
	// RunScript runs a named script against a previously initialized database connection
	RunScript(ctx context.Context, name string) error
	// Conn returns an existing, initialized database connection, or nil if one does not exist
	Conn() *sql.DB
	// Use closes any existing database connection and opens a new one using the given connection string
	Use(ctx context.Context, psqlInfo string) error
}
