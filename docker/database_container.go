package docker

import (
	"bytes"
	"context"
)

// DatabaseContainer extends [Container] interface with database interaction methods.
type DatabaseContainer interface {
	Container
	ResetDatabase(ctx context.Context) error
}

// Database holds database metadata.
type Database struct {
	Name         string
	ResetCommand string
}

// databaseContainer holds container and inner database metadata. Implements [DatabaseContainer] interface.
type databaseContainer struct {
	container
	database Database
}

// ResetDatabase executes database reset command in container.
func (dc *databaseContainer) ResetDatabase(ctx context.Context) error {
	buffer := bytes.Buffer{}
	return dc.Exec(ctx, dc.database.ResetCommand, &buffer)
}

// NewDatabaseContainer creates a new [DatabaseContainer] object.
func NewDatabaseContainer(image string, db Database) DatabaseContainer {
	return NewDatabaseContainerWithOptions(image, db, Options{})
}

// NewDatabaseContainerWithOptions creates a new [DatabaseContainer] object with optional attributes values specified.
func NewDatabaseContainerWithOptions(image string, db Database, options Options) DatabaseContainer {
	if options.StartTimeout == 0 {
		options.StartTimeout = defaultContainerStartTimeout
	}
	dbContainer := databaseContainer{database: db}
	dbContainer.image = image
	dbContainer.options = options
	return &dbContainer
}
