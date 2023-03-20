`testutils` is a collection of utility functions to be used in testing Go code. 

Author: Yaroslav Grebnov

Get started:

* `testutils` can be [installed using `go get` command](#installation)


`docker` package
----------------

`docker` package provides utility functions to work with Docker objects in Go code tests. 

In order to execute the package functions, Docker daemon must be available and running in the local environment. 

Essentially, it is a collection of functions wrapping `github.com/docker/docker` objects. The main goal of the package is to provide a simplified way of interaction with the Docker objects in Go code tests. 

1. Basic functionality - functions

`docker` package exposes functions for performing essential operations with Docker objects:

* `PullImage(name)` - pulls a Docker image identified by `name`,
* `CreateContainer(image, name, env, ports)` - pulls a Docker `image` and creates a new Docker `name` Docker container with `env` list of environment variables created inside the container and with exposed list of `ports`. Function returns the created container `id`,
* `StartContainer(id)` - starts `id` Docker container,
* `CreateStartContainer(image, name, env, ports)` - combines `CreateContainer` and `StartContainer` functions,
* `StopContainer(id)` - stops `id` Docker container,
* `RemoveContainer(id)` - removes `id` Docker container,
* `StopRemoveContainer(id)` - combines `StopContainer` and `RemoveContainer` functions.

All functions take context.Context parameter and return error.

Example:

```go
package somepackage

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/ygrebnov/testutils/docker"
)

func Test_SomeFunction(t *testing.T) {
    ctx := context.Background()
	env := []string{"MYVAR=my-var-value"}
	ports := []string{"8080:80"}
    containerID, err := docker.CreateStartContainer(ctx, "image/name", "test-container", env, ports)
    require.NoError(t, err)
    defer func() { require.NoError(t, docker.StopRemoveContainer(ctx, containerID)) }()

    // Your test here
}
```

In the example above, a few lines of code allow to:

* pull an `image/name` Docker image, 
* create and start a new Docker container with name `test-container`, 
* create `MYVAR` environment variable with value `my-var-value` inside the container,
* expose port `80` and bind it to the host port `8080`,
* stop and remove `test-container` at the end of the test.

2. Extended functionality - `Container` object

`docker` package also exposes a `Container` object which holds container data and synchronizes it with the corresponding Docker container on the host.

`Container` object is constructed using a builder. `Container` required attributes, `name` and `image`, values are set via the main constructor function, `NewContainerBuilder(name, image)`. Optional attributes, `env` and `ports` values can be set via additional constructor functions, `SetEnv(env []string)` and `ExposePorts(ports []string)`. The constructor functions can be chained. The last constructor function in the chain must be `Build()`, which aggregates all the attributes values set during the building process and creates the final `Container` object. Given that this package main objective is to provide the most simplified manner of working with Docker objects, the described above design approach has been chosen intentionally. An example of constructing a new `Container` object is provided below.

`Container` object exposed methods:

* `Create` - using the object attributes, pulls a Docker image, creates a new Docker container and environment variables inside the container,
* `Start` - fetches the corresponding Docker container data and saves it as object attributes values, starts the container,
* `CreateStart` - performs all the `Create` actions and starts the created container,
* `Stop` - fetches the corresponding Docker container data and saves it as object attributes values, stops the container,
* `Remove` - fetches the corresponding Docker container data and saves it as object attributes values, removes the container if it exists,
* `StopRemove` - fetches the corresponding Docker container data and saves it as object attributes values, stops and removes the container if it exists.

All methods take context.Context parameter and return error.

The main difference between using the `docker` package functions and `Container` object is that the latter allows to interact with Docker containers which may have been created before the code execution.

For example, if some test requires that a new container with a specific name should be created, it can be implemented as:

```go
package somepackage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ygrebnov/testutils/docker"
)

func Test_SomeFunction(t *testing.T) {
	ctx := context.Background()
	testContainer := docker.NewContainerBuilder("test-container", "image/name").
		SetEnv([]string{"MYVAR=my-var-value"}).
		ExposePorts([]string{"8080:80"}).
		Build()
	require.NoError(t, testContainer.StopRemove(ctx))  // stops and removes "test-container" Docker container if it exists on host
	require.NoError(t, testContainer.CreateStart(ctx)) // creates and starts a new "test-container" Docker container on host
	defer func() { require.NoError(t, testContainer.StopRemove(ctx)) }() // stops and removes "test-container" Docker container at the end of the test

	// Your test here
}
```


Installation
------------

To install `testutils`, use `go get` command:

```sh
go get github.com/github.com/ygrebnov/testutils
```

This will make the `github.com/github.com/ygrebnov/testutils/docker` package available for you.


Staying up to date
------------------

To update `testutils` to the latest version, use `go get github.com/github.com/ygrebnov/testutils@latest`.


Supported Go versions
---------------------

We currently support the latest major Go versions from 1.19 onwards.


License
-------

This project is licensed under the terms of the MIT license.