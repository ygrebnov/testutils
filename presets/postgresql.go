package presets

import "github.com/ygrebnov/testutils/docker"

var postgresqlPreset = newDatabaseContainerPreset("postgresql.yaml")

// NewCustomizedPostgresqlContainer returns a preset [github.com/ygrebnov/testutils/docker.DatabaseContainer] object with
// customized options values.
func NewCustomizedPostgresqlContainer(options docker.Options) docker.DatabaseContainer {
	return postgresqlPreset.asCustomizedContainer(options)
}

// NewPostgresqlContainer returns a preset [github.com/ygrebnov/testutils/docker.DatabaseContainer] object.
func NewPostgresqlContainer() docker.DatabaseContainer {
	return postgresqlPreset.asContainer()
}
