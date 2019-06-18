package assetstorer

import "github.com/asecurityteam/ipam-facade/pkg/domain"

// PostgresPhysicalAssetStorer stores physical assets in a PostgreSQL database.
type PostgresPhysicalAssetStorer struct {
	DB domain.SQLDB
}
