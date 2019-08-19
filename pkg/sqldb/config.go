package sqldb

import (
	"context"

	packr "github.com/gobuffalo/packr/v2"
)

// PostgresConfig contains the Postgres database configuration arguments
type PostgresConfig struct {
	Hostname     string
	Port         string
	Username     string
	Password     string
	DatabaseName string
}

// Name is used by the settings library to replace the default naming convention.
func (c *PostgresConfig) Name() string {
	return "Postgres"
}

// NewPostgresComponent generates a new, unitialized PostgresComponent
func NewPostgresComponent() *PostgresComponent {
	return &PostgresComponent{}
}

// PostgresComponent satisfies the settings library Component API,
// and may be used by the settings.NewComponent function.
type PostgresComponent struct{}

// Settings populates a set of defaults if none are provided via config.
func (*PostgresComponent) Settings() *PostgresConfig {
	return &PostgresConfig{}
}

// New constructs a DB from a config
func (*PostgresComponent) New(ctx context.Context, c *PostgresConfig) (*PostgresDB, error) {
	scripts := packr.New("scripts", "../../scripts")
	db := &PostgresDB{
		scripts: scripts.FindString,
	}
	if err := db.Init(ctx, c.Hostname, c.Port, c.Username, c.Password, c.DatabaseName); err != nil {
		return nil, err
	}
	return db, nil
}
