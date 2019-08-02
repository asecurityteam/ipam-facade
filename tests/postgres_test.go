//+build integration

package inttest

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/asecurityteam/ipam-facade/pkg/assetfetcher"
	"github.com/asecurityteam/ipam-facade/pkg/assetstorer"
	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/asecurityteam/ipam-facade/pkg/sqldb"
	"github.com/asecurityteam/settings"
	packr "github.com/gobuffalo/packr/v2"
	pq "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// connect to postgres database to wipe existing tables
	pgdb, err := connectToDB("postgres")
	if err != nil {
		panic(err.Error())
	}
	defer pgdb.Close()
	dbname := os.Getenv("POSTGRES_DATABASENAME")
	if err = wipeDatabase(pgdb, dbname); err != nil {
		panic(err.Error())
	}
	code := m.Run()
	if err = wipeDatabase(pgdb, dbname); err != nil {
		panic(err.Error())
	}
	os.Exit(code)
}

// TestNoDBRows verifies that sql.ErrNoRows is returned when no devices or subnets exist
func TestNoDBRows(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresConfigComponent := &sqldb.PostgresConfigComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresConfigComponent, db))
	defer db.Conn().Close()

	// code should tolerate no data in the tables
	fetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: db}
	_, err = fetcher.FetchPhysicalAsset(context.Background(), "0.0.0.0")
	require.Equal(t, domain.AssetNotFound{Inner: sql.ErrNoRows, IP: "0.0.0.0"}, err)
}

// TestSubnetOnly verifies that an IP address within a subnet will return a match, even when
// no corresponding device exists
func TestSubnetOnly(t *testing.T) {
	t.Parallel()

	customerID := rand.Int31()
	subnetID := rand.Int31()

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.Itoa(int(customerID)),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         strconv.Itoa(int(subnetID)),
				Network:    "1.0.0.0/24",
				Location:   "Home",
				CustomerID: strconv.Itoa(int(customerID)),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresConfigComponent := &sqldb.PostgresConfigComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresConfigComponent, db))
	defer db.Conn().Close()

	storer := &assetstorer.PostgresPhysicalAssetStorer{DB: db}
	err = storer.StorePhysicalAssets(ctx, ipamData)
	require.Nil(t, err)

	fetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: db}
	asset, err := fetcher.FetchPhysicalAsset(ctx, "1.0.0.1")
	require.Nil(t, err)

	expected := domain.PhysicalAsset{
		IP:            "1.0.0.1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Example Team",
		Network:       "1.0.0.0/24",
		Location:      "Home",
		DeviceID:      0,
		SubnetID:      int64(subnetID),
		CustomerID:    int64(customerID),
	}

	require.Equal(t, expected, asset)
}

// TestDeviceAndSubnet verifies that a query for an IP address with a device match
// returns both device and subnet information
func TestDeviceAndSubnet(t *testing.T) {
	t.Parallel()

	customerID1 := rand.Int31()
	customerID2 := rand.Int31()
	subnetID := rand.Int31()
	deviceID := rand.Int31()

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.Itoa(int(customerID1)),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
			{
				ID:            strconv.Itoa(int(customerID2)),
				ResourceOwner: "bob@example.com",
				BusinessUnit:  "Team Example",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         strconv.Itoa(int(subnetID)),
				Network:    "2.0.0.0/24",
				Location:   "Home",
				CustomerID: strconv.Itoa(int(customerID2)),
			},
		},
		Devices: []domain.Device{
			{
				ID:       strconv.Itoa(int(deviceID)),
				IP:       "2.0.0.1",
				SubnetID: strconv.Itoa(int(subnetID)),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresConfigComponent := &sqldb.PostgresConfigComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresConfigComponent, db))
	defer db.Conn().Close()

	storer := &assetstorer.PostgresPhysicalAssetStorer{DB: db}
	err = storer.StorePhysicalAssets(ctx, ipamData)
	require.Nil(t, err)

	fetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: db}
	asset, err := fetcher.FetchPhysicalAsset(ctx, "2.0.0.1")
	require.Nil(t, err)

	expected := domain.PhysicalAsset{
		IP:            "2.0.0.1",
		ResourceOwner: "bob@example.com",
		BusinessUnit:  "Team Example",
		Network:       "2.0.0.0/24",
		Location:      "Home",
		DeviceID:      int64(deviceID),
		SubnetID:      int64(subnetID),
		CustomerID:    int64(customerID2),
	}

	require.Equal(t, expected, asset)
}

// TestDeviceAndSubnet verifies that a query for an IP address with a device match
// returns both device and subnet information
func TestDeviceAndSubnetNoDeviceID(t *testing.T) {
	// I don't know if IPAM would ever return device info where the
	// device lacks an ID, but we're gonna handle it if it does...
	t.Parallel()

	customerID1 := rand.Int31()
	customerID2 := rand.Int31()
	subnetID := rand.Int31()

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.Itoa(int(customerID1)),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
			{
				ID:            strconv.Itoa(int(customerID2)),
				ResourceOwner: "bob@example.com",
				BusinessUnit:  "Team Example",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         strconv.Itoa(int(subnetID)),
				Network:    "2.0.0.0/24",
				Location:   "Home",
				CustomerID: strconv.Itoa(int(customerID2)),
			},
		},
		Devices: []domain.Device{
			{
				// ID intentionally omitted
				IP:       "2.0.0.1",
				SubnetID: strconv.Itoa(int(subnetID)),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresConfigComponent := &sqldb.PostgresConfigComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresConfigComponent, db))
	defer db.Conn().Close()

	storer := &assetstorer.PostgresPhysicalAssetStorer{DB: db}
	err = storer.StorePhysicalAssets(ctx, ipamData)
	require.Nil(t, err)

	fetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: db}
	asset, err := fetcher.FetchPhysicalAsset(ctx, "2.0.0.1")
	require.Nil(t, err)

	expected := domain.PhysicalAsset{
		IP:            "2.0.0.1",
		ResourceOwner: "bob@example.com",
		BusinessUnit:  "Team Example",
		Network:       "2.0.0.0/24",
		Location:      "Home",
		DeviceID:      int64(0), // zero value expected
		SubnetID:      int64(subnetID),
		CustomerID:    int64(customerID2),
	}

	require.Equal(t, expected, asset)
}

// TestOverlappingSubnet verifies that a query for an IP address will return the
// most specific subnet that matches, as measured by the subnet's netmask length
func TestOverlappingSubnet(t *testing.T) {
	t.Parallel()

	customerID := rand.Int31()
	subnetID1 := rand.Int31()
	subnetID2 := rand.Int31()

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.Itoa(int(customerID)),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         strconv.Itoa(int(subnetID1)),
				Network:    "3.0.0.0/24",
				Location:   "Home",
				CustomerID: strconv.Itoa(int(customerID)),
			},
			{
				ID:         strconv.Itoa(int(subnetID2)),
				Network:    "3.0.0.252/30",
				Location:   "Home",
				CustomerID: strconv.Itoa(int(customerID)),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresConfigComponent := &sqldb.PostgresConfigComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresConfigComponent, db))
	defer db.Conn().Close()

	storer := &assetstorer.PostgresPhysicalAssetStorer{DB: db}
	err = storer.StorePhysicalAssets(context.Background(), ipamData)
	require.Nil(t, err)

	fetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: db}
	asset, err := fetcher.FetchPhysicalAsset(context.Background(), "3.0.0.253")
	require.Nil(t, err)

	expected := domain.PhysicalAsset{
		IP:            "3.0.0.253",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Example Team",
		Network:       "3.0.0.252/30",
		Location:      "Home",
		DeviceID:      0,
		SubnetID:      int64(subnetID2),
		CustomerID:    int64(customerID),
	}

	require.Equal(t, expected, asset)
}

// TestOverlappingSubnetWithDevice verifies that a query for an IP address will
// return the subnet associated with an existing device, even if that subnet is
// not the most subnet that contains the given IP address
func TestOverlappingSubnetWithDevice(t *testing.T) {
	t.Parallel()

	customerID := rand.Int31()
	subnetID1 := rand.Int31()
	subnetID2 := rand.Int31()
	deviceID := rand.Int31()

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.Itoa(int(customerID)),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         strconv.Itoa(int(subnetID1)),
				Network:    "4.0.0.0/24",
				Location:   "Home",
				CustomerID: strconv.Itoa(int(customerID)),
			},
			{
				ID:         strconv.Itoa(int(subnetID2)),
				Network:    "4.0.0.252/30",
				Location:   "Home - Den",
				CustomerID: strconv.Itoa(int(customerID)),
			},
		},
		Devices: []domain.Device{
			{
				ID:       strconv.Itoa(int(deviceID)),
				IP:       "4.0.0.253",
				SubnetID: strconv.Itoa(int(subnetID1)),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresConfigComponent := &sqldb.PostgresConfigComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresConfigComponent, db))
	defer db.Conn().Close()

	storer := &assetstorer.PostgresPhysicalAssetStorer{DB: db}
	err = storer.StorePhysicalAssets(context.Background(), ipamData)
	require.Nil(t, err)

	fetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: db}
	asset, err := fetcher.FetchPhysicalAsset(context.Background(), "4.0.0.253")
	require.Nil(t, err)

	expected := domain.PhysicalAsset{
		IP:            "4.0.0.253",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Example Team",
		Network:       "4.0.0.0/24",
		Location:      "Home",
		DeviceID:      int64(deviceID),
		SubnetID:      int64(subnetID1),
		CustomerID:    int64(customerID),
	}

	require.Equal(t, expected, asset)
}

// returns a raw sql.DB object, rather than the storage.DB abstraction, so
// we can perform some Postgres cleanup/prep/checks that are test-specific
func connectToDB(dbname string) (*sql.DB, error) {
	host := os.Getenv("POSTGRES_HOSTNAME")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USERNAME")
	password := os.Getenv("POSTGRES_PASSWORD")
	if dbname == "" {
		dbname = os.Getenv("POSTGRES_DATABASENAME")
	}

	sslmode := "disable"
	if host != "localhost" && host != "postgres" {
		sslmode = "require"
	}
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	pgdb, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = pgdb.Ping()
	if err != nil {
		return nil, err
	}

	return pgdb, nil
}

// wipeDatabase is a utility function to drop a database
func wipeDatabase(db *sql.DB, dbName string) error {
	sqlFile := "0_wipe.sql"

	box := packr.New("box", "../scripts")
	_, err := box.Find(sqlFile)
	if err != nil {
		return err
	}
	s, err := box.FindString(sqlFile)
	if err != nil {
		return err
	}

	if _, err = db.Exec(fmt.Sprintf(s, dbName)); err != nil {
		if driverErr, ok := err.(*pq.Error); ok {
			if strings.EqualFold(driverErr.Code.Name(), "invalid_catalog_name") { // from https://www.postgresql.org/docs/11/errcodes-appendix.html
				// it's ok the DB does not exist; this might by the very first run
				return nil
			}
		}
		return err
	}

	return nil
}
