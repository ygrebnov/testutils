// Package docker is a collection of utilities to operate Docker objects in Go code tests
// in a simplified manner.
package docker

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerContainerFilters "github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// client defines client methods.
type client interface {
	pullImage(ctx context.Context, name string) error
	createContainer(ctx context.Context, image string, options *Options) (string, error)
	startContainer(ctx context.Context, id string) error
	createStartContainer(ctx context.Context, image string, options *Options) (string, error)
	fetchContainerData(ctx context.Context, container *container) error
	stopContainer(ctx context.Context, id string) error
	removeContainer(ctx context.Context, id string) error
	stopRemoveContainer(ctx context.Context, id string) error
	close()
}

// defaultClient holds Docker client handler. Implements client interface.
type defaultClient struct {
	handler dockerClient.CommonAPIClient
}

var (
	// cli points to a client
	cli client
	err error
	ok  bool
	// newClientFn is used to simplify testability of newClient function.
	newClientFn func(ops ...dockerClient.Opt) (*dockerClient.Client, error) = dockerClient.NewClientWithOpts
)

// newClient creates a new client object with a new Docker client handler.
// client is stored in a package private 'cli' variable.
func newClient() (client, error) {
	var c *dockerClient.Client

	c, err = newClientFn(
		dockerClient.FromEnv,
		dockerClient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	cli = &defaultClient{handler: c}
	return cli, nil
}

// getClient returns a pointer to client, stored in 'cli' variable or a newly created one.
func getClient() (client, error) {
	if cli != nil {
		return cli, nil
	}
	return newClient()
}

// close calls Docker client Close method.
func (c *defaultClient) close() {
	c.handler.Close()
}

// pullImage calls Docker client ImagePull method. Ignores method execution output.
func (c *defaultClient) pullImage(ctx context.Context, name string) error {
	var reader io.ReadCloser
	if reader, err = c.handler.ImagePull(ctx, name, types.ImagePullOptions{}); err != nil {
		return err
	}
	defer reader.Close()
	io.ReadAll(reader) // nolint: errcheck
	return nil
}

// createContainer creates a new Docker container and returns its id.
func (c *defaultClient) createContainer(ctx context.Context, image string, options *Options) (string, error) {
	var (
		hostPortString, containerPortString string
		healthcheck                         dockerContainer.HealthConfig
	)
	exposedPorts := make(nat.PortSet, len(options.ExposedPorts))
	portBindings := make(nat.PortMap, len(options.ExposedPorts))
	for _, port := range options.ExposedPorts {
		if hostPortString, containerPortString, ok = strings.Cut(port, ":"); !ok {
			return "", errIncorrectPortConfig
		}
		containerPort := nat.Port(containerPortString + "/tcp")
		exposedPorts[containerPort] = struct{}{}
		portBindings[containerPort] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: hostPortString}}
	}
	if len(options.Healthcheck) > 0 {
		healthcheck.Test = strings.Split("CMD-SHELL "+options.Healthcheck, " ")
		healthcheck.Retries = 29
		healthcheck.StartPeriod = time.Second * 2
		healthcheck.Interval = time.Second * 2
		healthcheck.Timeout = time.Second * 10
	}
	if err := c.pullImage(ctx, image); err != nil {
		return "", err
	}
	resp, err := c.handler.ContainerCreate(
		ctx,
		&dockerContainer.Config{
			Image:        image,
			Env:          options.EnvironmentVariables,
			ExposedPorts: exposedPorts,
			Healthcheck:  &healthcheck,
		},
		&dockerContainer.HostConfig{PortBindings: portBindings},
		nil, nil, options.Name,
	)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// startContainer calls Docker client ContainerStart method.
func (c *defaultClient) startContainer(ctx context.Context, id string) error {
	return c.handler.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

// createStartContainer creates a new Docker container and starts it. Returns created container id.
func (c *defaultClient) createStartContainer(ctx context.Context, image string, options *Options) (string, error) {
	id, err := c.createContainer(ctx, image, options)
	if err != nil {
		return "", err
	}

	return id, c.startContainer(ctx, id)
}

// fetchContainerData fetches Docker container data and saves it into container object.
// Container object must have either non-empty name or id field value.
func (c *defaultClient) fetchContainerData(ctx context.Context, container *container) error {
	filters := dockerContainerFilters.NewArgs()

	switch {
	case len(container.options.Name) > 0:
		filters.Add("name", "/"+container.options.Name)
	case len(container.id) > 0:
		filters.Add("id", container.id)
	default:
		return errEmptyContainerNameAndID
	}

	containers, err := c.handler.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: filters})
	switch {
	case err != nil:
		return err
	case len(containers) == 0:
		return errContainerNotFound
	}
	container.id = containers[0].ID
	container.state = containers[0].State
	container.status = containers[0].Status
	return nil
}

// stopContainer calls Docker client ContainerStop method.
func (c *defaultClient) stopContainer(ctx context.Context, id string) error {
	return c.handler.ContainerStop(ctx, id, dockerContainer.StopOptions{})
}

// removeContainer calls Docker client ContainerRemove method.
func (c *defaultClient) removeContainer(ctx context.Context, id string) error {
	return c.handler.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})
}

// stopRemoveContainer stops and removes Docker container.
func (c *defaultClient) stopRemoveContainer(ctx context.Context, id string) error {
	if err := c.stopContainer(ctx, id); err != nil {
		return err
	}
	return c.removeContainer(ctx, id)
}

// PullImage pulls a Docker image with the given name.
func PullImage(ctx context.Context, name string) error {
	if len(name) == 0 {
		return errEmptyImageName
	}
	c, err := getClient()
	if err != nil {
		return err
	}
	defer c.close()
	return c.pullImage(ctx, name)
}

// CreateContainer creates a new Docker container and returns its id.
func CreateContainer(ctx context.Context, image string, options *Options) (string, error) {
	c, err := getClient()
	if err != nil {
		return "", err
	}
	defer c.close()
	return c.createContainer(ctx, image, options)
}

// StartContainer starts Docker container.
func StartContainer(ctx context.Context, id string) error {
	c, err := getClient()
	if err != nil {
		return err
	}
	defer c.close()
	return c.startContainer(ctx, id)
}

// CreateStartContainer creates a new Docker container and starts it. Returns created container id.
func CreateStartContainer(ctx context.Context, image string, options *Options) (string, error) {
	c, err := getClient()
	if err != nil {
		return "", err
	}
	defer c.close()
	return c.createStartContainer(ctx, image, options)
}

// fetchContainerData fetches Docker container data and saves in into container object.
func fetchContainerData(ctx context.Context, container *container) error {
	c, err := getClient()
	if err != nil {
		return err
	}
	defer c.close()
	return c.fetchContainerData(ctx, container)
}

// StopContainer stops Docker container.
func StopContainer(ctx context.Context, id string) error {
	c, err := getClient()
	if err != nil {
		return err
	}
	defer c.close()
	return c.stopContainer(ctx, id)
}

// RemoveContainer removes Docker container.
func RemoveContainer(ctx context.Context, id string) error {
	c, err := getClient()
	if err != nil {
		return err
	}
	defer c.close()
	return c.stopContainer(ctx, id)
}

// StopRemoveContainer stops and removes Docker container.
func StopRemoveContainer(ctx context.Context, id string) error {
	c, err := getClient()
	if err != nil {
		return err
	}
	defer c.close()
	return c.stopRemoveContainer(ctx, id)
}
