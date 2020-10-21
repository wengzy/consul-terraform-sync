package driver

import (
	"context"

	"github.com/hashicorp/consul-terraform-sync/template"
)

//go:generate mockery --name=Driver --filename=driver.go  --output=../mocks/driver

// Driver describes the interface for using an Sync driver to carry out changes
// downstream to update network infrastructure.
type Driver interface {
	// Init initializes the driver and environment
	Init(ctx context.Context) error

	// InitTask initializes the task that the driver executes
	InitTask(task Task, force bool) (template.Template, error)

	// ApplyTask applies change for the task managed by the driver
	ApplyTask(ctx context.Context, taskName string) error

	// Version returns the version of the driver.
	Version() string
}
