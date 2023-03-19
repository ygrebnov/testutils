package docker

import (
	"context"

	"github.com/pkg/errors"
)

// Container defines container methods.
type Container interface {
	Create(ctx context.Context) error
	Start(ctx context.Context) error
	CreateStart(ctx context.Context) error
	fetchData(ctx context.Context) error
	Stop(ctx context.Context) error
	Remove(ctx context.Context) error
	StopRemove(ctx context.Context) error
	validate(ctx context.Context) error
}

// container holds container data. Implements Container interface.
type container struct {
	id     string
	name   string
	image  string
	env    []string
	ports  []string
	state  string
	status string
}

var (
	errEmptyContainerName = errors.New("empty container name")
	errEmptyImageName     = errors.New("empty image name")
	errContainerNotFound  = errors.New("container not found")
)

// Create attempts to create a Docker container on the host.
// Container object must have non-empty image and name fields values.
func (c *container) Create(ctx context.Context) error {
	if err = c.validate(ctx); err != nil {
		return err
	}
	c.id, err = CreateContainer(ctx, c.image, c.name, c.env, c.ports)
	return err
}

// Start attempts to start Docker container on the host.
// If the Docker container has 'running' state, does nothing.
// Container object must have non-empty name field value.
func (c *container) Start(ctx context.Context) error {
	if err = c.fetchData(ctx); err != nil {
		return err
	}
	if c.state == "running" {
		return nil
	}
	return StartContainer(ctx, c.id)
}

// CreateStart attempts to create and start a Docker container on the host.
// Before that, attempts to pull an image with the container.image name.
// Container object must have non-empty image and name fields values.
func (c *container) CreateStart(ctx context.Context) error {
	if err = c.validate(ctx); err != nil {
		return err
	}
	if err = PullImage(ctx, c.image); err != nil {
		return err
	}
	c.id, err = CreateStartContainer(ctx, c.image, c.name, c.env, c.ports)
	return err
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

// validate checks if container object has non-empty name and image name field values.
func (c *container) validate(ctx context.Context) error {
	if len(c.name) == 0 {
		return errEmptyContainerName
	}
	if len(c.image) == 0 {
		return errEmptyImageName
	}
	return nil
}

// NewContainer creates a new container object.
func NewContainer(name string, image string, env []string, ports []string) Container {
	return &container{name: name, image: image, env: env, ports: ports}
}
