//+build integration

package inttest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"

	packr "github.com/gobuffalo/packr/v2"
	pq "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/asecurityteam/ipam-facade/pkg/assetfetcher"
	"github.com/asecurityteam/ipam-facade/pkg/assetstorer"
	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/asecurityteam/ipam-facade/pkg/sqldb"
	"github.com/asecurityteam/settings"
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
	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresComponent := &sqldb.PostgresComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresComponent, db))

	defer func() {
		if dbErr := db.Conn().Close(); dbErr != nil {
			fmt.Println("Error when closing:", dbErr)
		}
	}()

	// code should tolerate no data in the tables
	fetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: db}
	_, err = fetcher.FetchPhysicalAsset(context.Background(), "0.0.0.0")
	require.Equal(t, domain.AssetNotFound{Inner: sql.ErrNoRows, IP: "0.0.0.0"}, err)
}

// TestSubnetOnly verifies that an IP address within a subnet will return a match, even when
// no corresponding device exists
func TestSubnetOnly(t *testing.T) {
	customerID, _ := rand.Int(rand.Reader, big.NewInt(1000))
	subnetID, _ := rand.Int(rand.Reader, big.NewInt(1000))

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.FormatInt(customerID.Int64(), 10),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         strconv.FormatInt(subnetID.Int64(), 10),
				Network:    "1.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID.Int64(), 10),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresComponent := &sqldb.PostgresComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresComponent, db))
	defer func() {
		if dbErr := db.Conn().Close(); dbErr != nil {
			fmt.Println("Error when closing:", dbErr)
		}
	}()

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
		SubnetID:      subnetID.Int64(),
		CustomerID:    customerID.Int64(),
	}

	require.Equal(t, expected, asset)
}

// TestDeviceAndSubnet verifies that a query for an IP address with a device match
// returns both device and subnet information
func TestDeviceAndSubnet(t *testing.T) {
	customerID1, _ := rand.Int(rand.Reader, big.NewInt(1000))
	customerID2, _ := rand.Int(rand.Reader, big.NewInt(1000))
	subnetID, _ := rand.Int(rand.Reader, big.NewInt(1000))
	deviceID, _ := rand.Int(rand.Reader, big.NewInt(1000))

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.FormatInt(customerID1.Int64(), 10),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
			{
				ID:            strconv.FormatInt(customerID2.Int64(), 10),
				ResourceOwner: "bob@example.com",
				BusinessUnit:  "Team Example",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         strconv.FormatInt(subnetID.Int64(), 10),
				Network:    "2.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID2.Int64(), 10),
			},
		},
		Devices: []domain.Device{
			{
				ID:       strconv.FormatInt(deviceID.Int64(), 10),
				IP:       "2.0.0.1",
				SubnetID: strconv.FormatInt(subnetID.Int64(), 10),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresComponent := &sqldb.PostgresComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresComponent, db))
	defer func() {
		if dbErr := db.Conn().Close(); dbErr != nil {
			fmt.Println("Error when closing:", dbErr)
		}
	}()

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
		DeviceID:      deviceID.Int64(),
		SubnetID:      subnetID.Int64(),
		CustomerID:    customerID2.Int64(),
	}

	require.Equal(t, expected, asset)
}

// TestDeviceAndSubnet verifies that a query for an IP address with a device match
// returns both device and subnet information
func TestDeviceAndSubnetNoDeviceID(t *testing.T) {
	// I don't know if IPAM would ever return device info where the
	// device lacks an ID, but we're gonna handle it if it does...
	customerID1, _ := rand.Int(rand.Reader, big.NewInt(1000))
	customerID2, _ := rand.Int(rand.Reader, big.NewInt(1000))
	subnetID, _ := rand.Int(rand.Reader, big.NewInt(1000))

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.FormatInt(customerID1.Int64(), 10),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
			{
				ID:            strconv.FormatInt(customerID2.Int64(), 10),
				ResourceOwner: "bob@example.com",
				BusinessUnit:  "Team Example",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         strconv.FormatInt(subnetID.Int64(), 10),
				Network:    "2.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID2.Int64(), 10),
			},
		},
		Devices: []domain.Device{
			{
				// ID intentionally omitted
				IP:       "2.0.0.1",
				SubnetID: strconv.FormatInt(subnetID.Int64(), 10),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresComponent := &sqldb.PostgresComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresComponent, db))
	defer func() {
		if dbErr := db.Conn().Close(); dbErr != nil {
			fmt.Println("Error when closing:", dbErr)
		}
	}()

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
		SubnetID:      subnetID.Int64(),
		CustomerID:    customerID2.Int64(),
	}

	require.Equal(t, expected, asset)
}

// TestOverlappingSubnet verifies that a query for an IP address will return the
// most specific subnet that matches, as measured by the subnet's netmask length
func TestOverlappingSubnet(t *testing.T) {
	customerID, _ := rand.Int(rand.Reader, big.NewInt(1000))
	subnetID1, _ := rand.Int(rand.Reader, big.NewInt(1000))
	subnetID2, _ := rand.Int(rand.Reader, big.NewInt(1000))

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.FormatInt(customerID.Int64(), 10),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         strconv.FormatInt(subnetID1.Int64(), 10),
				Network:    "3.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID.Int64(), 10),
			},
			{
				ID:         strconv.FormatInt(subnetID2.Int64(), 10),
				Network:    "3.0.0.252",
				MaskBits:   30,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID.Int64(), 10),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresComponent := &sqldb.PostgresComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresComponent, db))
	defer func() {
		if dbErr := db.Conn().Close(); dbErr != nil {
			fmt.Println("Error when closing:", dbErr)
		}
	}()

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
		SubnetID:      subnetID2.Int64(),
		CustomerID:    customerID.Int64(),
	}

	require.Equal(t, expected, asset)
}

// TestOverlappingSubnetWithDevice verifies that a query for an IP address will
// return the subnet associated with an existing device, even if that subnet is
// not the most subnet that contains the given IP address
func TestOverlappingSubnetWithDevice(t *testing.T) {
	customerID, _ := rand.Int(rand.Reader, big.NewInt(1000))
	subnetID1, _ := rand.Int(rand.Reader, big.NewInt(1000))
	subnetID2, _ := rand.Int(rand.Reader, big.NewInt(1000))
	deviceID, _ := rand.Int(rand.Reader, big.NewInt(1000))

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.FormatInt(customerID.Int64(), 10),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         strconv.FormatInt(subnetID1.Int64(), 10),
				Network:    "4.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID.Int64(), 10),
			},
			{
				ID:         strconv.FormatInt(subnetID2.Int64(), 10),
				Network:    "4.0.0.252",
				MaskBits:   30,
				Location:   "Home - Den",
				CustomerID: strconv.FormatInt(customerID.Int64(), 10),
			},
		},
		Devices: []domain.Device{
			{
				ID:       strconv.FormatInt(deviceID.Int64(), 10),
				IP:       "4.0.0.253",
				SubnetID: strconv.FormatInt(subnetID1.Int64(), 10),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresComponent := &sqldb.PostgresComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresComponent, db))
	defer func() {
		if dbErr := db.Conn().Close(); dbErr != nil {
			fmt.Println("Error when closing:", dbErr)
		}
	}()
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
		DeviceID:      deviceID.Int64(),
		SubnetID:      subnetID1.Int64(),
		CustomerID:    customerID.Int64(),
	}

	require.Equal(t, expected, asset)
}

func TestFetchSubnet(t *testing.T) {
	customerID1, _ := rand.Int(rand.Reader, big.NewInt(1000))
	customerID2, _ := rand.Int(rand.Reader, big.NewInt(1000))

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            strconv.FormatInt(customerID1.Int64(), 10),
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
			{
				ID:            strconv.FormatInt(customerID2.Int64(), 10),
				ResourceOwner: "bob@example.com",
				BusinessUnit:  "Team Example",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         "10",
				Network:    "1.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID1.Int64(), 10),
			},
			{
				ID:         "11",
				Network:    "2.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID2.Int64(), 10),
			},
			{
				ID:         "12",
				Network:    "3.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID1.Int64(), 10),
			},
			{
				ID:         "13",
				Network:    "4.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID2.Int64(), 10),
			},
			{
				ID:         "14",
				Network:    "5.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID1.Int64(), 10),
			},
			{
				ID:         "15",
				Network:    "6.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID2.Int64(), 10),
			},
			{
				ID:         "16",
				Network:    "7.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID1.Int64(), 10),
			},
			{
				ID:         "17",
				Network:    "8.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID2.Int64(), 10),
			},
			{
				ID:         "18",
				Network:    "9.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID1.Int64(), 10),
			},
			{
				ID:         "19",
				Network:    "10.0.0.0",
				MaskBits:   24,
				Location:   "Home",
				CustomerID: strconv.FormatInt(customerID2.Int64(), 10),
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresComponent := &sqldb.PostgresComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresComponent, db))
	defer func() {
		if dbErr := db.Conn().Close(); dbErr != nil {
			fmt.Println("Error when closing:", dbErr)
		}
	}()
	storer := &assetstorer.PostgresPhysicalAssetStorer{DB: db}
	err = storer.StorePhysicalAssets(ctx, ipamData)
	require.Nil(t, err)

	fetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: db}

	limit := 2
	for offset := 0; offset < len(ipamData.Subnets); offset += limit {
		asset, err := fetcher.FetchSubnets(ctx, limit, offset)
		require.Nil(t, err)

		expected := []domain.AssetSubnet{
			{
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
				Network:       fmt.Sprintf("%d.0.0.0/24", offset+1),
				Location:      "Home",
			},
			{
				ResourceOwner: "bob@example.com",
				BusinessUnit:  "Team Example",
				Network:       fmt.Sprintf("%d.0.0.0/24", offset+2),
				Location:      "Home",
			},
		}

		require.Equal(t, expected, asset)
	}
}

func TestFetchIPs(t *testing.T) {

	ipamData := domain.IPAMData{
		Customers: []domain.Customer{
			{
				ID:            "1",
				ResourceOwner: "alice@example.com",
				BusinessUnit:  "Example Team",
			},
			{
				ID:            "2",
				ResourceOwner: "bob@example.com",
				BusinessUnit:  "Team Example",
			},
		},
		Subnets: []domain.Subnet{
			{
				ID:         "1",
				Network:    "1.0.0.0",
				MaskBits:   25,
				Location:   "Home",
				CustomerID: "1",
			},
			{
				ID:         "2",
				Network:    "1.0.0.128",
				MaskBits:   25,
				Location:   "Home",
				CustomerID: "2",
			},
		},
		Devices: []domain.Device{
			{
				ID:       "3",
				IP:       "1.0.0.1",
				SubnetID: "1",
			},
			{
				ID:       "4",
				IP:       "1.0.0.2",
				SubnetID: "1",
			},
			{
				ID:       "5",
				IP:       "1.0.0.3",
				SubnetID: "1",
			},
			{
				ID:       "6",
				IP:       "1.0.0.4",
				SubnetID: "1",
			},
			{
				ID:       "7",
				IP:       "1.0.0.129",
				SubnetID: "2",
			},
			{
				ID:       "8",
				IP:       "1.0.0.130",
				SubnetID: "2",
			},
			{
				ID:       "9",
				IP:       "1.0.0.131",
				SubnetID: "2",
			},
			{
				ID:       "10",
				IP:       "1.0.0.132",
				SubnetID: "2",
			},
		},
	}

	ctx := context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	require.Nil(t, err)

	postgresComponent := &sqldb.PostgresComponent{}
	db := new(sqldb.PostgresDB)
	require.Nil(t, settings.NewComponent(ctx, source, postgresComponent, db))
	defer func() {
		if dbErr := db.Conn().Close(); dbErr != nil {
			fmt.Println("Error when closing:", dbErr)
		}
	}()
	storer := &assetstorer.PostgresPhysicalAssetStorer{DB: db}
	err = storer.StorePhysicalAssets(ctx, ipamData)
	require.Nil(t, err)

	fetcher := &assetfetcher.PostgresPhysicalAssetFetcher{DB: db}

	limit := 2
	ips := make([]domain.AssetIP, 0, len(ipamData.Devices))
	for offset := 0; offset < len(ipamData.Devices); offset += limit {
		result, err := fetcher.FetchIPs(ctx, limit, offset)
		require.Nil(t, err)

		ips = append(ips, result...)
	}
	expected := []domain.AssetIP{
		{
			IP:            "1.0.0.1",
			Network:       "1.0.0.0/25",
			ResourceOwner: "alice@example.com",
			BusinessUnit:  "Example Team",
			Location:      "Home",
		},
		{
			IP:            "1.0.0.2",
			Network:       "1.0.0.0/25",
			ResourceOwner: "alice@example.com",
			BusinessUnit:  "Example Team",
			Location:      "Home",
		},
		{
			IP:            "1.0.0.3",
			Network:       "1.0.0.0/25",
			ResourceOwner: "alice@example.com",
			BusinessUnit:  "Example Team",
			Location:      "Home",
		},
		{
			IP:            "1.0.0.4",
			Network:       "1.0.0.0/25",
			ResourceOwner: "alice@example.com",
			BusinessUnit:  "Example Team",
			Location:      "Home",
		},
		{
			IP:            "1.0.0.129",
			Network:       "1.0.0.128/25",
			ResourceOwner: "bob@example.com",
			BusinessUnit:  "Team Example",
			Location:      "Home",
		},
		{
			IP:            "1.0.0.130",
			Network:       "1.0.0.128/25",
			ResourceOwner: "bob@example.com",
			BusinessUnit:  "Team Example",
			Location:      "Home",
		},
		{
			IP:            "1.0.0.131",
			Network:       "1.0.0.128/25",
			ResourceOwner: "bob@example.com",
			BusinessUnit:  "Team Example",
			Location:      "Home",
		},
		{
			IP:            "1.0.0.132",
			Network:       "1.0.0.128/25",
			ResourceOwner: "bob@example.com",
			BusinessUnit:  "Team Example",
			Location:      "Home",
		},
	}
	require.Equal(t, expected, ips)
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
