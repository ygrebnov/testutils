// Package presets contains a collection of preset [github.com/ygrebnov/testutils/docker.Container] objects.
package presets // import "github.com/ygrebnov/testutils/presets"

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/ygrebnov/testutils/docker"
	"gopkg.in/yaml.v3"
)

// preset represents a type capable of producing preset and customizable [docker.Container] objects.
type preset[T any] interface {
	asContainer() T
	asCustomizedContainer(options *docker.Options) T
}

type containerPreset = preset[docker.Container]

// defaultContainerPreset holds container and image data.
type defaultContainerPreset struct {
	Container presetContainer `yaml:"container"`
	Image     presetImage     `yaml:"image"`
}

// presetContainer holds preset container data.
type presetContainer struct {
	Name        string   `yaml:"name"`
	Env         envs     `yaml:"env,omitempty"`
	Ports       []string `yaml:"ports,omitempty"`
	Healthcheck string   `yaml:"healthcheck"`
}

type envs []env

// toSlice converts `envs` to a slice of strings.
func (e envs) toSlice() []string {
	if len(e) == 0 {
		return nil
	}

	slice := make([]string, len(e))
	for i, el := range e {
		var stringVal string
		switch typedVal := el.Value.(type) {
		case int:
			stringVal = strconv.Itoa(typedVal)
		case string:
			stringVal = typedVal
		default:
			panic(errors.New("unhandled preset.env value type"))
		}
		slice[i] = fmt.Sprintf("%s=%s", el.Name, stringVal)
	}

	return slice
}

// env holds preset container environment variables data.
type env struct {
	Name  string `yaml:"name"`
	Value any    `yaml:"value"`
}

// presetImage holds preset container image data.
type presetImage struct {
	Name string `yaml:"name"`
}

// newContainerPreset creates a new `containerPreset` object.
func newContainerPreset(valuesFile string) containerPreset {
	p := new(defaultContainerPreset)

	parsePresetValues(valuesFile, p)

	return p
}

// parsePresetValues sets given `preset` object attributes with values from the given yaml file.
func parsePresetValues(valuesFile string, preset any) {
	// TODO: return error.
	_, currFile, _, ok := runtime.Caller(0)
	if !ok {
		panic(errors.New("cannot locate preset values file"))
	}

	valuesFilePath := filepath.Join(filepath.Dir(currFile), valuesFile)
	valuesData, err := os.ReadFile(valuesFilePath)
	if err != nil {
		panic(err)
	}
	if err = yaml.Unmarshal(valuesData, preset); err != nil {
		panic(err)
	}
}

// asContainer returns a [docker.Container] object with preset attribute values.
func (p *defaultContainerPreset) asContainer() docker.Container {
	return docker.NewContainerWithOptions(p.Image.Name, p.getPresetContainerOptions())
}

// asCustomizedContainer returns a [docker.Container] with preset attribute values overwritten by customized ones.
func (p *defaultContainerPreset) asCustomizedContainer(options *docker.Options) docker.Container {
	return docker.NewContainerWithOptions(p.Image.Name, p.combineContainerOptions(options))
}

// getPresetContainerOptions returns a [docker.Options] object with attributes values from preset yaml file.
func (p *defaultContainerPreset) getPresetContainerOptions() *docker.Options {
	return &docker.Options{
		Name:                 p.Container.Name,
		Healthcheck:          p.Container.Healthcheck,
		EnvironmentVariables: p.Container.Env.toSlice(),
		ExposedPorts:         p.Container.Ports,
	}
}

func (p *defaultContainerPreset) combineContainerOptions(options *docker.Options) *docker.Options {
	combinedOptions := p.getPresetContainerOptions()
	if options == nil {
		return combinedOptions
	}

	if options.Name != "" {
		combinedOptions.Name = options.Name
	}

	if len(options.EnvironmentVariables) > 0 {
		combinedOptions.EnvironmentVariables = options.EnvironmentVariables
	}

	if len(options.ExposedPorts) > 0 {
		combinedOptions.ExposedPorts = options.ExposedPorts
	}

	if options.Healthcheck != "" {
		combinedOptions.Healthcheck = options.Healthcheck
	}

	if options.StartTimeout > 0 {
		combinedOptions.StartTimeout = options.StartTimeout
	}

	return combinedOptions
}
