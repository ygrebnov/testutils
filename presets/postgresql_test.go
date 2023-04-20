package presets

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygrebnov/testutils/docker"
)

func Test_PostgresqlPreset(t *testing.T) {
	expectedContainer := docker.NewContainerWithOptions(
		"postgres",
		docker.Options{
			Healthcheck:          "pg_isready",
			EnvironmentVariables: []string{"POSTGRES_USER=postgres", "POSTGRES_PASSWORD=postgres", "PGPORT=5432"},
			ExposedPorts:         []string{"5432:5432"},
		})

	require.Equal(t, expectedContainer, NewPostgresqlContainer())
}
