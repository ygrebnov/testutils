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
	tests := []struct {
		name          string
		setupMocks    func()
		expectedError error
		expectedCli   client
	}{
		{
			name:        "nominal",
			expectedCli: &defaultClient{handler: &mockedClient},
		},
		{
			name:          "error",
			setupMocks:    func() { errNewClient = errNewClientMock },
			expectedError: errNewClientMock,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cli = nil // cli should be reset to initial state before each test
			if test.setupMocks != nil {
				test.setupMocks()
			}
			c, err := getClient(
				injectables{
					newClientFn: func(...dockerClient.Opt) (*dockerClient.Client, error) {
						return &mockedClient, errNewClient
					},
				},
			)
			require.ErrorIs(t, err, test.expectedError)
			require.Equal(t, test.expectedCli, c)
		})
	}
}
