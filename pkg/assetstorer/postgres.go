package assetstorer

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/pkg/errors"
)

const (
	tableCustomers = "customers"
	tableSubnets   = "subnets"
	tableDevices   = "devices"
)

// PostgresPhysicalAssetStorer stores physical assets in a PostgreSQL database.
type PostgresPhysicalAssetStorer struct {
	DB domain.SQLDB
}

// StorePhysicalAssets stores physical asset device, subnet, and customer data in a a PostgreSQL database.
func (s *PostgresPhysicalAssetStorer) StorePhysicalAssets(ctx context.Context, ipamData domain.IPAMData) error {
	tx, err := s.DB.Conn().BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = s.savePhysicalAssets(ctx, ipamData, tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(rollbackErr, err.Error())
		}
		return err
	}
	return tx.Commit()
}

func (s *PostgresPhysicalAssetStorer) savePhysicalAssets(ctx context.Context, ipamData domain.IPAMData, tx *sql.Tx) error {
	for _, customer := range ipamData.Customers {
		if err := s.storeCustomer(ctx, customer, tx); err != nil {
			return err
		}
	}

	for _, subnet := range ipamData.Subnets {
		if err := s.storeSubnet(ctx, subnet, tx); err != nil {
			return err
		}
	}

	for _, device := range ipamData.Devices {
		if err := s.storeDevice(ctx, device, tx); err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresPhysicalAssetStorer) storeDevice(ctx context.Context, device domain.Device, tx *sql.Tx) error {
	// nolint
	stmt := fmt.Sprintf(`INSERT INTO %s VALUES ($1, $2, $3)`, tableDevices)

	if _, err := tx.ExecContext(ctx, stmt, device.ID, device.IP, device.SubnetID); err != nil {
		return err
	}

	return nil
}

func (s *PostgresPhysicalAssetStorer) storeSubnet(ctx context.Context, subnet domain.Subnet, tx *sql.Tx) error {
	// nolint
	stmt := fmt.Sprintf(`INSERT INTO %s VALUES ($1, $2, $3, $4)`, tableSubnets)

	if _, err := tx.ExecContext(ctx, stmt, subnet.ID, subnet.Network, subnet.Location, subnet.CustomerID); err != nil {
		return err
	}

	return nil
}

func (s *PostgresPhysicalAssetStorer) storeCustomer(ctx context.Context, customer domain.Customer, tx *sql.Tx) error {
	// nolint
	stmt := fmt.Sprintf(`INSERT INTO %s VALUES ($1, $2, $3)`, tableCustomers)

	if _, err := tx.ExecContext(ctx, stmt, customer.ID, customer.ResourceOwner, customer.BusinessUnit); err != nil {
		return err
	}

	return nil
}