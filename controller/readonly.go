package controller

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/hashicorp/consul-terraform-sync/config"
	"github.com/hashicorp/consul-terraform-sync/driver"
	"github.com/hashicorp/consul-terraform-sync/template"
	"github.com/hashicorp/hcat"
)

var _ Controller = (*ReadOnly)(nil)

// ReadOnly is the controller to run in read-only mode
type ReadOnly struct {
	conf       *config.Config
	driver     driver.Driver
	fileReader func(string) ([]byte, error)
	watcher    template.Watcher
	resolver   template.Resolver
}

// NewReadOnly configures and initializes a new ReadOnly controller
func NewReadOnly(conf *config.Config) (*ReadOnly, error) {
	d, err := newDriver(conf)
	if err != nil {
		return nil, err
	}
	watcher, err := newWatcher(conf)
	if err != nil {
		return nil, err
	}
	return &ReadOnly{
		conf:       conf,
		driver:     d,
		fileReader: ioutil.ReadFile,
		watcher:    watcher,
		resolver:   hcat.NewResolver(),
	}, nil
}

// Init initializes the controller before it can be run
func (ro *ReadOnly) Init(ctx context.Context) error {
	log.Printf("[INFO] (ctrl) initializing driver")
	if err := ro.driver.Init(ctx); err != nil {
		log.Printf("[ERR] (ctrl) error initializing driver: %s", err)
		return err
	}
	log.Printf("[INFO] (ctrl) driver initialized")

	log.Printf("[INFO] (ctrl) initializing all tasks")
	tasks := newDriverTasks(ro.conf)
	for _, task := range tasks {
		log.Printf("[DEBUG] (ctrl) initializing task %q", task.Name)
		tmpl, err := ro.driver.InitTask(task, true)
		if err != nil {
			log.Printf("[ERR] (ctrl) error initializing task %q: %s", task.Name, err)
			return err
		}

		rw.taskTemplates[task.Name] = tmpl
	}

	return nil
}

// Run runs the controller in read-only mode by checking Consul catalog once for
// latest and using the driver to plan network infrastructure changes
func (ro *ReadOnly) Run(ctx context.Context) error {
	log.Println("[INFO] (ctrl) executing all tasks once through")

	completed := make(map[string]bool, len(ro.taskTemplates))
	for {
		done := true
		for taskName := range ro.taskTemplates {
			if !completed[taskName] {
				complete, err := ro.checkApply(ctx, taskName)
				if err != nil {
					return err
				}
				completed[taskName] = complete
				if !complete && done {
					done = false
				}
			}
		}

		if done {
			log.Println("[INFO] (ctrl) all tasks completed once")
			return nil
		}

		select {
		case err := <-ro.watcher.WaitCh(ctx):
			if err != nil {
				log.Printf("[ERR] (ctrl) error watching template dependencies: %s", err)
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
