`testutils` is a collection of utility functions to be used in testing Go code. 

Author: Yaroslav Grebnov

Get started:

* `testutils` can be [installed using `go get` command](#installation)

Features:

* [Utility functions to work with Docker objects](#docker-package),
* [Collection of preconfigured Docker container objects](#presets-package)

`docker` package
----------------

`docker` package provides utility functions to work with Docker objects in Go code tests. 

In order to execute the package functions, Docker daemon must be available and running in the local environment. 

Essentially, it is a collection of functions wrapping `github.com/docker/docker` objects. The main goal of the package is to provide a simplified way of interaction with the Docker objects in Go code tests. 

1. Basic functionality - functions

`docker` package exposes functions for performing essential operations with Docker objects:

* `PullImage(name)` - pulls a Docker image identified by `name`,
* `CreateContainer(image, options)` - pulls a Docker `image` and creates a new Docker container. Optional container attributes values can be specified in `options` argument. Optional attributes list can be found below. Function returns the created container `id`,
* `StartContainer(id)` - starts Docker container identified by given `id`,
* `CreateStartContainer(image, options)` - combines `CreateContainer` and `StartContainer` functions,
* `StopContainer(id)` - stops `id` Docker container,
* `RemoveContainer(id)` - removes `id` Docker container,
* `StopRemoveContainer(id)` - combines `StopContainer` and `RemoveContainer` functions.

All functions take context.Context parameter and return error.

Example, without optional attributes:

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
    containerID, err := docker.CreateStartContainer(ctx, "image/name", nil)
    require.NoError(t, err)
    defer func() { require.NoError(t, docker.StopRemoveContainer(ctx, containerID)) }()

    // Your test here
}
```

In the example above, a few lines of code allow to:

* pull an `image/name` Docker image,
* create and start a new Docker container,
* stop and remove the container at the end of the test.

Optional attributes list:

* `Name` - container name,
* `EnvironmentVariables` - a list of environment variables to be created inside the container. Format is `name=value`,
* `ExposedPorts` - a list of exposed ports. Format is `container_port:host_port`,
* `Healthcheck` - a command to check whether the service inside container has started. Healthcheck commands are automatically prefixed with `CMD-SHELL`,
* `StartTimeout` - service inside the container start timeout in seconds. The default value is `60`.

Example, with optional attributes:

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
	options := docker.Options{
		Name: "test-container",
		EnvironmentVariables: []string{"MYVAR=my-var-value"}, 
		ExposedPorts: []string{"8080:80"},
	}
    containerID, err := docker.CreateStartContainer(ctx, "image/name", &options)
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

`Container` object can be created using two constructors. 

The first constructor, `NewContainer` takes only one parameter `image` (image name) which is the only one `Container` object attribute required on creation. 

The second constructor, `NewContainerWithOptions` in addition to `image`, takes one more `options` parameter. It allows to specify `Container` object optional attributes values (the list can be found above). Optional attributes values can be specified only on new `Container` object creation. Examples of constructing new `Container` objects are provided below.

`Container` object exposed methods:

* `Create` - using the object attributes, pulls a Docker image, creates a new Docker container with all the specified attributes,
* `Start` - starts the container and waits until it starts. If `Options.Healthcheck` has been specified, also waits until the service inside the container starts,
* `CreateStart` - performs all the `Create` actions and starts the created container,
* `Stop` - stops the container,
* `Remove` - removes the container if it exists,
* `StopRemove` - stops and removes the container if it exists.

All methods take context.Context parameter and return error.

An example of using basic `NewContainer` constructor:

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
	testContainer := docker.NewContainer("image/name")
	require.NoError(t, testContainer.CreateStart(ctx)) // using "image/name" image, creates and starts a new Docker container on host
	defer func() { require.NoError(t, testContainer.StopRemove(ctx)) }() // Docker container at the end of the test

	// Your test here
}
```

The main difference between using the `docker` package functions and `Container` object is that the latter allows to interact with Docker containers which may have been created before the code execution. For example, if some test requires that a new container with a specific name should be created, it can be implemented as:

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
	options := docker.Options{
		Name: "test-container",
		EnvironmentVariables: []string{"MYVAR=my-var-value"},
		ExposedPorts: []string{"8080:80"},
	}
	testContainer := docker.NewContainerWithOptions("image/name", options)
	require.NoError(t, testContainer.StopRemove(ctx))  // stops and removes "test-container" Docker container if it exists on host
	require.NoError(t, testContainer.CreateStart(ctx)) // creates and starts a new "test-container" Docker container on host
	defer func() { require.NoError(t, testContainer.StopRemove(ctx)) }() // stops and removes "test-container" Docker container at the end of the test

	// Your test here
}
```


`presets` package
----------------

`presets` package contains a collection of preset `github.com/ygrebnov/testutils/docker.Container` objects. The main idea here is to provide ready-to-use objects with most commonly used configuration already applied. For example, while creating a PostgreSQL container, we may set values of `POSTGRES_PASSWORD`, `POSTGRES_USER`, `PGPORT` environment variables, set a healthcheck based on `pg_isready` command, and expose `5432` port. `presets` package provides a preset `github.com/ygrebnov/testutils/docker.Container` object with such configuration.

Each preset allows to create a new preconfigured `github.com/ygrebnov/testutils/docker.Container` object and a new preconfigured object with customizable optional attributes. For example, for the case when we want to expose port `5433` instead of port `5432` in a PostgreSQL container.

List of `presets`:

* PostgreSQL - preconfigured `github.com/ygrebnov/testutils/docker.Container` object can be obtained using `NewPostgresqlContainer()` function, or the same object, but customizable - using `NewCustomizedPostgresqlContainer(options docker.Options)` function.

Basic example of using presets in tests:

```go
package somepackage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ygrebnov/testutils/presets"
)

func Test_SomeFunction(t *testing.T) {
	ctx := context.Background()
	testContainer := presets.NewPostgresqlContainer()
	require.NoError(t, testContainer.CreateStart(ctx)) // creates and starts a new PostgreSQL Docker container on host
	defer func() { require.NoError(t, testContainer.StopRemove(ctx)) }() // stops and removes PostgreSQL Docker container at the end of the test

	// Your test here
}
```

An example of using presets with customized attributes values:

```go
package somepackage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ygrebnov/testutils/docker"
	"github.com/ygrebnov/testutils/presets"
)

func Test_SomeFunction(t *testing.T) {
	ctx := context.Background()
	options := docker.Options{ExposedPorts: []string{"5432:5433"}} // container port 5432 is bound to the host 5433 port, other configuration remains unchanged
	testContainer := presets.NewCustomizedPostgresqlContainer(options)
	require.NoError(t, testContainer.CreateStart(ctx)) // creates and starts a new PostgreSQL Docker container on host
	defer func() { require.NoError(t, testContainer.StopRemove(ctx)) }() // stops and removes PostgreSQL Docker container at the end of the test

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