package docker

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

// mockedDockerClient wraps 'dockerClient.Client' type in order to mock its methods.
type mockedDockerClient struct {
	dockerClient.Client
}

// ImagePull is a mocked 'dockerClient.Client' type method.
func (mdc *mockedDockerClient) ImagePull(
	ctx context.Context,
	refStr string,
	options types.ImagePullOptions,
) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), mockedImagePullError
}

// ContainerCreate is a mocked 'dockerClient.Client' type method.
func (mdc *mockedDockerClient) ContainerCreate(
	ctx context.Context,
	config *dockerContainer.Config,
	hostConfig *dockerContainer.HostConfig,
	networkingConfig *network.NetworkingConfig,
	platform *specs.Platform,
	containerName string,
) (dockerContainer.CreateResponse, error) {
	return dockerContainer.CreateResponse{ID: mockedContainerID}, mockedContainerCreateError
}

// ContainerStart is a mocked 'dockerClient.Client' type method.
func (mdc *mockedDockerClient) ContainerStart(
	ctx context.Context,
	containerID string,
	options types.ContainerStartOptions,
) error {
	return nil
}

// ContainerList is a mocked 'dockerClient.Client' type method.
func (mdc *mockedDockerClient) ContainerList(
	ctx context.Context,
	options types.ContainerListOptions,
) ([]types.Container, error) {
	return mockedContainerList, mockedContainerListError
}

// ContainerStop is a mocked 'dockerClient.Client' type method.
func (mdc *mockedDockerClient) ContainerStop(
	ctx context.Context,
	containerID string,
	options dockerContainer.StopOptions,
) error {
	return nil
}

// ContainerRemove is a mocked 'dockerClient.Client' type method.
func (mdc *mockedDockerClient) ContainerRemove(
	ctx context.Context,
	containerID string,
	options types.ContainerRemoveOptions,
) error {
	return nil
}

// Close is a mocked 'dockerClient.Client' type method.
func (mdc *mockedDockerClient) Close() error {
	return nil
}

// resetMocks resets mocks to their default state.
func resetMocks() {
	mockedImagePullError = nil
	mockedContainerCreateError = nil
	mockedContainerListError = nil
	mockedContainerList = []types.Container{mockedRunningContainer.asTypesContainer()}
}

// mockedContainer holds a mocked container data. It is used to store data in one object and
// convert it to external 'types.Container' and internal 'container' types in tests.
type mockedContainer struct {
	id     string
	name   string
	image  string
	state  string
	status string
}

// asTypesContainer converts mocked container into an object of external 'types.Container' type.
func (mc *mockedContainer) asTypesContainer() types.Container {
	return types.Container{ID: mc.id, Names: []string{"/" + mc.name}, Image: mc.image, State: mc.state, Status: mc.status}
}

// asContainer converts mocked container into an object of internal 'container' type.
func (mc *mockedContainer) asContainer() *container {
	return &container{id: mc.id, name: mc.name, image: mc.image, state: mc.state, status: mc.status}
}

var (
	mockedContainerID                                                          = "mockedContainerID"
	mockedContainerName                                                        = "mockedContainerName"
	mockedImageName                                                            = "mockedImageName"
	mockedImagePullError, mockedContainerCreateError, mockedContainerListError error
	mockedContainerList                                                        []types.Container
	mockedCreatedContainer                                                     = mockedContainer{
		id:    mockedContainerID,
		name:  mockedContainerName,
		image: mockedImageName,
	}
	mockedRunningContainer = mockedContainer{
		id:     mockedContainerID,
		name:   mockedContainerName,
		image:  mockedImageName,
		state:  "running",
		status: "Up 8 hours",
	}

	mockedInvalidImageName      = "mockedInvalidImageName"
	mockedInvalidImageContainer = mockedContainer{
		id:    mockedContainerID,
		name:  mockedContainerName,
		image: mockedInvalidImageName,
	}
	mockedEmptyNameContainer = mockedContainer{
		id:    mockedContainerID,
		image: mockedInvalidImageName,
	}
	mockedEmptyImageContainer = mockedContainer{
		id:   mockedContainerID,
		name: mockedContainerName,
	}
	mockedNotFoundContainer = mockedContainer{
		name:  mockedContainerName,
		image: mockedInvalidImageName,
	}
	errInvalidImagePullMock       = errors.New("mockedInvalidImagePullError")
	errDuplicateContainerNameMock = errors.New("mockedDuplicateContainerNameError")
	errContainerListTechnicalMock = errors.New("mockedContainerListTechnicalError")

	mockedEmptyContainerList     = []types.Container{}
	mockedCreatedInContainerList = []types.Container{mockedCreatedContainer.asTypesContainer()}
)

func Test_container(t *testing.T) {
	// cli points to a client with a mocked handler.
	cli = &defaultClient{handler: &mockedDockerClient{}}
	tests := []struct {
		name string
		// setupMocks is a hook used to dynamically inject mocks' values, which are specific for a test scenario.
		setupMocks    func()
		containerData mockedContainer
		// function points to Container methods.
		function      func(_ Container, ctx context.Context) error //
		expectedError error
	}{
		{"create", nil, mockedCreatedContainer, Container.Create, nil},
		{"create_empty_container_name", nil, mockedEmptyNameContainer, Container.Create, errEmptyContainerName},
		{"create_duplicate_container_name", func() { mockedContainerCreateError = errDuplicateContainerNameMock }, mockedCreatedContainer, Container.Create, errDuplicateContainerNameMock},
		{"create_empty_image_name", nil, mockedEmptyImageContainer, Container.Create, errEmptyImageName},
		{"create_invalid_image", func() { mockedImagePullError = errContainerListTechnicalMock }, mockedInvalidImageContainer, Container.Create, errContainerListTechnicalMock},

		{"start_created_container", func() { mockedContainerList = mockedCreatedInContainerList }, mockedCreatedContainer, Container.Start, nil},
		{"start_already_running_container", nil, mockedRunningContainer, Container.Start, nil},
		{"start_empty_container_name", nil, mockedEmptyNameContainer, Container.Start, errEmptyContainerName},
		{"start_container_notfound", func() { mockedContainerList = mockedEmptyContainerList }, mockedCreatedContainer, Container.Start, errContainerNotFound},
		{"start_container_data_fetch_error", func() {
			mockedContainerList = mockedEmptyContainerList
			mockedContainerListError = errContainerListTechnicalMock
		}, mockedCreatedContainer, Container.Start, errContainerListTechnicalMock},

		{"createStart", nil, mockedCreatedContainer, Container.CreateStart, nil},
		{"createStart_empty_container_name", nil, mockedEmptyNameContainer, Container.CreateStart, errEmptyContainerName},
		{"createStart_duplicate_container_name", func() { mockedContainerCreateError = errDuplicateContainerNameMock }, mockedCreatedContainer, Container.CreateStart, errDuplicateContainerNameMock},
		{"createStart_empty_image_name", nil, mockedEmptyImageContainer, Container.CreateStart, errEmptyImageName},
		{"createStart_invalid_image", func() { mockedImagePullError = errInvalidImagePullMock }, mockedInvalidImageContainer, Container.CreateStart, errInvalidImagePullMock},

		{"fetchData", nil, mockedRunningContainer, Container.fetchData, nil},
		{"fetchData_technical_error", func() { mockedContainerListError = errContainerListTechnicalMock }, mockedCreatedContainer, Container.fetchData, errContainerListTechnicalMock},

		{"stop", nil, mockedRunningContainer, Container.Stop, nil},
		{"stop_empty_container_name", nil, mockedEmptyNameContainer, Container.Stop, errEmptyContainerName},
		{"stop_container_notfound", func() { mockedContainerList = mockedEmptyContainerList }, mockedCreatedContainer, Container.Stop, errContainerNotFound},
		{"stop_container_data_fetch_error", func() {
			mockedContainerList = mockedEmptyContainerList
			mockedContainerListError = errContainerListTechnicalMock
		}, mockedRunningContainer, Container.Stop, errContainerListTechnicalMock},

		{"remove", nil, mockedRunningContainer, Container.Remove, nil},
		{"remove_empty_container_name", nil, mockedEmptyNameContainer, Container.Remove, errEmptyContainerName},
		{"remove_container_notfound", func() { mockedContainerList = mockedEmptyContainerList }, mockedNotFoundContainer, Container.Remove, nil},
		{"remove_container_data_fetch_error", func() {
			mockedContainerList = mockedEmptyContainerList
			mockedContainerListError = errContainerListTechnicalMock
		}, mockedRunningContainer, Container.Remove, errContainerListTechnicalMock},

		{"stopRemove", nil, mockedRunningContainer, Container.StopRemove, nil},
		{"stopRemove_empty_container_name", nil, mockedEmptyNameContainer, Container.StopRemove, errEmptyContainerName},
		{"stopRemove_container_notfound", func() { mockedContainerList = mockedEmptyContainerList }, mockedNotFoundContainer, Container.StopRemove, nil},
		{"stopRemove_container_data_fetch_error", func() {
			mockedContainerList = mockedEmptyContainerList
			mockedContainerListError = errContainerListTechnicalMock
		}, mockedRunningContainer, Container.StopRemove, errContainerListTechnicalMock},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resetMocks()
			if test.setupMocks != nil {
				test.setupMocks()
			}
			c := NewContainer(test.containerData.name, test.containerData.image, nil, nil)
			require.ErrorIs(t, test.function(c, context.Background()), test.expectedError)
			if test.expectedError == nil {
				require.Equal(t, test.containerData.asContainer(), c)
			}
		})
	}
}
