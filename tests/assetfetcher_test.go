// +build integration

package inttest

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	packr "github.com/gobuffalo/packr/v2"
	pq "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asecurityteam/ipam-facade/pkg/assetfetcher"
	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/asecurityteam/ipam-facade/pkg/sqldb"
	"github.com/asecurityteam/settings"
)

var ctx context.Context
var conn *sql.DB
var db *sqldb.PostgresDB
var fetcher assetfetcher.PostgresPhysicalAssetFetcher

func TestMain(m *testing.M) {

	// wipe the database entirely, which will result in testing DB.Init
	// handling of lack of pre-existing database
	sslmode := "disable"
	host := os.Getenv("POSTGRES_HOSTNAME")
	if host != "localhost" && host != "postgres" {
		sslmode = "require"
	}
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USERNAME"), os.Getenv("POSTGRES_PASSWORD"), "postgres", sslmode)
	pgdb, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err.Error())
	}
	defer pgdb.Close()

	if err = wipeDatabase(pgdb, os.Getenv("POSTGRES_DATABASENAME")); err != nil {
		panic(err.Error())
	}

	ctx = context.Background()
	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}

	postgresConfigComponent := &sqldb.PostgresConfigComponent{}
	db = new(sqldb.PostgresDB)
	if err = settings.NewComponent(ctx, source, postgresConfigComponent, db); err != nil {
		panic(err.Error())
	}

	conn, err = connectToDB()
	if err != nil {
		panic(err.Error())
	}

	fetcher = assetfetcher.PostgresPhysicalAssetFetcher{DB: db}

	os.Exit(m.Run())
}

// TestNoDBRows verifies that sql.ErrNoRows is returned when no devices or subnets exist
func TestNoDBRows(t *testing.T) {
	before(t, db)

	// code should tolerate no data in the tables
	_, err := fetcher.FetchPhysicalAsset(ctx, "127.0.0.1")
	assert.Equal(t, domain.AssetNotFound{Inner: sql.ErrNoRows, IP: "127.0.0.1"}, err)
}

// TestSubnetOnly verifies that an IP address within a subnet will return a match, even when
// no corresponding device exists
func TestSubnetOnly(t *testing.T) {
	before(t, db)

	insertCustomer(t, db, 1, "alice@example.com", "Example Team")
	insertSubnet(t, db, 1, "127.0.0.0/24", "Home", 1)

	asset, err := fetcher.FetchPhysicalAsset(ctx, "127.0.0.1")
	if err != nil {
		t.Fatal(err.Error())
	}

	expected := domain.PhysicalAsset{
		IP:            "127.0.0.1",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Example Team",
		Network:       "127.0.0.0/24",
		Location:      "Home",
		DeviceID:      0,
		SubnetID:      1,
		CustomerID:    1,
	}

	assert.Equal(t, expected, asset)
}

// TestDeviceAndSubnet verifies that a query for an IP address with a device match
// returns both device and subnet information
func TestDeviceAndSubnet(t *testing.T) {
	before(t, db)

	insertCustomer(t, db, 1, "alice@example.com", "Example Team")
	insertCustomer(t, db, 2, "bob@example.com", "Team Example")
	insertSubnet(t, db, 1, "127.0.0.0/24", "Home", 2)
	insertDevice(t, db, 1, "127.0.0.1", 1)

	asset, err := fetcher.FetchPhysicalAsset(ctx, "127.0.0.1")
	if err != nil {
		t.Fatal(err.Error())
	}

	expected := domain.PhysicalAsset{
		IP:            "127.0.0.1",
		ResourceOwner: "bob@example.com",
		BusinessUnit:  "Team Example",
		Network:       "127.0.0.0/24",
		Location:      "Home",
		DeviceID:      1,
		SubnetID:      1,
		CustomerID:    2,
	}

	assert.Equal(t, expected, asset)
}

// TestOverlappingSubnet verifies that a query for an IP address will return the
// most specific subnet that matches, as measured by the subnet's netmask length
func TestOverlappingSubnet(t *testing.T) {
	before(t, db)

	insertCustomer(t, db, 1, "alice@example.com", "Example Team")
	insertSubnet(t, db, 1, "127.0.0.0/24", "Home", 1)
	insertSubnet(t, db, 2, "127.0.0.252/30", "Home", 1)

	asset, err := fetcher.FetchPhysicalAsset(ctx, "127.0.0.253")
	if err != nil {
		t.Fatal(err.Error())
	}

	expected := domain.PhysicalAsset{
		IP:            "127.0.0.253",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Example Team",
		Network:       "127.0.0.252/30",
		Location:      "Home",
		DeviceID:      0,
		SubnetID:      2,
		CustomerID:    1,
	}

	assert.Equal(t, expected, asset)
}

// TestOverlappingSubnetWithDevice verifies that a query for an IP address will
// return the subnet associated with an existing device, even if that subnet is
// not the most subnet that contains the given IP address
func TestOverlappingSubnetWithDevice(t *testing.T) {
	before(t, db)

	insertCustomer(t, db, 2, "alice@example.com", "Example Team")
	insertSubnet(t, db, 2, "127.0.0.0/24", "Home", 2)
	insertSubnet(t, db, 3, "127.0.0.252/30", "Home - Den", 2)
	insertDevice(t, db, 2, "127.0.0.253", 2)

	asset, err := fetcher.FetchPhysicalAsset(ctx, "127.0.0.253")
	if err != nil {
		t.Fatal(err.Error())
	}

	expected := domain.PhysicalAsset{
		IP:            "127.0.0.253",
		ResourceOwner: "alice@example.com",
		BusinessUnit:  "Example Team",
		Network:       "127.0.0.0/24",
		Location:      "Home",
		DeviceID:      2,
		SubnetID:      2,
		CustomerID:    2,
	}

	assert.Equal(t, expected, asset)
}

// returns a raw sql.DB object, rather than the storage.DB abstraction, so
// we can perform some Postgres cleanup/prep/checks that are test-specific
func connectToDB() (*sql.DB, error) {
	host := os.Getenv("POSTGRES_HOSTNAME")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USERNAME")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DATABASENAME")

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

// before is the function all tests should call to ensure no state is carried over
// from prior tests
func before(t *testing.T, db *sqldb.PostgresDB) {
	require.NoError(t, db.RunScript(context.Background(), "1_clean.sql"))
	require.NoError(t, db.RunScript(context.Background(), "2_create.sql"))
}

// dropTables is a utility function called by "before"
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

func insertCustomer(t *testing.T, db *sqldb.PostgresDB, id int, resourceOwner string, businessUnit string) {
	_, err := db.Conn().ExecContext(
		ctx, "INSERT INTO customers (id, resource_owner, business_unit) VALUES ($1, $2, $3)",
		id, resourceOwner, businessUnit)
	require.Nil(t, err)
}

func insertSubnet(t *testing.T, db *sqldb.PostgresDB, id int, network, location string, customerID int) {
	_, err := db.Conn().ExecContext(
		ctx, "INSERT INTO subnets (id, network, location, customer_id) VALUES ($1, $2, $3, $4)",
		id, network, location, customerID)
	require.Nil(t, err)
}

func insertDevice(t *testing.T, db *sqldb.PostgresDB, id int, ip string, subnetID int) {
	_, err := db.Conn().ExecContext(
		ctx, "INSERT INTO devices (id, ip, subnet_id) VALUES ($1, $2, $3)", id, ip, subnetID)
	require.Nil(t, err)
}
