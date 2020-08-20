// +build e2e

package e2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hashicorp/consul-nia/config"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/require"
)

// tempDirPrefix is the prefix for the directory for a given e2e test
// where files generated from e2e are stored. This directory is
// destroyed after e2e testing if no errors.
const tempDirPrefix = "tmp_"

// resourcesDir is the sub-directory of tempDir where the
// Terraform resources created from running consul-nia are stored
const resourcesDir = "resources"

// configFile is the name of the nia config file
const configFile = "config.hcl"

func TestE2EBasic(t *testing.T) {
	t.Parallel()

	srv, err := newTestConsulServer(t)
	require.NoError(t, err, "failed to start consul server")
	defer srv.Stop()

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "basic")
	err = makeTempDir(tempDir)
	// no defer to delete directory: only delete at end of test if no errors
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, configFile)
	err = makeConfig(configPath, twoTaskConfig(srv.HTTPAddr, tempDir))
	require.NoError(t, err)

	err = runConsulNIA(configPath)
	require.NoError(t, err)

	files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", tempDir, resourcesDir))
	require.NoError(t, err)
	require.Equal(t, 3, len(files))

	contents, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/consul_service_api.txt", tempDir, resourcesDir))
	require.NoError(t, err)
	require.Equal(t, "1.2.3.4", string(contents))

	contents, err = ioutil.ReadFile(fmt.Sprintf("%s/%s/consul_service_web.txt", tempDir, resourcesDir))
	require.NoError(t, err)
	require.Equal(t, "5.6.7.8", string(contents))

	contents, err = ioutil.ReadFile(fmt.Sprintf("%s/%s/consul_service_db.txt", tempDir, resourcesDir))
	require.NoError(t, err)
	require.Equal(t, "10.10.10.10", string(contents))

	// check statefiles exist
	status, err := checkStateFile(srv.HTTPAddr, dbTaskName)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	status, err = checkStateFile(srv.HTTPAddr, webTaskName)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)

	adminRemoveDir(tempDir)
}

func TestE2ERestartConsulNIA(t *testing.T) {
	t.Parallel()

	srv, err := newTestConsulServer(t)
	require.NoError(t, err, "failed to start consul server")
	defer srv.Stop()

	tempDir := fmt.Sprintf("%s%s", tempDirPrefix, "restart")
	err = makeTempDir(tempDir)
	// no defer to delete directory: only delete at end of test if no errors
	require.NoError(t, err)

	configPath := filepath.Join(tempDir, configFile)
	err = makeConfig(configPath, oneTaskConfig(srv.HTTPAddr, tempDir))
	require.NoError(t, err)

	err = runConsulNIA(configPath)
	require.NoError(t, err)

	// rerun nia. confirm no errors e.g. recreating workspaces
	err = runConsulNIA(configPath)
	require.NoError(t, err)

	adminRemoveDir(tempDir)
}

func newTestConsulServer(t *testing.T) (*testutil.TestServer, error) {
	log.SetOutput(ioutil.Discard)
	srv, err := testutil.NewTestServerConfig(func(c *testutil.TestServerConfig) {
		c.LogLevel = "warn"
		c.Stdout = ioutil.Discard
		c.Stderr = ioutil.Discard
	})
	if err != nil {
		return nil, err
	}

	// Register services
	srv.AddAddressableService(t, "api", testutil.HealthPassing,
		"1.2.3.4", 8080, []string{})
	srv.AddAddressableService(t, "web", testutil.HealthPassing,
		"5.6.7.8", 8000, []string{})
	srv.AddAddressableService(t, "db", testutil.HealthPassing,
		"10.10.10.10", 8000, []string{})
	return srv, nil
}

func makeTempDir(tempDir string) error {
	_, err := os.Stat(tempDir)
	if !os.IsNotExist(err) {
		log.Printf("[WARN] temp dir %s was not cleared out after last test. Deleting.", tempDir)
		if err = adminRemoveDir(tempDir); err != nil {
			return err
		}
	}
	return os.Mkdir(tempDir, os.ModePerm)
}

func makeConfig(configPath, contents string) error {
	f, err := os.Create(configPath)
	if err != nil {
		return nil
	}
	defer f.Close()
	config := []byte(contents)
	_, err = f.Write(config)
	return err
}

func runConsulNIA(configPath string) error {
	cmd := exec.Command("sudo", "consul-nia", fmt.Sprintf("--config-file=%s", configPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// adminRemoveDir removes temporary directory created for a test. consul-nia
// (run in sudo mode) creates sub-directories with admin. need 'sudo' to remove.
func adminRemoveDir(tempDir string) error {
	cmd := exec.Command("sudo", "rm", "-r", tempDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkStateFile(consulAddr, taskname string) (int, error) {
	u := fmt.Sprintf("http://%s/v1/kv/%s-env:%s", consulAddr, config.DefaultTFBackendKVPath, taskname)

	resp, err := http.Get(u)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}