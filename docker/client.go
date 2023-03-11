package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
)

// client defines client methods.
type client interface {
	pullImage(ctx context.Context, name string) error
	createContainer(ctx context.Context, image string, name string, env []string) (string, error)
	startContainer(ctx context.Context, id string) error
	createStartContainer(ctx context.Context, image string, name string, env []string) (string, error)
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
	// newClientFn is used to simplify testability of newClient function.
	newClientFn func(ops ...dockerClient.Opt) (*dockerClient.Client, error) = dockerClient.NewClientWithOpts
)

// newClient attempts to create a new client with a new Docker client handler.
// client is stored in a package private 'cli' variable.
func newClient(ctx context.Context) (client, error) {
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
func getClient(ctx context.Context) (client, error) {
	if cli != nil {
		return cli, nil
	}
	return newClient(ctx)
}

// close calls Docker client Close method.
func (c *defaultClient) close() {
	c.handler.Close()
}

// pullImage calls Docker client ImagePull method. Ignores method execution output.
func (c *defaultClient) pullImage(ctx context.Context, name string) error {
	if _, err := c.handler.ImagePull(ctx, name, types.ImagePullOptions{}); err != nil {
		return err
	}
	return nil
}

// createContainer attempts to pull image and then calls Docker client ContainerCreate method.
// Returns created container id.
func (c *defaultClient) createContainer(ctx context.Context, image string, name string, env []string) (string, error) {
	if err := c.pullImage(ctx, image); err != nil {
		return "", err
	}
	resp, err := c.handler.ContainerCreate(
		ctx, &dockerContainer.Config{Image: image, Env: env}, nil, nil, nil, name,
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

// createStartContainer attempts to create and to start a container.
// Returns created container id.
func (c *defaultClient) createStartContainer(ctx context.Context, image string, name string, env []string) (string, error) {
	id, err := c.createContainer(ctx, image, name, env)
	if err != nil {
		return "", err
	}

	return id, c.startContainer(ctx, id)
}

// fetchContainerData calls Docker client ContainerList method with a container name filter.
// Fetched data is saved into container object.
func (c *defaultClient) fetchContainerData(ctx context.Context, container *container) error {
	if len(container.name) == 0 {
		return errEmptyContainerName
	}
	filters := filters.NewArgs()
	filters.Add("name", container.name)
	containers, err := c.handler.ContainerList(ctx, types.ContainerListOptions{Filters: filters})
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

// stopRemoveContainer attempts to stop and remove container.
func (c *defaultClient) stopRemoveContainer(ctx context.Context, id string) error {
	if err := c.stopContainer(ctx, id); err != nil {
		return err
	}
	return c.removeContainer(ctx, id)
}

// PullImage attempts to pull an image.
func PullImage(ctx context.Context, name string) error {
	c, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer c.close()
	return c.pullImage(ctx, name)
}

// CreateContainer attempts to create a new container.
// Returns created container id.
func CreateContainer(ctx context.Context, image string, name string, env []string) (string, error) {
	c, err := getClient(ctx)
	if err != nil {
		return "", err
	}
	defer c.close()
	return c.createContainer(ctx, image, name, env)
}

// StartContainer attempts to start container.
func StartContainer(ctx context.Context, id string) error {
	c, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer c.close()
	return c.startContainer(ctx, id)
}

// CreateStartContainer attempts to create and start a new container.
// Returns created container id.
func CreateStartContainer(ctx context.Context, image string, name string, env []string) (string, error) {
	c, err := getClient(ctx)
	if err != nil {
		return "", err
	}
	defer c.close()
	return c.createStartContainer(ctx, image, name, env)
}

// fetchContainerData attempts to fetch Docker container data.
func fetchContainerData(ctx context.Context, container *container) error {
	c, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer c.close()
	return c.fetchContainerData(ctx, container)
}

// StopContainer attempts to stop container.
func StopContainer(ctx context.Context, id string) error {
	c, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer c.close()
	return c.stopContainer(ctx, id)
}

// RemoveContainer attempts to remove container.
func RemoveContainer(ctx context.Context, id string) error {
	c, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer c.close()
	return c.stopContainer(ctx, id)
}

// StopRemoveContainer attempts to stop and remove container.
func StopRemoveContainer(ctx context.Context, id string) error {
	c, err := getClient(ctx)
	if err != nil {
		return err
	}
	defer c.close()
	return c.stopRemoveContainer(ctx, id)
}