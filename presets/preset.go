// Package presets contains a collection of preset [github.com/ygrebnov/testutils/docker.Container] objects.
package presets // import "github.com/ygrebnov/testutils/presets"

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/ygrebnov/testutils/docker"
)

// preset represents a type capable of producing preset and customizable [docker.Container] objects.
type preset[T any] interface {
	asContainer() T
	asCustomizedContainer(options docker.Options) T
}

// nolint: unused
type containerPreset = preset[docker.Container]

// defaultContainerPreset is a default implementation of `preset` interface.
// defaultContainerPreset holds container and image data.
type defaultContainerPreset struct {
	Container presetContainer `yaml:"container"`
	Image     presetImage     `yaml:"image"`
}

// presetContainer holds preset container data.
type presetContainer struct {
	Name        string               `yaml:"name"`
	Env         []presetContainerEnv `yaml:"env,omitempty"`
	Ports       []string             `yaml:"ports,omitempty"`
	Healthcheck string               `yaml:"healthcheck"`
}

// presetContainerEnv holds preset container environment variables data.
type presetContainerEnv struct {
	Name  string `yaml:"name"`
	Value any    `yaml:"value"`
}

// presetImage holds preset container image data.
type presetImage struct {
	Name string `yaml:"name"`
}

// newContainerPreset creates a new `containerPreset` object.
// nolint: unused
func newContainerPreset(valuesFile string) containerPreset {
	p := new(defaultContainerPreset)
	parsePresetValues(valuesFile, p)
	return p
}

// parsePresetValues sets given `preset` object attributes with values from the given yaml file.
func parsePresetValues(valuesFile string, preset any) {
	_, currFile, _, ok := runtime.Caller(0)
	if !ok {
		panic(errors.New("cannot locate preset values file"))
	}
	valuesFilePath := filepath.Join(filepath.Dir(currFile), valuesFile)
	valuesData, err := os.ReadFile(valuesFilePath)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(valuesData, preset); err != nil {
		panic(err)
	}
}

// asContainer returns a [docker.Container] object with preset attribute values.
// nolint: unused
func (p *defaultContainerPreset) asContainer() docker.Container {
	return docker.NewContainerWithOptions(p.Image.Name, p.getPresetContainerOptions())
}

// asCustomizedContainer returns a [docker.Container] with preset attribute values overwritten by customized ones.
// nolint: unused
func (p *defaultContainerPreset) asCustomizedContainer(options docker.Options) docker.Container {
	return docker.NewContainerWithOptions(p.Image.Name, p.combineContainerOptions(options))
}

// getPresetContainerOptions returns a [docker.Options] object with attributes values from preset yaml file.
// nolint: unused
func (p *defaultContainerPreset) getPresetContainerOptions() docker.Options {
	env := make([]string, 0, len(p.Container.Env))
	for _, el := range p.Container.Env {
		var stringVal string
		switch typedVal := el.Value.(type) {
		case int:
			stringVal = strconv.Itoa(typedVal)
		case string:
			stringVal = typedVal
		default:
			panic(errors.New("unhandled preset.env value type"))
		}
		env = append(env, fmt.Sprintf("%s=%s", el.Name, stringVal))
	}
	return docker.Options{
		Name:                 p.Container.Name,
		Healthcheck:          p.Container.Healthcheck,
		EnvironmentVariables: env,
		ExposedPorts:         p.Container.Ports,
	}
}

// nolint: unused
func (p *defaultContainerPreset) combineContainerOptions(options docker.Options) docker.Options {
	combinedOptions := p.getPresetContainerOptions()
	if len(options.Name) > 0 {
		combinedOptions.Name = options.Name
	}
	if len(options.EnvironmentVariables) > 0 {
		combinedOptions.EnvironmentVariables = options.EnvironmentVariables
	}
	if len(options.ExposedPorts) > 0 {
		combinedOptions.ExposedPorts = options.ExposedPorts
	}
	if len(options.Healthcheck) > 0 {
		combinedOptions.Healthcheck = options.Healthcheck
	}
	if options.StartTimeout > 0 {
		combinedOptions.StartTimeout = options.StartTimeout
	}
	return combinedOptions
}
