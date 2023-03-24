package presets

import "github.com/ygrebnov/testutils/docker"

var postgresqlPreset = newPreset("values/postgresql.yml")

// NewPostgresqlContainerBuilder returns a [github.com/ygrebnov/testutils/docker.ContainerBuilder] object with
// the minimum required environmental variables created and the default 5432 port exported.
//
// List of created environment variables with assigned values:
//
// - POSTGRES_USER=postgres,
//
// - POSTGRES_PASSWORD=postgresqlPresetPassword,
//
// - PGPORT=5432.
//
// Before building a [github.com/ygrebnov/testutils/docker.Container] from an object returned by
// [NewPostgresqlContainerBuilder], the latter can be modified by calling
// [github.com/ygrebnov/testutils/docker.ContainerBuilder] SetEnv() and/or ExposePorts() methods.
func NewPostgresqlContainerBuilder() docker.ContainerBuilder {
	return postgresqlPreset.asContainerBuilder()
}

// NewPostgresqlContainer returns a built [github.com/ygrebnov/testutils/docker.Container] pre-configured like
// the object returned by [NewPostgresqlContainerBuilder].
func NewPostgresqlContainer() docker.Container {
	return postgresqlPreset.asContainer()
}
