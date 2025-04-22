package docker

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	dockerClient "github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

// mockedDockerClient wraps [dockerClient.Client] type in order to mock its methods.
type mockedDockerClient struct {
	dockerClient.Client
}

// ImagePull is a mocked [dockerClient.Client] type method.
func (mdc *mockedDockerClient) ImagePull(
	_ context.Context,
	_ string,
	_ image.PullOptions,
) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), mockedImagePullError
}

// ContainerCreate is a mocked [dockerClient.Client] type method.
func (mdc *mockedDockerClient) ContainerCreate(
	_ context.Context,
	_ *dockerContainer.Config,
	_ *dockerContainer.HostConfig,
	_ *network.NetworkingConfig,
	_ *specs.Platform,
	_ string,
) (dockerContainer.CreateResponse, error) {
	return dockerContainer.CreateResponse{ID: mockedContainerID}, mockedContainerCreateError
}

// ContainerStart is a mocked [dockerClient.Client] type method.
func (mdc *mockedDockerClient) ContainerStart(
	_ context.Context,
	_ string,
	_ dockerContainer.StartOptions,
) error {
	return nil
}

// ContainerList is a mocked [dockerClient.Client] type method.
func (mdc *mockedDockerClient) ContainerList(
	_ context.Context,
	_ dockerContainer.ListOptions,
) ([]dockerContainer.Summary, error) {
	return mockedContainerListValues.next()
}

// ContainerStop is a mocked [dockerClient.Client] type method.
func (mdc *mockedDockerClient) ContainerStop(
	_ context.Context,
	_ string,
	_ dockerContainer.StopOptions,
) error {
	return nil
}

// ContainerRemove is a mocked [dockerClient.Client] type method.
func (mdc *mockedDockerClient) ContainerRemove(
	_ context.Context,
	_ string,
	_ dockerContainer.RemoveOptions,
) error {
	return nil
}

// Close is a mocked [dockerClient.Client] type method.
func (mdc *mockedDockerClient) Close() error {
	return nil
}

// containerListMockValue contains a pair of mocked [dockerClient.Client.ContainerList] method return values.
type containerListMockValue struct {
	mockContainers []dockerContainer.Summary
	mockError      error
}

// containerListMockValues contains a slice of mocked [dockerClient.Client.ContainerList] method return values.
type containerListMockValues struct {
	values []containerListMockValue
	size   int
	index  int
}

// next iteratively returns containerListMockValues values to caller.
func (cl *containerListMockValues) next() ([]dockerContainer.Summary, error) {
	if cl.size == 1 {
		return cl.values[0].mockContainers, cl.values[0].mockError
	}
	value := cl.values[cl.index]
	cl.index++
	return value.mockContainers, value.mockError
}

// newContainerListMockValues creates a new containerListMockValues object.
func newContainerListMockValues(values ...containerListMockValue) containerListMockValues {
	return containerListMockValues{values: values, size: len(values), index: 0}
}

// resetMocks resets mocks to their default state.
func resetMocks() {
	mockedImagePullError = nil
	mockedContainerCreateError = nil
	mockedContainerListValues = newContainerListMockValues(
		containerListMockValue{mockedRunningInContainerList, nil},
	)
}

// mockedContainer holds mocked container data. It is used to store data in one object and
// convert it to external 'types.Container' and internal 'container' types in tests.
type mockedContainer struct {
	id     string
	name   string
	image  string
	state  string
	status string
}

// asTypesContainer converts mocked container into an object of external [dockerContainer.Summary] type.
func (mc *mockedContainer) asTypesContainer() dockerContainer.Summary {
	return dockerContainer.Summary{ID: mc.id, Names: []string{"/" + mc.name}, Image: mc.image, State: mc.state, Status: mc.status}
}

// asContainer converts mocked container into an object of internal 'container' type.
func (mc *mockedContainer) asContainer() *container {
	return &container{
		id:     mc.id,
		image:  mc.image,
		state:  mc.state,
		status: mc.status,
		options: &Options{
			Name:         mc.name,
			StartTimeout: 60,
		},
	}
}

var (
	mockedContainerID                                = "mockedContainerID"
	mockedContainerName                              = "mockedContainerName"
	mockedImageName                                  = "mockedImageName"
	mockedImagePullError, mockedContainerCreateError error
	mockedContainerListValues                        containerListMockValues
	mockedCreatedContainer                           = mockedContainer{
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

	mockedEmptyContainerList                = []dockerContainer.Summary{}
	mockedCreatedInContainerList            = []dockerContainer.Summary{mockedCreatedContainer.asTypesContainer()}
	mockedRunningInContainerList            = []dockerContainer.Summary{mockedRunningContainer.asTypesContainer()}
	mockedContainerListValuesCreatedRunning = newContainerListMockValues(
		containerListMockValue{mockedCreatedInContainerList, nil},
		containerListMockValue{mockedRunningInContainerList, nil},
	)
	mockedContainerListValuesEmpty = newContainerListMockValues(
		containerListMockValue{mockedEmptyContainerList, nil},
	)
	mockedContainerListValuesEmptyTechnical = newContainerListMockValues(
		containerListMockValue{mockedEmptyContainerList, errContainerListTechnicalMock},
	)
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
		{"create_empty_container_name", nil, mockedEmptyNameContainer, Container.Create, nil},
		{"create_duplicate_container_name", func() { mockedContainerCreateError = errDuplicateContainerNameMock }, mockedCreatedContainer, Container.Create, errDuplicateContainerNameMock},
		{"create_empty_image_name", nil, mockedEmptyImageContainer, Container.Create, errEmptyImageName},
		{"create_invalid_image", func() { mockedImagePullError = errContainerListTechnicalMock }, mockedInvalidImageContainer, Container.Create, errContainerListTechnicalMock},

		{"start_created_container", func() {
			mockedContainerListValues = mockedContainerListValuesCreatedRunning
		}, mockedRunningContainer, Container.Start, nil},
		{"start_already_running_container", nil, mockedRunningContainer, Container.Start, nil},
		{"start_empty_container_name_and_id", nil, mockedEmptyNameContainer, Container.Start, errEmptyContainerNameAndID},
		{"start_container_notfound", func() {
			mockedContainerListValues = mockedContainerListValuesEmpty
		}, mockedCreatedContainer, Container.Start, errContainerNotFound},
		{"start_container_data_fetch_error", func() {
			mockedContainerListValues = mockedContainerListValuesEmptyTechnical
		}, mockedCreatedContainer, Container.Start, errContainerListTechnicalMock},

		{"createStart", nil, mockedRunningContainer, Container.CreateStart, nil},
		{"createStart_empty_container_name", nil, mockedRunningContainer, Container.CreateStart, nil},
		{"createStart_duplicate_container_name", func() { mockedContainerCreateError = errDuplicateContainerNameMock }, mockedCreatedContainer, Container.CreateStart, errDuplicateContainerNameMock},
		{"createStart_empty_image_name", nil, mockedEmptyImageContainer, Container.CreateStart, errEmptyImageName},
		{"createStart_invalid_image", func() { mockedImagePullError = errInvalidImagePullMock }, mockedInvalidImageContainer, Container.CreateStart, errInvalidImagePullMock},

		{"stop", nil, mockedRunningContainer, Container.Stop, nil},
		{"stop_empty_container_name_and_id", nil, mockedEmptyNameContainer, Container.Stop, errEmptyContainerNameAndID},
		{"stop_container_notfound", func() {
			mockedContainerListValues = mockedContainerListValuesEmpty
		}, mockedCreatedContainer, Container.Stop, errContainerNotFound},
		{"stop_container_data_fetch_error", func() {
			mockedContainerListValues = mockedContainerListValuesEmptyTechnical
		}, mockedRunningContainer, Container.Stop, errContainerListTechnicalMock},

		{"remove", nil, mockedRunningContainer, Container.Remove, nil},
		{"remove_empty_container_name_and_id", nil, mockedEmptyNameContainer, Container.Remove, errEmptyContainerNameAndID},
		{"remove_container_notfound", func() {
			mockedContainerListValues = mockedContainerListValuesEmpty
		}, mockedNotFoundContainer, Container.Remove, nil},
		{"remove_container_data_fetch_error", func() {
			mockedContainerListValues = mockedContainerListValuesEmptyTechnical
		}, mockedRunningContainer, Container.Remove, errContainerListTechnicalMock},

		{"stopRemove", nil, mockedRunningContainer, Container.StopRemove, nil},
		{"stopRemove_empty_container_name_and_id", nil, mockedEmptyNameContainer, Container.StopRemove, errEmptyContainerNameAndID},
		{"stopRemove_container_notfound", func() {
			mockedContainerListValues = mockedContainerListValuesEmpty
		}, mockedNotFoundContainer, Container.StopRemove, nil},
		{"stopRemove_container_data_fetch_error", func() {
			mockedContainerListValues = mockedContainerListValuesEmptyTechnical
		}, mockedRunningContainer, Container.StopRemove, errContainerListTechnicalMock},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resetMocks()
			if test.setupMocks != nil {
				test.setupMocks()
			}
			c := NewContainerWithOptions(
				test.containerData.image,
				&Options{Name: test.containerData.name},
			)
			require.ErrorIs(t, test.function(c, context.Background()), test.expectedError)
			if test.expectedError == nil {
				require.Equal(t, test.containerData.asContainer(), c)
			}
		})
	}
}
