package docker

import (
	"errors"
	"testing"

	dockerClient "github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

var (
	mockedClient     = dockerClient.Client{}
	errNewClient     error
	errNewClientMock = errors.New("mockedNewClientError")
)

func Test_NewClient(t *testing.T) {
	// newClientFn points to a function returning a mocked Docker client and error.
	newClientFn = func(ops ...dockerClient.Opt) (*dockerClient.Client, error) {
		return &mockedClient, errNewClient
	}

	tests := []struct {
		name string
		// setupMocks is a hook used to dynamically inject mocks' values, which are specific for a test scenario.
		setupMocks    func()
		expectedError error
		expectedCli   client
	}{
		{"nominal", nil, nil, &defaultClient{handler: &mockedClient}},
		{"error", func() { errNewClient = errNewClientMock }, errNewClientMock, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cli = nil // cli should be reset to initial state before each test
			if test.setupMocks != nil {
				test.setupMocks()
			}
			c, err := getClient()
			require.ErrorIs(t, err, test.expectedError)
			require.Equal(t, test.expectedCli, c)
		})
	}
}
