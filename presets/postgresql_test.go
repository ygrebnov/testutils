package presets

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ygrebnov/testutils/docker"
)

func Test_PostgresqlPreset(t *testing.T) {
	expectedContainerBuilder := docker.NewContainerBuilder("postgresqlPresetContainer", "postgres").
		SetEnv([]string{"POSTGRES_USER=postgres", "POSTGRES_PASSWORD=postgresqlPresetPassword", "PGPORT=5432"}).
		ExposePorts([]string{"5432:5432"}).
		Healthcheck("pg_isready")
	expectedContainer := expectedContainerBuilder.Build()

	require.Equal(t, expectedContainerBuilder, NewPostgresqlContainerBuilder())
	require.Equal(t, expectedContainer, NewPostgresqlContainer())
}

func Test_PostgresqlPresetContainerBuilderUniqueness(t *testing.T) {
	p1 := NewPostgresqlContainerBuilder()
	p2 := NewPostgresqlContainerBuilder()
	p1.ExposePorts([]string{"5435:5435"})
	require.NotEqual(t, p1, p2)
}
