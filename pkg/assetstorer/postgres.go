package assetstorer

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
	"github.com/pkg/errors"
)

const (
	insertCustomerStatement = `INSERT INTO customers VALUES ($1, $2, $3)`
	insertSubnetStatement   = `INSERT INTO subnets VALUES ($1, $2, $3, $4)`
	insertIPStatement       = `INSERT INTO ips VALUES (DEFAULT, $1, $2, $3)`
	clearCustomerStatement  = `DELETE FROM customers`
	clearSubnetStatement    = `DELETE FROM subnets`
	clearIPStatement        = `DELETE FROM ips`
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

	err = clearTables(ctx, tx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Wrap(rollbackErr, err.Error())
		}
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
		if err := s.storeIP(ctx, device, tx); err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresPhysicalAssetStorer) storeCustomer(ctx context.Context, customer domain.Customer, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, insertCustomerStatement, customer.ID, customer.ResourceOwner, customer.BusinessUnit); err != nil {
		return err
	}

	return nil
}

func (s *PostgresPhysicalAssetStorer) storeSubnet(ctx context.Context, subnet domain.Subnet, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, insertSubnetStatement, subnet.ID, fmt.Sprintf("%s/%d", subnet.Network, subnet.MaskBits), subnet.Location, newNullString(subnet.CustomerID)); err != nil {
		return err
	}

	return nil
}

func (s *PostgresPhysicalAssetStorer) storeIP(ctx context.Context, device domain.Device, tx *sql.Tx) error {
	var deviceID *string
	if device.ID != "" {
		deviceID = &device.ID
	}
	if _, err := tx.ExecContext(ctx, insertIPStatement, device.IP, device.SubnetID, deviceID); err != nil {
		return err
	}

	return nil
}

func clearTables(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, clearCustomerStatement); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, clearSubnetStatement); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, clearIPStatement); err != nil {
		return err
	}

	return nil
}

func newNullString(s string) sql.NullString {
	if len(s) == 0 || s == "0" {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
