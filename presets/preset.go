// Package presets contains a collection of preset [github.com/ygrebnov/testutils/docker.Container] objects.
package presets

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

type preset interface {
	asContainerBuilder() docker.ContainerBuilder
	asContainer() docker.Container
}

type defaultPreset struct {
	Container presetContainer `yaml:"container"`
	Image     presetImage     `yaml:"image"`
}

type presetContainer struct {
	Name        string               `yaml:"name"`
	Env         []presetContainerEnv `yaml:"env,omitempty"`
	Ports       []string             `yaml:"ports,omitempty"`
	Healthcheck string               `yaml:"healthcheck"`
}

type presetContainerEnv struct {
	Name  string `yaml:"name"`
	Value any    `yaml:"value"`
}

type presetImage struct {
	Name string `yaml:"name"`
}

func newPreset(valuesFile string) preset {
	p := defaultPreset{}
	_, currFile, _, ok := runtime.Caller(0)
	if !ok {
		panic(errors.New("cannot locate preset values file"))
	}
	valuesFilePath := filepath.Join(filepath.Dir(currFile), valuesFile)
	valuesData, err := os.ReadFile(valuesFilePath)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(valuesData, &p); err != nil {
		panic(err)
	}
	return &p
}

func (p *defaultPreset) asContainerBuilder() docker.ContainerBuilder {
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

	return docker.NewContainerBuilder(p.Container.Name, p.Image.Name).
		SetEnv(env).
		ExposePorts(p.Container.Ports).
		Healthcheck(p.Container.Healthcheck)
}

func (p *defaultPreset) asContainer() docker.Container {
	return p.asContainerBuilder().Build()
}
