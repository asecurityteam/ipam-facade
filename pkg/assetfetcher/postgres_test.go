package assetfetcher

import (
	context "context"
	sql "database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

func TestFetchPhysicalAssetDeviceFound(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer mockdb.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocksqldb := NewMockSQLDB(ctrl)
	mocksqldb.EXPECT().Conn().Return(mockdb)
	rows := sqlmock.NewRows([]string{
		"ip", "resource_owner", "business_unit", "network", "location",
		"device_id", "subnet_id", "customer_id"}).AddRow(
		"127.0.0.1", "alice@example.com", "Acme", "127.0.0.1/32", "Home", 1, 1, 1)
	mock.ExpectQuery("SELECT").WillReturnRows(rows).RowsWillBeClosed()
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	expectedAsset := domain.PhysicalAsset{
		IP:            "127.0.0.1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Acme",
		Network:       "127.0.0.1/32",
		Location:      "Home",
		DeviceID:      1,
		SubnetID:      1,
		CustomerID:    1,
	}

	asset, err := fetcher.FetchPhysicalAsset(context.Background(), "127.0.0.1")
	require.Nil(t, err)
	require.Equal(t, expectedAsset, asset)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFetchPhysicalAssetSubnetFound(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer mockdb.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocksqldb := NewMockSQLDB(ctrl)
	mocksqldb.EXPECT().Conn().Return(mockdb)
	rows := sqlmock.NewRows([]string{
		"ip", "resource_owner", "business_unit", "network", "location",
		"device_id", "subnet_id", "customer_id"}).AddRow(
		nil, "alice@example.com", "Acme", "127.0.0.1/32", "Home", nil, 1, 1)
	mock.ExpectQuery("SELECT").WillReturnRows(rows).RowsWillBeClosed()
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	expectedAsset := domain.PhysicalAsset{
		IP:            "127.0.0.1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Acme",
		Network:       "127.0.0.1/32",
		Location:      "Home",
		DeviceID:      0,
		SubnetID:      1,
		CustomerID:    1,
	}

	asset, err := fetcher.FetchPhysicalAsset(context.Background(), "127.0.0.1")
	require.Nil(t, err)
	require.Equal(t, expectedAsset, asset)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFetchPhysicalAssetNoResults(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer mockdb.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocksqldb := NewMockSQLDB(ctrl)
	mocksqldb.EXPECT().Conn().Return(mockdb)
	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	_, err = fetcher.FetchPhysicalAsset(context.Background(), "127.0.0.1")
	require.Equal(t, domain.AssetNotFound{Inner: sql.ErrNoRows, IP: "127.0.0.1"}, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFetchPhysicalAssetUnexpectedError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer mockdb.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocksqldb := NewMockSQLDB(ctrl)
	mocksqldb.EXPECT().Conn().Return(mockdb)
	dberr := errors.New("unexpected error")
	mock.ExpectQuery("SELECT").WillReturnError(dberr)
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	_, err = fetcher.FetchPhysicalAsset(context.Background(), "127.0.0.1")
	require.Equal(t, dberr, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}