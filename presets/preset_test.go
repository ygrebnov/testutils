package presets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewContainerPreset(t *testing.T) {
	// TODO: replace by on-the-fly file creation.
	valuesFile := "../testdata/container_preset.yaml"

	expectedPreset := &defaultContainerPreset{
		Container: presetContainer{
			Name:        "test-container",
			Env:         envs{{Name: "ENV_VAR", Value: "value"}},
			Ports:       []string{"8080:80"},
			Healthcheck: "curl -f http://localhost || exit 1",
		},
		Image: presetImage{
			Name: "test-image",
		},
	}

	actualPreset := newContainerPreset(valuesFile).(*defaultContainerPreset)

	require.Equal(t, expectedPreset, actualPreset)
}
