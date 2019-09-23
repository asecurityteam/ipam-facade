package assetfetcher

import (
	"context"
	"database/sql"

	"github.com/asecurityteam/ipam-facade/pkg/domain"
)

const fetchByIPQuery = `SELECT host(i.ip) as ip, c.resource_owner as resource_owner,
							c.business_unit as business_unit, text(s.network) as network,
							s.location as location, device_id, s.id as subnet_id,
							c.id as customer_id
			            FROM ips i
					  	RIGHT OUTER JOIN subnets s ON
					  		i.subnet_id = s.id
						AND i.ip = $1
						LEFT OUTER JOIN customers c ON s.customer_id = c.id
						WHERE s.network >>= $1
						ORDER BY i.device_id IS NOT NULL DESC, masklen(s.network) DESC
						LIMIT 1;`

const fetchSubnetsQuery = `SELECT network, location, resource_owner, business_unit
						FROM subnets
						LEFT JOIN customers ON
							subnets.customer_id=customers.id
						LIMIT $1 OFFSET $2;`

const fetchIPsQuery = `SELECT ip, network, location, resource_owner, business_unit
						FROM ips
						LEFT JOIN subnets ON
							ips.subnet_id=subnets.id
						LEFT JOIN customers ON
							subnets.customer_id=customers.id
						LIMIT $1 OFFSET $2;`

// PostgresPhysicalAssetFetcher physical assets from a PostgreSQL database by IP address.
type PostgresPhysicalAssetFetcher struct {
	DB domain.SQLDB
}

// FetchPhysicalAsset queries the SQL DB for a physical asset by the given IP address.
func (f *PostgresPhysicalAssetFetcher) FetchPhysicalAsset(ctx context.Context, ipAddress string) (domain.PhysicalAsset, error) {
	var asset domain.PhysicalAsset
	var ip sql.NullString
	var deviceID sql.NullInt64
	var assetResourceOwner sql.NullString
	var assetBusinessUnit sql.NullString
	var assetCustomerID sql.NullInt64
	err := f.DB.Conn().QueryRowContext(ctx, fetchByIPQuery, ipAddress).Scan(
		&ip, &assetResourceOwner, &assetBusinessUnit, &asset.Network,
		&asset.Location, &deviceID, &asset.SubnetID, &assetCustomerID)
	if assetCustomerID.Valid {
		// if we have a customerID from our query, we'll surely have the rest too:
		asset.CustomerID = assetCustomerID.Int64
		asset.ResourceOwner = assetResourceOwner.String
		asset.BusinessUnit = assetBusinessUnit.String
	}
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

// FetchSubnets fetches a single page of subnets from the data store
func (f *PostgresPhysicalAssetFetcher) FetchSubnets(ctx context.Context, limit, offset int) ([]domain.AssetSubnet, error) {
	rows, err := f.DB.Conn().QueryContext(ctx, fetchSubnetsQuery, limit, offset)
	if err != nil {
		return nil, err
	}

	subnets := make([]domain.AssetSubnet, 0, limit)
	for rows.Next() {
		var network string
		var location sql.NullString
		var resourceOwner sql.NullString
		var businessUnit sql.NullString
		if err := rows.Scan(&network, &location, &resourceOwner, &businessUnit); err != nil {
			// this would indicate an error in our schema or ordering of variables.
			// either case would be a terminal error, so we close the rows at best effort and return.
			_ = rows.Close()
			return nil, err
		}
		subnet := domain.AssetSubnet{
			Network: network,
		}
		if location.Valid {
			subnet.Location = location.String
		}
		if resourceOwner.Valid {
			subnet.ResourceOwner = resourceOwner.String
		}
		if businessUnit.Valid {
			subnet.BusinessUnit = businessUnit.String
		}
		subnets = append(subnets, subnet)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	return subnets, nil
}

// FetchIPs fetches a single page of IP addresses from the data store
func (f *PostgresPhysicalAssetFetcher) FetchIPs(ctx context.Context, limit, offset int) ([]domain.AssetIP, error) {
	rows, err := f.DB.Conn().QueryContext(ctx, fetchIPsQuery, limit, offset)
	if err != nil {
		return nil, err
	}

	ips := make([]domain.AssetIP, 0, limit)
	for rows.Next() {
		var ipAddr string
		var network string
		var location sql.NullString
		var resourceOwner sql.NullString
		var businessUnit sql.NullString
		if err := rows.Scan(&ipAddr, &network, &location, &resourceOwner, &businessUnit); err != nil {
			// this would indicate an error in our schema or ordering of variables.
			// either case would be a terminal error, so we close the rows at best effort and return.
			_ = rows.Close()
			return nil, err
		}
		ip := domain.AssetIP{
			IP:      ipAddr,
			Network: network,
		}
		if location.Valid {
			ip.Location = location.String
		}
		if resourceOwner.Valid {
			ip.ResourceOwner = resourceOwner.String
		}
		if businessUnit.Valid {
			ip.BusinessUnit = businessUnit.String
		}
		ips = append(ips, ip)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	return ips, nil
}
