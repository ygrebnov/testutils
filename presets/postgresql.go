package presets

import "github.com/ygrebnov/testutils/docker"

var postgresqlPreset = newPreset("values/postgresql.yml")

// NewCustomizedPostgresqlContainer returns a preset [github.com/ygrebnov/testutils/docker.Container] object with
// customized options values.
func NewCustomizedPostgresqlContainer(options docker.Options) docker.Container {
	return postgresqlPreset.asCustomizedContainer(options)
}

// NewPostgresqlContainer returns a preset [github.com/ygrebnov/testutils/docker.Container] object.
func NewPostgresqlContainer() docker.Container {
	return postgresqlPreset.asContainer()
}
