package presets

import "github.com/ygrebnov/testutils/docker"

type databaseContainerPreset = preset[docker.DatabaseContainer]

type defaultDatabaseContainerPreset struct {
	defaultContainerPreset `yaml:",inline"`
	Database               presetDatabase `yaml:"database"`
}

// presetDatabase holds database preset inner database data.
type presetDatabase struct {
	Name         string `yaml:"name"`
	ResetCommand string `yaml:"reset_command"`
}

// asContainer returns a [docker.Container] object with preset attribute values.
// nolint: unused
func (p *defaultDatabaseContainerPreset) asContainer() docker.DatabaseContainer {
	return docker.NewDatabaseContainerWithOptions(p.Image.Name, p.getPresetDatabase(), p.getPresetContainerOptions())
}

// asCustomizedContainer returns a [docker.Container] with preset attribute values overwritten by customized ones.
// nolint: unused
func (p *defaultDatabaseContainerPreset) asCustomizedContainer(options docker.Options) docker.DatabaseContainer {
	return docker.NewDatabaseContainerWithOptions(p.Image.Name, p.getPresetDatabase(), p.combineContainerOptions(options))
}

// nolint: unused
func (p *defaultDatabaseContainerPreset) getPresetDatabase() docker.Database {
	return docker.Database{Name: p.Database.Name, ResetCommand: p.Database.ResetCommand}
}

// newDatabaseContainerPreset creates a new `databaseContainerPreset` object.
func newDatabaseContainerPreset(valuesFile string) databaseContainerPreset {
	p := new(defaultDatabaseContainerPreset)
	parsePresetValues(valuesFile, p)
	return p
}
