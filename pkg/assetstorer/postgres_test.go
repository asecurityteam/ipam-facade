package assetstorer

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestPostgresPhysicalAssetStorer_StorePhysicalAssets_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSQLDB := NewMockSQLDB(ctrl)

	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err)
	defer mockdb.Close()
	mockSQLDB.EXPECT().Conn().Return(mockdb)

	device := domain.Device{
		ID:       "1",
		IP:       "127.0.0.1",
		SubnetID: "2",
	}
	subnet := domain.Subnet{
		ID:         "1",
		Network:    "127.0.0.0/31",
		MaskBits:   1,
		Location:   "",
		CustomerID: "1",
	}
	customer := domain.Customer{
		ID:            "1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Security",
	}

	ipamData := domain.IPAMData{
		Devices:   []domain.Device{device},
		Subnets:   []domain.Subnet{subnet},
		Customers: []domain.Customer{customer},
	}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM customers").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM subnets").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM ips").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO customers").WithArgs(customer.ID, customer.ResourceOwner, customer.BusinessUnit).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO subnets").WithArgs(subnet.ID, subnet.Network, subnet.Location, subnet.CustomerID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO ips").WithArgs(device.IP, device.SubnetID, device.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	storer := PostgresPhysicalAssetStorer{DB: mockSQLDB}
	e := storer.StorePhysicalAssets(context.Background(), ipamData)
	require.Nil(t, e)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestPostgresPhysicalAssetStorer_StorePhysicalAssetsNoDeviceID_Success(t *testing.T) {
	// I don't know if IPAM would ever return device info where the
	// device lacks an ID, but we're gonna handle it if it does...
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSQLDB := NewMockSQLDB(ctrl)

	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err)
	defer mockdb.Close()
	mockSQLDB.EXPECT().Conn().Return(mockdb)

	device := domain.Device{
		// ID:       "1",  // intentionally commented out
		IP:       "127.0.0.1",
		SubnetID: "2",
	}
	subnet := domain.Subnet{
		ID:         "1",
		Network:    "127.0.0.0/31",
		MaskBits:   1,
		Location:   "",
		CustomerID: "1",
	}
	customer := domain.Customer{
		ID:            "1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Security",
	}

	ipamData := domain.IPAMData{
		Devices:   []domain.Device{device},
		Subnets:   []domain.Subnet{subnet},
		Customers: []domain.Customer{customer},
	}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM customers").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM subnets").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM ips").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO customers").WithArgs(customer.ID, customer.ResourceOwner, customer.BusinessUnit).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO subnets").WithArgs(subnet.ID, subnet.Network, subnet.Location, subnet.CustomerID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO ips").WithArgs(device.IP, device.SubnetID, nil).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	storer := PostgresPhysicalAssetStorer{DB: mockSQLDB}
	e := storer.StorePhysicalAssets(context.Background(), ipamData)
	require.Nil(t, e)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestPostgresPhysicalAssetStorer_StorePhysicalAssetsNoCustomerID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSQLDB := NewMockSQLDB(ctrl)

	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err)
	defer mockdb.Close()
	mockSQLDB.EXPECT().Conn().Return(mockdb)

	subnet := domain.Subnet{
		ID:         "1",
		Network:    "127.0.0.0/31",
		MaskBits:   1,
		Location:   "",
		CustomerID: "0", // the zero value when IPAM returns "null" for the customer_id of a subnet
	}

	ipamData := domain.IPAMData{
		Subnets: []domain.Subnet{subnet},
	}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM customers").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM subnets").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM ips").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO subnets").WithArgs(subnet.ID, subnet.Network, subnet.Location, sql.NullString{}).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	storer := PostgresPhysicalAssetStorer{DB: mockSQLDB}
	e := storer.StorePhysicalAssets(context.Background(), ipamData)
	require.Nil(t, e)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestPostgresPhysicalAssetStorer_StorePhysicalAssets_RollbackError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSQLDB := NewMockSQLDB(ctrl)

	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err)
	defer mockdb.Close()
	mockSQLDB.EXPECT().Conn().Return(mockdb)

	device := domain.Device{
		ID:       "1",
		IP:       "127.0.0.1",
		SubnetID: "1",
	}
	subnet := domain.Subnet{
		ID:         "1",
		Network:    "127.0.0.0/31",
		MaskBits:   1,
		Location:   "",
		CustomerID: "1",
	}
	customer := domain.Customer{
		ID:            "1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Security",
	}

	ipamData := domain.IPAMData{
		Devices:   []domain.Device{device},
		Subnets:   []domain.Subnet{subnet},
		Customers: []domain.Customer{customer},
	}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM customers").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM subnets").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM ips").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO customers").WithArgs(customer.ID, customer.ResourceOwner, customer.BusinessUnit).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO subnets").WithArgs(subnet.ID, subnet.Network, subnet.Location, subnet.CustomerID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO ips").WithArgs(device.IP, device.SubnetID, device.ID).WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback().WillReturnError(fmt.Errorf("rollback error"))

	storer := PostgresPhysicalAssetStorer{DB: mockSQLDB}
	e := storer.StorePhysicalAssets(context.Background(), ipamData)
	require.Error(t, e)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestPostgresPhysicalAssetStorer_StorePhysicalAssets_TxBeginError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSQLDB := NewMockSQLDB(ctrl)

	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err)
	defer mockdb.Close()
	mockSQLDB.EXPECT().Conn().Return(mockdb)

	mock.ExpectBegin().WillReturnError(fmt.Errorf("could not start transaction"))

	storer := PostgresPhysicalAssetStorer{DB: mockSQLDB}
	e := storer.StorePhysicalAssets(context.Background(), domain.IPAMData{})
	require.Error(t, e)
}

func TestPostgresPhysicalAssetStorer_storeSubnet_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSQLDB := NewMockSQLDB(ctrl)

	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err)
	defer mockdb.Close()
	mockSQLDB.EXPECT().Conn().Return(mockdb)

	subnet := domain.Subnet{
		ID:         "1",
		Network:    "127.0.0.0/31",
		MaskBits:   1,
		Location:   "",
		CustomerID: "1",
	}
	ipamData := domain.IPAMData{
		Subnets: []domain.Subnet{subnet},
	}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM customers").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM subnets").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM ips").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO subnets").WithArgs(subnet.ID, subnet.Network, subnet.Location, subnet.CustomerID).WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	storer := PostgresPhysicalAssetStorer{DB: mockSQLDB}
	e := storer.StorePhysicalAssets(context.Background(), ipamData)
	require.Error(t, e)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestPostgresPhysicalAssetStorer_storeCustomer_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSQLDB := NewMockSQLDB(ctrl)

	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err)
	defer mockdb.Close()
	mockSQLDB.EXPECT().Conn().Return(mockdb)

	customer := domain.Customer{
		ID:            "1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Security",
	}

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{customer},
	}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM customers").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM subnets").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM ips").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO customers").WithArgs(customer.ID, customer.ResourceOwner, customer.BusinessUnit).WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	storer := PostgresPhysicalAssetStorer{DB: mockSQLDB}
	e := storer.StorePhysicalAssets(context.Background(), ipamData)
	require.Error(t, e)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestPostgresPhysicalAssetStorer_storeIP_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSQLDB := NewMockSQLDB(ctrl)

	mockdb, mock, err := sqlmock.New()
	require.Nil(t, err)
	defer mockdb.Close()
	mockSQLDB.EXPECT().Conn().Return(mockdb)

	device := domain.Device{
		ID:       "1",
		IP:       "127.0.0.1",
		SubnetID: "2",
	}
	ipamData := domain.IPAMData{
		Devices: []domain.Device{device},
	}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM customers").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM subnets").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM ips").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO ips").WithArgs(device.IP, device.SubnetID, device.ID).WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	storer := PostgresPhysicalAssetStorer{DB: mockSQLDB}
	e := storer.StorePhysicalAssets(context.Background(), ipamData)
	require.Error(t, e)
	require.Nil(t, mock.ExpectationsWereMet())
}
