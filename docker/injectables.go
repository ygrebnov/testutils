package docker

import dockerClient "github.com/docker/docker/client"

// injectables contains injectable functions.
// It is used to inject dependencies into the code.
type injectables struct {
	newClientFn func(opt ...dockerClient.Opt) (*dockerClient.Client, error)
}

// getNewClientFn returns a specified injectable function or a default one.
func (i *injectables) getNewClientFn() func(opt ...dockerClient.Opt) (*dockerClient.Client, error) {
	if i.newClientFn != nil {
		return i.newClientFn
	}

	return dockerClient.NewClientWithOpts
}
