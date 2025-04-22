package presets

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ygrebnov/testutils/docker"
)

func TestDefaultDatabaseContainerPreset_AsContainer(t *testing.T) {
	p := &defaultDatabaseContainerPreset{
		defaultContainerPreset: defaultContainerPreset{
			Image: presetImage{Name: "postgres"},
		},
		Database: presetDatabase{
			Name:         "testdb",
			ResetCommand: "dropdb -f testdb; createdb testdb",
		},
	}

	expectedContainer := docker.NewDatabaseContainerWithOptions(
		"postgres",
		docker.Database{
			Name:         "testdb",
			ResetCommand: "dropdb -f testdb; createdb testdb",
		},
		&docker.Options{
			StartTimeout: 60,
		},
	)

	actualContainer := p.asContainer()

	require.Equal(t, expectedContainer, actualContainer)
}
