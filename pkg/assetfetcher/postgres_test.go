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

func TestFetchPhysicalAssetSubnetFoundNoJoinedCustomer(t *testing.T) {
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
		nil, "alice@example.com", "Acme", "127.0.0.1/32", "Home", nil, 1, nil)
	mock.ExpectQuery("SELECT").WillReturnRows(rows).RowsWillBeClosed()
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	// fields are intentionally commented out
	expectedAsset := domain.PhysicalAsset{
		IP: "127.0.0.1",
		// ResourceOwner: "alice@example.com",
		// BusinessUnit:  "Acme",
		Network:  "127.0.0.1/32",
		Location: "Home",
		DeviceID: 0,
		SubnetID: 1,
		// CustomerID: 1,
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

func TestFetchSubnets(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer mockdb.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocksqldb := NewMockSQLDB(ctrl)
	mocksqldb.EXPECT().Conn().Return(mockdb)
	rows := sqlmock.NewRows([]string{
		"network", "location", "resource_owner", "business_unit"}).
		AddRow("127.0.0.1/32", "Home", "alice@example.com", "Acme").
		AddRow("127.0.0.2/32", "Home", "alice@example.com", "Acme")
	mock.ExpectQuery("SELECT").WillReturnRows(rows).RowsWillBeClosed()
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	expected := []domain.AssetSubnet{
		{
			ResourceOwner: "alice@example.com",
			BusinessUnit:  "Acme",
			Network:       "127.0.0.1/32",
			Location:      "Home",
		},
		{
			ResourceOwner: "alice@example.com",
			BusinessUnit:  "Acme",
			Network:       "127.0.0.2/32",
			Location:      "Home",
		},
	}

	subnets, err := fetcher.FetchSubnets(context.Background(), 2, 0)
	require.Nil(t, err)
	require.Equal(t, expected, subnets)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFetchSubnetsQueryError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer mockdb.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocksqldb := NewMockSQLDB(ctrl)
	mocksqldb.EXPECT().Conn().Return(mockdb)
	mock.ExpectQuery("SELECT").WillReturnError(errors.New(""))
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	_, err = fetcher.FetchSubnets(context.Background(), 2, 0)
	require.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFetchSubnetsScanError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer mockdb.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocksqldb := NewMockSQLDB(ctrl)
	mocksqldb.EXPECT().Conn().Return(mockdb)
	rows := sqlmock.NewRows([]string{
		"network", "location", "resource_owner", "business_unit"}).
		AddRow(nil, "Home", "alice@example.com", "Acme")
	mock.ExpectQuery("SELECT").WillReturnRows(rows).RowsWillBeClosed()
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	_, err = fetcher.FetchSubnets(context.Background(), 2, 0)
	require.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFetchIPs(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer mockdb.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocksqldb := NewMockSQLDB(ctrl)
	mocksqldb.EXPECT().Conn().Return(mockdb)
	rows := sqlmock.NewRows([]string{
		"ip", "network", "location", "resource_owner", "business_unit"}).
		AddRow("127.0.0.1", "127.0.0.1/32", "Home", "alice@example.com", "Acme").
		AddRow("127.0.0.1", "127.0.0.2/32", "Home", "alice@example.com", "Acme")
	mock.ExpectQuery("SELECT").WillReturnRows(rows).RowsWillBeClosed()
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	expected := []domain.AssetIP{
		{
			IP:            "127.0.0.1",
			ResourceOwner: "alice@example.com",
			BusinessUnit:  "Acme",
			Network:       "127.0.0.1/32",
			Location:      "Home",
		},
		{
			IP:            "127.0.0.1",
			ResourceOwner: "alice@example.com",
			BusinessUnit:  "Acme",
			Network:       "127.0.0.2/32",
			Location:      "Home",
		},
	}

	ips, err := fetcher.FetchIPs(context.Background(), 2, 0)
	require.Nil(t, err)
	require.Equal(t, expected, ips)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFetchIPsQueryError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer mockdb.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocksqldb := NewMockSQLDB(ctrl)
	mocksqldb.EXPECT().Conn().Return(mockdb)
	mock.ExpectQuery("SELECT").WillReturnError(errors.New(""))
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	_, err = fetcher.FetchIPs(context.Background(), 2, 0)
	require.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFetchIPsScanError(t *testing.T) {
	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err, "an error '%s' was not expected when opening a stub database connection", err)
	defer mockdb.Close()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocksqldb := NewMockSQLDB(ctrl)
	mocksqldb.EXPECT().Conn().Return(mockdb)
	rows := sqlmock.NewRows([]string{
		"ip", "network", "location", "resource_owner", "business_unit"}).
		AddRow(nil, "127.0.0.1/32", "Home", "alice@example.com", "Acme")
	mock.ExpectQuery("SELECT").WillReturnRows(rows).RowsWillBeClosed()
	fetcher := PostgresPhysicalAssetFetcher{DB: mocksqldb}

	_, err = fetcher.FetchIPs(context.Background(), 2, 0)
	require.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
