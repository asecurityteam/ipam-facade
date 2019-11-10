package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	pq "github.com/lib/pq"
)

const (
	createScript             = "2_create.sql"
	pqUniqueViolationErrCode = "23505"
)

// PostgresDB is a SQLDB implementation that uses a PostgreSQL database connection pool.
type PostgresDB struct {
	conn    *sql.DB
	scripts func(name string) (string, error)
	once    sync.Once
}

// RunScript executes a SQL script from disk against the database.
func (db *PostgresDB) RunScript(ctx context.Context, name string) error {
	scriptContent, err := db.scripts(name)
	if err != nil {
		return err
	}
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, scriptContent); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback from %s because of %s", err.Error(), rbErr.Error())
		}
		return err
	}
	return tx.Commit()
}

// Init initializes a connection to a Postgres database according to the environment variables POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DATABASE
func (db *PostgresDB) Init(ctx context.Context, host, port, username, password, dbname string) error {
	var initerr error
	db.once.Do(func() {

		if db.conn == nil {
			sslmode := "disable"
			if host != "localhost" && host != "postgres" {
				sslmode = "require"
			}
			// we establish a connection against a known-to-exist dbname so we can check
			// if we need to create our desired dbname
			psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
				"password=%s dbname=%s sslmode=%s",
				host, port, username, password, "postgres", sslmode)
			pgdb, err := sql.Open("postgres", psqlInfo)
			if err != nil {
				initerr = err
				return // from the unnamed once.Do function
			}

			db.conn = pgdb

			dbExists, err := db.doesDBExist(dbname)
			if err != nil {
				initerr = err
				return // from the unnamed once.Do function
			}

			if !dbExists {
				err = db.create(dbname)
				switch err.(type) {
				case nil:
				case *pq.Error:
					// There can be race conditions if multiple instances of this service
					// attempt to create the underlying database concurrently. If the error
					// code corresponds to duplicate key usage, it is ignored.
					pqErr := err.(*pq.Error)
					if pqErr.Code != pqUniqueViolationErrCode {
						initerr = pqErr
						return
					}
				default:
					initerr = err
					return // from the unnamed once.Do function
				}
			}

			psqlInfo = fmt.Sprintf("host=%s port=%s user=%s "+
				"password=%s dbname=%s sslmode=%s",
				host, port, username, password, dbname, sslmode)
			err = db.Use(ctx, psqlInfo)
			if err != nil {
				initerr = err
				return // from the unnamed once.Do function
			}

		}
		err := db.RunScript(ctx, createScript)
		switch err.(type) {
		case nil:
		case *pq.Error:
			// The same race condition for creating the database exists for
			// concurrently creating the tables, in which case the error
			// is ignored.
			pqErr := err.(*pq.Error)
			if pqErr.Code != pqUniqueViolationErrCode {
				initerr = pqErr
				return
			}
		default:
			initerr = err
			return // from the unnamed once.Do function
		}

	})
	return initerr
}

// Conn returns the currently initialized and open DB connection if one exists, or nil
func (db *PostgresDB) Conn() *sql.DB {
	return db.conn
}

// Use closes any existing database connection, then opens, pings, and sets a new one
// based on the connection string provided in format:
// "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"
func (db *PostgresDB) Use(ctx context.Context, psqlInfo string) error {
	err := db.conn.Close()
	if err != nil {
		return err
	}

	pgdb, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	err = pgdb.Ping()
	if err != nil {
		return err
	}

	db.conn = pgdb
	return nil
}

func (db *PostgresDB) doesDBExist(dbName string) (bool, error) {
	row := db.conn.QueryRow("SELECT datname FROM pg_catalog.pg_database WHERE lower(datname) = lower($1);", dbName)
	var id string
	if err := row.Scan(&id); err != nil {
		switch err {
		case sql.ErrNoRows:
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (db *PostgresDB) create(name string) error {

	_, err := db.conn.Exec("CREATE DATABASE " + name + ";") // nolint
	if err != nil {
		return err
	}

	return nil
}

// CheckDependencies checks to see if the db connection is available
func (db *PostgresDB) CheckDependencies(ctx context.Context) error {
	return db.conn.Ping()
}
