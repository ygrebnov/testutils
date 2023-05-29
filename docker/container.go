package docker

import (
	"bytes"
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
	Exec(ctx context.Context, command string, buffer *bytes.Buffer) error
}

// container holds container data. Implements Container interface.
type container struct {
	id, image, state, status string
	options                  Options
}

// Options holds container optional attributes values which can be set on new container object creation.
type Options struct {
	Name, Healthcheck                  string
	EnvironmentVariables, ExposedPorts []string
	StartTimeout                       int
}

var (
	errEmptyContainerNameAndID = errors.New("empty container name and id")
	errEmptyImageName          = errors.New("empty image name")
	errContainerNotFound       = errors.New("container not found")
	errContainerStartTimeout   = errors.New("container start timeout")
	errIncorrectPortConfig     = errors.New(`incorrect port configuration, expected format is: "containerPort:hostPort"`)
)

// Create creates a new Docker container and saves its id to the container object.
func (c *container) Create(ctx context.Context) error {
	if err = PullImage(ctx, c.image); err != nil {
		return err
	}
	c.id, err = CreateContainer(ctx, c.image, &c.options)
	return err
}

// Start starts Docker container and waits until it is in `running` state. In case healthcheck is defined for the container,
// also waits for service inside the container to finish starting.
func (c *container) Start(ctx context.Context) error {
	var started bool
	started, err = c.HasStarted(ctx)
	if err != nil {
		return err
	} else if started {
		return nil
	}

	if err = StartContainer(ctx, c.id); err != nil {
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

// CreateStart creates a new Docker container and starts it.
func (c *container) CreateStart(ctx context.Context) error {
	if err = c.Create(ctx); err != nil {
		return err
	}
	return c.Start(ctx)
}

// fetchData fetches Docker container data and stores it in the container object.
func (c *container) fetchData(ctx context.Context) error {
	return fetchContainerData(ctx, c)
}

// Stop stops Docker container.
func (c *container) Stop(ctx context.Context) error {
	if len(c.id) == 0 {
		if err = c.fetchData(ctx); err != nil {
			return err
		}
	}
	return StopContainer(ctx, c.id)
}

// Remove removes Docker container.
func (c *container) Remove(ctx context.Context) error {
	// fetchData is called in any case, even if container id is non-empty, because fetchData can return errContainerNotFound.
	// In this way, we avoid returning this error to the caller and allow him to proceed the program normal flow execution.
	err = c.fetchData(ctx)
	switch err {
	case errContainerNotFound:
		return nil
	case nil:
		return RemoveContainer(ctx, c.id)
	}
	return err
}

// StopRemove stops Docker container and removes it.
func (c *container) StopRemove(ctx context.Context) error {
	// fetchData is called in any case, even if container id is non-empty, because fetchData can return errContainerNotFound.
	// In this way, we avoid returning this error to the caller and allow him to proceed the program normal flow execution.
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
// Container is considered as started if its state is 'running' and not 'health: starting'.
func (c *container) HasStarted(ctx context.Context) (bool, error) {
	if err = c.fetchData(ctx); err != nil {
		return false, err
	}
	return c.state == containerStateRunning && !strings.Contains(c.status, "health: "+types.Starting), nil
}

// Exec executes shell command in container.
func (c *container) Exec(ctx context.Context, command string, buffer *bytes.Buffer) error {
	return ExecCommand(ctx, c.id, command, buffer)
}

// NewContainer creates a new [Container] object.
func NewContainer(image string) Container {
	return NewContainerWithOptions(image, Options{})
}

// NewContainerWithOptions creates a new [Container] object with optional attributes values specified.
func NewContainerWithOptions(image string, options Options) Container {
	if options.StartTimeout == 0 {
		options.StartTimeout = defaultContainerStartTimeout
	}
	return &container{image: image, options: options}
}
