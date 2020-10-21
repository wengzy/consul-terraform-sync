package driver

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	tmpl "github.com/hashicorp/consul-terraform-sync/template"
	"github.com/hashicorp/consul-terraform-sync/template/tftmpl"
	"github.com/pkg/errors"
)

const (
	// Number of times to retry in addition to initial attempt
	defaultRetry = 2

	// Types of clients that are alternatives to the default Terraform CLI client
	developmentClient = "development"
	testClient        = "test"

	// Permissions for created directories and files
	workingDirPerms = os.FileMode(0750) // drwxr-x---
	filePerms       = os.FileMode(0640) // -rw-r-----

	errSuggestion = "remove Terraform from the configured path or specify a new path to safely install a compatible version."
)

var (
	_ Driver = (*Terraform)(nil)

	errUnsupportedTerraformVersion = fmt.Errorf("unsupported Terraform version: %s", errSuggestion)
	errIncompatibleTerraformBinary = fmt.Errorf("incompatible Terraform binary: %s", errSuggestion)
)

// Terraform is an Sync driver that uses the Terraform CLI to interface with
// low-level network infrastructure.
type Terraform struct {
	log               bool
	persistLog        bool
	path              string
	workingDir        string
	backend           map[string]interface{}
	requiredProviders map[string]interface{}

	workers    map[string]*worker
	workersMux sync.Mutex

	version    string
	clientType string
}

// TerraformConfig configures the Terraform driver
type TerraformConfig struct {
	Log               bool
	PersistLog        bool
	Path              string
	WorkingDir        string
	Backend           map[string]interface{}
	RequiredProviders map[string]interface{}

	// empty/unknown string will default to TerraformCLI client
	ClientType string
}

// NewTerraform configures and initializes a new Terraform driver
func NewTerraform(config *TerraformConfig) *Terraform {
	return &Terraform{
		log:               config.Log,
		persistLog:        config.PersistLog,
		path:              config.Path,
		workingDir:        config.WorkingDir,
		backend:           config.Backend,
		requiredProviders: config.RequiredProviders,
		workers:           make(map[string]*worker),
		clientType:        config.ClientType,
	}
}

// Init initializes the Terraform local environment. The Terraform binary is
// installed to the configured path.
func (tf *Terraform) Init(ctx context.Context) error {
	if _, err := os.Stat(tf.workingDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tf.workingDir, workingDirPerms); err != nil {
			log.Printf("[ERR] (driver.terraform) error creating base work directory: %s", err)
			return err
		}
	}

	if isTFInstalled(tf.path) {
		tfVersion, compatible, err := isTFCompatible(ctx, tf.workingDir, tf.path)
		if err != nil {
			if strings.Contains(err.Error(), "exec format error") {
				return errIncompatibleTerraformBinary
			}
			return err
		}

		if !compatible {
			return errUnsupportedTerraformVersion
		}

		tf.version = tfVersion.String()
		log.Printf("[INFO] (driver.terraform) skipping install, terraform %s "+
			"already exists at path %s/terraform", tf.version, tf.path)
		return nil
	}

	log.Printf("[INFO] (driver.terraform) installing terraform to path '%s'", tf.path)
	if err := tf.install(ctx); err != nil {
		log.Printf("[ERR] (driver.terraform) error installing terraform: %s", err)
		return err
	}
	log.Printf("[INFO] (driver.terraform) successfully installed terraform")
	return nil
}

// Version returns the Terraform CLI version for the Terraform driver.
func (tf *Terraform) Version() string {
	return tf.version
}

// InitTask initializes the task by creating the Terraform root module and
// client to execute task.
func (tf *Terraform) InitTask(task Task, force bool) (tmpl.Template, error) {
	modulePath := filepath.Join(tf.workingDir, task.Name)
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		if err := os.Mkdir(modulePath, workingDirPerms); err != nil {
			log.Printf("[ERR] (driver.terraform) error creating task work directory: %s", err)
			return nil, err
		}
	}

	services := make([]*tftmpl.Service, len(task.Services))
	for i, s := range task.Services {
		services[i] = &tftmpl.Service{
			Datacenter:  s.Datacenter,
			Description: s.Description,
			Name:        s.Name,
			Namespace:   s.Namespace,
			Tag:         s.Tag,
		}
	}

	var vars tftmpl.Variables
	for _, vf := range task.VarFiles {
		tfvars, err := tftmpl.LoadModuleVariables(vf)
		if err != nil {
			return nil, err
		}

		if len(vars) == 0 {
			vars = tfvars
			continue
		}

		for k, v := range tfvars {
			vars[k] = v
		}
	}

	input := tftmpl.RootModuleInputData{
		Backend:      tf.backend,
		Providers:    task.Providers,
		ProviderInfo: task.ProviderInfo,
		Services:     services,
		Task: tftmpl.Task{
			Description: task.Description,
			Name:        task.Name,
			Source:      task.Source,
			Version:     task.Version,
		},
		Variables: vars,
	}
	input.Init()

	if err := tftmpl.InitRootModule(&input, modulePath, filePerms, force); err != nil {
		return nil, err
	}

	worker, err := newWorker(&workerConfig{
		task:       task,
		clientType: tf.clientType,
		log:        tf.log,
		persistLog: tf.persistLog,
		path:       tf.path,
		workingDir: tf.workingDir,
	})
	if err != nil {
		log.Printf("[ERR] (driver.terraform) init worker error: %s", err)
		return nil, err
	}

	tf.workersMux.Lock()
	tf.workers[task.Name] = worker
	tf.workersMux.Unlock()

	tmplPath := filepath.Join(tf.workingDir, task.Name, tftmpl.TFVarsTmplFilename)
	return tmpl.NewTaskTemplate(tmplPath)
}

// ApplyTask applies the task changes.
func (tf *Terraform) ApplyTask(ctx context.Context, taskName string) error {
	tf.workersMux.Lock()
	w, ok := tf.workers[taskName]
	tf.workersMux.Unlock()
	if !ok {
		return fmt.Errorf("task %s has not been initialized", taskName)
	}

	if w.inited {
		log.Printf("[TRACE] (driver.terraform) workspace for task already "+
			"initialized, skipping for '%s'", taskName)
	} else {
		log.Printf("[TRACE] (driver.terraform) initializing workspace '%s'", taskName)
		if err := w.init(ctx); err != nil {
			log.Printf("[ERR] (driver.terraform) error initializing workspace, "+
				"skipping apply for '%s'", taskName)
			return errors.Wrap(err, fmt.Sprintf("error tf-init for '%s'", taskName))
		}
	}

	log.Printf("[TRACE] (driver.terraform) apply '%s'", taskName)
	if err := w.apply(ctx); err != nil {
		return errors.Wrap(err, fmt.Sprintf("error tf-apply for '%s'", taskName))
	}
	return nil
}
