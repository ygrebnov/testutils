package docker

import (
	"context"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
)

const (
	containerStateRunning        = "running"
	defaultContainerStartTimeout = 60
)

// Container defines container methods.
type Container interface {
	Create(ctx context.Context) error
	Start(ctx context.Context) error
	CreateStart(ctx context.Context) error
	Stop(ctx context.Context) error
	Remove(ctx context.Context) error
	StopRemove(ctx context.Context) error
	HasStarted(ctx context.Context) (bool, error)
}

// container holds container data. Implements Container interface.
type container struct {
	id      string
	name    string
	image   string
	options ContainerOptions
	state   string
	status  string
}

// ContainerOptions holds container optional attributes. [ContainerOptions] is used on creating new containers.
type ContainerOptions struct {
	EnvironmentVariables, ExposedPorts []string
	Healthcheck                        string
	StartTimeout                       int
}

var (
	errEmptyContainerName    = errors.New("empty container name")
	errEmptyImageName        = errors.New("empty image name")
	errContainerNotFound     = errors.New("container not found")
	errContainerStartTimeout = errors.New("container start timeout")
	errIncorrectPortConfig   = errors.New(`incorrect port configuration, expected format is: "containerPort:hostPort"`)
)

// Create attempts to create a Docker container on the host.
// Container object must have non-empty image and name fields values.
func (c *container) Create(ctx context.Context) error {
	if err = c.validate(ctx); err != nil {
		return err
	}
	if err = PullImage(ctx, c.image); err != nil {
		return err
	}
	c.id, err = CreateContainer(ctx, c.image, c.name, c.options)
	return err
}

// Start attempts to start Docker container on the host.
// Waits until container is in `running` state. In case a healthcheck is defined, also waits for service
// inside the container to finish starting.
// Container object must have non-empty name field value.
func (c *container) Start(ctx context.Context) error {
	started, err := c.HasStarted(ctx)
	if err != nil {
		return err
	} else if started {
		return nil
	}

	if err := StartContainer(ctx, c.id); err != nil {
		return err
	}

	t := 0
	for t < c.options.StartTimeout {
		if started, _ = c.HasStarted(ctx); started {
			break
		}
		time.Sleep(time.Second * 1)
		t++
	}
	if !started {
		return errContainerStartTimeout
	}

	return nil
}

// CreateStart attempts to create and start a Docker container on the host.
// Container object must have non-empty image and name fields values.
func (c *container) CreateStart(ctx context.Context) error {
	if err = c.Create(ctx); err != nil {
		return err
	}
	return c.Start(ctx)
}

// fetchData attempts to fetch Docker container data and store it in the container object.
// Container object must have non-empty name field value.
func (c *container) fetchData(ctx context.Context) error {
	if len(c.name) == 0 {
		return errEmptyContainerName
	}
	return fetchContainerData(ctx, c)
}

// Stop attempts to stop Docker container on the host.
// Container object must have non-empty name field value.
func (c *container) Stop(ctx context.Context) error {
	if err = c.fetchData(ctx); err != nil {
		return err
	}
	return StopContainer(ctx, c.id)
}

// Remove attempts to remove Docker container from the host.
// Container object must have non-empty name field value.
func (c *container) Remove(ctx context.Context) error {
	err = c.fetchData(ctx)
	switch err {
	case errContainerNotFound:
		return nil
	case nil:
		return RemoveContainer(ctx, c.id)
	}
	return err
}

// StopRemove attempts to stop and remove Docker container from the host.
// Container object must have non-empty name field value.
func (c *container) StopRemove(ctx context.Context) error {
	err = c.fetchData(ctx)
	switch err {
	case errContainerNotFound:
		return nil
	case nil:
		return StopRemoveContainer(ctx, c.id)
	}
	return err
}

// HasStarted returns container state and healthiness check status. Can be used to check whether both, a container
// and a service inside it have started.
func (c *container) HasStarted(ctx context.Context) (bool, error) {
	if err = c.fetchData(ctx); err != nil {
		return false, err
	}
	return c.state == containerStateRunning && !strings.Contains(c.status, "health: "+types.Starting), nil
}

// validate checks if [Container] object has non-empty `name` and `image` field values.
func (c *container) validate(ctx context.Context) error {
	if len(c.name) == 0 {
		return errEmptyContainerName
	}
	if len(c.image) == 0 {
		return errEmptyImageName
	}
	return nil
}

// ContainerBuilder defines methods to set [Container] object optional attributes values.
type ContainerBuilder interface {
	// SetEnv defines a list of environment variables to be created inside a container.
	//
	// Example:
	//
	//	c := NewContainerBuilder("containerName", "imageName").
	//		SetEnv([]string{"MY_VAR=myvarvalue"}).
	//		Build()
	//	if err := c.Create(context.Context.Background()); err != nil {
	//		panic(err)
	//	}
	SetEnv(env []string) ContainerBuilder
	// ExposePorts defines a list of container ports to be exposed.
	// List elements must have format: "hostPort:containerPort".
	//
	// Example:
	//
	//	c := NewContainerBuilder("containerName", "imageName").ExposePorts([]string{"8080:80"}).Build()
	//	if err := c.Create(context.Context.Background()); err != nil {
	//		panic(err)
	//	}
	ExposePorts(ports []string) ContainerBuilder
	// Healthcheck defines a command to check container healthiness.
	Healthcheck(command string) ContainerBuilder
	// Build creates a new [Container] object after setting container's required and optional attributes values.
	Build() Container
}

// containerBuilder holds attributes values for a new container being built. Implements [ContainerBuilder] interface.
type containerBuilder struct {
	name, image string
	options     ContainerOptions
}

// SetEnv sets [ContainerBuilder] optional `env` attribute value.
func (cb *containerBuilder) SetEnv(env []string) ContainerBuilder {
	cb.options.EnvironmentVariables = env
	return cb
}

// ExposePorts sets [ContainerBuilder] optional `ports` attribute value.
func (cb *containerBuilder) ExposePorts(ports []string) ContainerBuilder {
	cb.options.ExposedPorts = ports
	return cb
}

// Healthcheck sets [ContainerBuilder] optional `healthcheck` attribute value.
func (cb *containerBuilder) Healthcheck(command string) ContainerBuilder {
	cb.options.Healthcheck = command
	return cb
}

// Build creates a new [Container] object from [ContainerBuilder] one.
func (cb *containerBuilder) Build() Container {
	return &container{name: cb.name, image: cb.image, options: cb.options}
}

// NewContainerBuilder creates a new [ContainerBuilder] object. This object is then used to set optional
// container attributes values and finally create a new [Container] object.
//
// Example:
//
//	c := NewContainerBuilder("containerName", "imageName").
//		SetEnv([]string{"MY_VAR=myvarvalue"}).
//		ExposePorts([]string{"8080:80"}).
//		Healthcheck("is_service_running").
//		Build()
//	if err := c.Create(context.Context.Background()); err != nil {
//		panic(err)
//	}
func NewContainerBuilder(name string, image string) ContainerBuilder {
	return &containerBuilder{name: name, image: image, options: ContainerOptions{StartTimeout: defaultContainerStartTimeout}}
}
