package sqldb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	postgresConfig := PostgresConfig{
		Hostname:     "localhost",
		Port:         "99",
		Username:     "me!",
		Password:     "mypassword!",
		DatabaseName: "name",
	}
	assert.Equal(t, "Postgres", postgresConfig.Name())
}

func TestShouldReturnSame(t *testing.T) {
	postgresComponent := NewPostgresComponent()
	postgresConfig := postgresComponent.Settings()
	assert.NotNil(t, postgresConfig)
	assert.Empty(t, postgresConfig.DatabaseName)
}

func TestShouldFailToMakeNewDB(t *testing.T) {
	postgresConfig := PostgresConfig{}

	postgresComponent := NewPostgresComponent()
	_, err := postgresComponent.New(context.Background(), &postgresConfig)
	assert.NotNil(t, err)
}
