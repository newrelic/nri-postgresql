//go:build integration
// +build integration

package tests

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

const (
	// docker compose service names
	serviceNameNRI            = "nri-postgresql"
	serviceNamePostgres90     = "postgres-9-0"
	serviceNamePostgres91     = "postgres-9-1"
	serviceNamePostgresLatest = "postgres-latest-supported"
)

func executeDockerCompose(serviceName string, envVars []string) (string, string, error) {
	cmdLine := []string{"compose", "run"}
	for i := range envVars {
		cmdLine = append(cmdLine, "-e")
		cmdLine = append(cmdLine, envVars[i])
	}
	cmdLine = append(cmdLine, serviceName)
	fmt.Printf("executing: docker %s\n", strings.Join(cmdLine, " "))
	cmd := exec.Command("docker", cmdLine...)
	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()

	stdout := outbuf.String()
	stderr := errbuf.String()
	return stdout, stderr, err
}

func TestMain(m *testing.M) {
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}

func TestSuccessConnection(t *testing.T) {
	t.Parallel()
	defaultEnvVars := []string{
		"USERNAME=postgres",
		"PASSWORD=example",
		"DATABASE=demo",
		"COLLECTION_LIST=ALL",
	}
	testCases := []struct {
		Name     string
		Hostname string
		Schema   string
		EnvVars  []string
	}{
		{
			Name:     "Testing Metrics and inventory for Postgres v9.0x",
			Hostname: serviceNamePostgres90,
			Schema:   "jsonschema90.json",
			EnvVars:  []string{},
		},
		{
			Name:     "Testing Metrics and inventory for Postgres v9.1x",
			Hostname: serviceNamePostgres91,
			Schema:   "jsonschema91.json",
			EnvVars:  []string{},
		},
		{
			Name:     "Testing Metrics and inventory for latest Postgres supported version",
			Hostname: serviceNamePostgresLatest,
			Schema:   "jsonschema-latest.json",
			EnvVars:  []string{},
		},
		{
			Name:     "Inventory only for latest Postgres supported version",
			Hostname: serviceNamePostgresLatest,
			Schema:   "jsonschema-inventory-latest.json",
			EnvVars:  []string{"INVENTORY=true"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			envVars := []string{
				fmt.Sprintf("HOSTNAME=%s", tc.Hostname),
			}
			envVars = append(envVars, defaultEnvVars...)
			envVars = append(envVars, tc.EnvVars...)
			stdout, _, err := executeDockerCompose(serviceNameNRI, envVars)
			assert.Nil(t, err)
			assert.NotEmpty(t, stdout)
			err = validateJSONSchema(tc.Schema, stdout)
			assert.Nil(t, err)
		})
	}
}

func TestMissingRequiredVars(t *testing.T) {
	envVars := []string{
		"HOSTNAME=" + serviceNamePostgresLatest,
		"DATABASE=demo",
	}
	_, stderr, err := executeDockerCompose(serviceNameNRI, envVars)
	assert.NotNil(t, err)
	assert.Contains(t, stderr, "invalid configuration: must specify a username and password")
}

func TestIgnoringDB(t *testing.T) {
	envVars := []string{
		"HOSTNAME=" + serviceNamePostgresLatest,
		"USERNAME=postgres",
		"PASSWORD=example",
		"DATABASE=demo",
		"COLLECTION_LIST=ALL", // The instance has 2 DB: 'demo' and 'postgres'
		`COLLECTION_IGNORE_DATABASE_LIST=["demo"]`,
	}
	stdout, _, err := executeDockerCompose(serviceNameNRI, envVars)
	assert.Nil(t, err)
	assert.Contains(t, stdout, `"database:postgres"`)
	assert.NotContains(t, stdout, `"database:demo"`)
}

func validateJSONSchema(fileName string, input string) error {
	pwd, err := os.Getwd()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	schemaURI := fmt.Sprintf("file://%s", filepath.Join(pwd, "testdata", fileName))
	log.Info("loading schema from %s", schemaURI)
	schemaLoader := gojsonschema.NewReferenceLoader(schemaURI)
	documentLoader := gojsonschema.NewStringLoader(input)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("Error loading JSON schema, error: %v", err)
	}

	if result.Valid() {
		return nil
	}
	fmt.Printf("Errors for JSON schema: '%s'\n", schemaURI)
	for _, desc := range result.Errors() {
		fmt.Printf("\t- %s\n", desc)
	}
	fmt.Printf("\n")
	return fmt.Errorf("The output of the integration doesn't have expected JSON format")
}
