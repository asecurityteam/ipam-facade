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

	if err := thedb.Init(context.Background(), postgresConfig.Hostname, postgresConfig.Port, postgresConfig.Username, postgresConfig.Password, postgresConfig.DatabaseName); err == nil {
		t.Errorf("DB.Init should have returned a non-nil error")
	}
}

func TestDoesDBExistTrue(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT datname FROM pg_catalog.pg_database WHERE").WithArgs("somename").WillReturnRows(rows).RowsWillBeClosed()

	exists, _ := thedb.doesDBExist("somename")
	if !exists {
		t.Errorf("DB.doesDBExist should have returned true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDoesDBExistFalse(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	mock.ExpectQuery("SELECT datname FROM pg_catalog.pg_database WHERE").WithArgs("somename").WillReturnError(sql.ErrNoRows)

	exists, _ := thedb.doesDBExist("somename")
	if exists {
		t.Errorf("DB.doesDBExist should have returned false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDoesDBExistError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	mock.ExpectQuery("SELECT datname FROM pg_catalog.pg_database WHERE").WithArgs("somename").WillReturnError(errors.New("unexpected error"))

	_, err = thedb.doesDBExist("somename")
	if err == nil {
		t.Errorf("DB.doesDBExist should have returned a non-nil error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateDBSuccess(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	mock.ExpectExec("CREATE DATABASE").WillReturnResult(sqlmock.NewResult(1, 1))

	err = thedb.create("somename")
	if err != nil {
		t.Errorf("DB.create should have returned a nil error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateDBError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	mock.ExpectExec("CREATE DATABASE").WillReturnError(errors.New("unexpected error"))

	err = thedb.create("somename")
	if err == nil {
		t.Errorf("DB.create should have returned a non-nil error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBUseError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockdb.Close()

	thedb := PostgresDB{
		conn: mockdb,
	}

	mock.ExpectClose().WillReturnError(errors.New("unexpected error"))

	err = thedb.Use(context.Background(), "somename")
	if err == nil {
		t.Errorf("DB.use should have returned a non-nil error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
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
}
