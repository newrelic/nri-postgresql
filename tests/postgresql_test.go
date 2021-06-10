// +build integration

package tests

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	containerName = "nri-postgresql"
)

func executeDockerCompose(containerName string, envVars []string) (string, string, error) {
	cmdLine := []string{"run"}
	for i := range envVars {
		cmdLine = append(cmdLine, "-e")
		cmdLine = append(cmdLine, envVars[i])
	}
	cmdLine = append(cmdLine, containerName)
	fmt.Printf("executing: docker-compose %s\n", strings.Join(cmdLine, " "))
	cmd := exec.Command("docker-compose", cmdLine...)
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
			Name:     "Testing Metrics for Postgres v9.0x",
			Hostname: "postgres-9-0",
			Schema:   "jsonschema90.json",
			EnvVars:  []string{"METRIC=true"},
		},
		{
			Name:     "Testing Metrics for Postgres v9.1x",
			Hostname: "postgres-9-1",
			Schema:   "jsonschema91.json",
			EnvVars:  []string{"METRIC=true"},
		},
		{
			Name:     "Testing Metrics for Postgres v9.2x +",
			Hostname: "postgres-9-2",
			Schema:   "jsonschema92.json",
			EnvVars:  []string{"METRIC=true"},
		},
		{
			Name:     "Testing Postgres Inventory",
			Hostname: "postgres-9-2",
			Schema:   "jsonschema-inventory.json",
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
			stdout, _, err := executeDockerCompose(containerName, envVars)
			assert.Nil(t, err)
			assert.NotEmpty(t, stdout)
			err = validateJSONSchema(tc.Schema, stdout)
			assert.Nil(t, err)
		})
	}
}

func TestMissingRequiredVars(t *testing.T) {
	hostname := "postgres-9-2"
	envVars := []string{
		fmt.Sprintf("HOSTNAME=%s", hostname),
		"DATABASE=demo",
	}
	_, stderr, err := executeDockerCompose(containerName, envVars)
	assert.NotNil(t, err)
	assert.Contains(t, stderr, "invalid configuration: must specify a username and password")
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
