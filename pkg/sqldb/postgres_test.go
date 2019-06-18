package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

var scriptText = "SELECT 1"
var scriptFound = func(string) (string, error) { return scriptText, nil }

func TestFetchPhysicalAssetsErrors(t *testing.T) {
	thedb := PostgresDB{}

	postgresConfig := PostgresConfig{
		Hostname:     "this is not a hostname",
		Port:         "99",
		Username:     "me!",
		Password:     "mypassword!",
		DatabaseName: "name",
	}

	require.Error(t, thedb.Init(context.Background(), postgresConfig.Hostname, postgresConfig.Port,
		postgresConfig.Username, postgresConfig.Password, postgresConfig.DatabaseName),
		"DB.Init should have returned a non-nil error")
}

func TestDoesDBExistTrue(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection")
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT datname FROM pg_catalog.pg_database WHERE").WithArgs("somename").WillReturnRows(rows).RowsWillBeClosed()

	exists, _ := thedb.doesDBExist("somename")
	require.True(t, exists, "DB.doesDBExist should have returned true")
	require.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations: %s")
}

func TestDoesDBExistFalse(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection")
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	mock.ExpectQuery("SELECT datname FROM pg_catalog.pg_database WHERE").WithArgs("somename").WillReturnError(sql.ErrNoRows)

	exists, _ := thedb.doesDBExist("somename")
	require.False(t, exists, "DB.doesDBExist should have returned false")
	require.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations: %s")
}

func TestDoesDBExistError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection")
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	mock.ExpectQuery("SELECT datname FROM pg_catalog.pg_database WHERE").WithArgs("somename").WillReturnError(errors.New("unexpected error"))

	_, err = thedb.doesDBExist("somename")
	require.Error(t, err, "DB.doesDBExist should have returned a non-nil error")
	require.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations: %s")
}

func TestCreateDBSuccess(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection")
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	mock.ExpectExec("CREATE DATABASE").WillReturnResult(sqlmock.NewResult(1, 1))

	err = thedb.create("somename")
	require.Nil(t, err, "DB.create should have returned a nil error")
	require.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations: %s")
}

func TestCreateDBError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection")
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	mock.ExpectExec("CREATE DATABASE").WillReturnError(errors.New("unexpected error"))

	err = thedb.create("somename")
	require.Error(t, err, "DB.create should have returned a non-nil error")
	require.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations: %s")
}

func TestDBUseError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection")
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	mock.ExpectClose().WillReturnError(errors.New("unexpected error"))

	err = thedb.Use(context.Background(), "somename")
	require.Error(t, err, "DB.use should have returned a non-nil error")
	require.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations: %s")
}

func TestRunScriptTxFailBegin(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockdb.Close()

	mock.ExpectBegin().WillReturnError(errors.New("tx fail"))
	thedb := PostgresDB{
		conn:    mockdb,
		scripts: scriptFound,
	}

	require.Error(t, thedb.RunScript(context.Background(), "script1"))
	require.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations: %s")
}

func TestRunScriptTxRollbackOnFail(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockdb.Close()

	mock.ExpectBegin()
	mock.ExpectExec(scriptText).WillReturnError(errors.New("bad query"))
	mock.ExpectRollback()
	thedb := PostgresDB{
		conn:    mockdb,
		scripts: scriptFound,
	}

	require.Error(t, thedb.RunScript(context.Background(), "script1"))
	require.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations: %s")
}

func TestRunScriptTxRollbackFail(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockdb.Close()

	mock.ExpectBegin()
	mock.ExpectExec(scriptText).WillReturnError(errors.New("bad query"))
	mock.ExpectRollback().WillReturnError(errors.New("bad rollback"))
	thedb := PostgresDB{
		conn:    mockdb,
		scripts: scriptFound,
	}

	require.Error(t, thedb.RunScript(context.Background(), "script1"))
	require.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations: %s")
}

func TestRunScriptTxCommit(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockdb.Close()

	mock.ExpectBegin()
	mock.ExpectExec(scriptText).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	thedb := PostgresDB{
		conn:    mockdb,
		scripts: scriptFound,
	}
	require.NoError(t, thedb.RunScript(context.Background(), "script1"))
	require.Nil(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations: %s")
}
