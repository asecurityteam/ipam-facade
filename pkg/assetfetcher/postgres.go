package assetfetcher

import (
	"context"
	"database/sql"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

const fetchByIPQuery = `SELECT host(d.ip) as ip, c.resource_owner as resource_owner,
							c.business_unit as business_unit, text(s.network) as network,
							s.location as location, d.id as device_id, s.id as subnet_id,
							c.id as customer_id
			            FROM devices d
					  	RIGHT OUTER JOIN subnets s ON
					  		d.subnet_id = s.id
				  		AND d.ip = $1
						INNER JOIN customers c ON s.customer_id = c.id
						WHERE s.network >>= $1
						ORDER BY d.id IS NOT NULL DESC, masklen(s.network) DESC
						LIMIT 1;`

// PostgresPhysicalAssetFetcher physical assets from a PostgreSQL database by IP address.
type PostgresPhysicalAssetFetcher struct {
	DB domain.SQLDB
}

// FetchPhysicalAsset queries the SQL DB for a physical asset by the given IP address.
func (f *PostgresPhysicalAssetFetcher) FetchPhysicalAsset(ctx context.Context, ipAddress string) (domain.PhysicalAsset, error) {
	var asset domain.PhysicalAsset
	var ip sql.NullString
	var deviceID sql.NullInt64
	err := f.DB.Conn().QueryRowContext(ctx, fetchByIPQuery, ipAddress).Scan(
		&ip, &asset.ResourceOwner, &asset.BusinessUnit, &asset.Network,
		&asset.Location, &deviceID, &asset.SubnetID, &asset.CustomerID)
	switch {
	case err == sql.ErrNoRows:
		return domain.PhysicalAsset{}, domain.AssetNotFound{Inner: err, IP: ipAddress}
	case err != nil:
		return domain.PhysicalAsset{}, err
	}

	if deviceID.Valid {
		asset.DeviceID = deviceID.Int64
		asset.IP = ip.String
	} else {
		asset.IP = ipAddress
		asset.DeviceID = 0
	}

	return asset, nil
}
