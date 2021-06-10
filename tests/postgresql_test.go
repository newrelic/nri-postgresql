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
	// Different schemas files for testing different version of postgres v9.0.x, v9.1.x and 9.2.x or above
	schemav90       = "jsonschema90.json"
	schemav91       = "jsonschema91.json"
	schemav92       = "jsonschema92.json"
	schemaInventory = "jsonschema92.json"
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

func TestSuccessConnection90(t *testing.T) {
	hostname := "db90"
	envVars := []string{
		fmt.Sprintf("HOSTNAME=%s", hostname),
		"USERNAME=postgres",
		"PASSWORD=example",
		"DATABASE=demo",
		"COLLECTION_LIST=ALL",
		"METRIC=true",
	}
	stdout, _, err := executeDockerCompose(containerName, envVars)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	err = validateJSONSchema(schemav90, stdout)
	assert.Nil(t, err)
}

func TestSuccessConnection91(t *testing.T) {
	hostname := "db91"
	envVars := []string{
		fmt.Sprintf("HOSTNAME=%s", hostname),
		"USERNAME=postgres",
		"PASSWORD=example",
		"DATABASE=demo",
		"COLLECTION_LIST=ALL",
		"METRIC=true",
	}
	stdout, _, err := executeDockerCompose(containerName, envVars)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	err = validateJSONSchema(schemav91, stdout)
	assert.Nil(t, err)
}

func TestSuccessConnection92(t *testing.T) {
	hostname := "db92"
	envVars := []string{
		fmt.Sprintf("HOSTNAME=%s", hostname),
		"USERNAME=postgres",
		"PASSWORD=example",
		"DATABASE=demo",
		"COLLECTION_LIST=ALL",
		"METRIC=true",
	}
	stdout, _, err := executeDockerCompose(containerName, envVars)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	err = validateJSONSchema(schemav92, stdout)
	assert.Nil(t, err)
}

func TestSuccessConnectionInventory(t *testing.T) {
	hostname := "db92"
	envVars := []string{
		fmt.Sprintf("HOSTNAME=%s", hostname),
		"USERNAME=postgres",
		"PASSWORD=example",
		"DATABASE=demo",
		"COLLECTION_LIST=ALL",
		"INVENTORY=true",
	}
	stdout, _, err := executeDockerCompose(containerName, envVars)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	err = validateJSONSchema(schemaInventory, stdout)
	assert.Nil(t, err)
}

func TestMissingRequiredVars(t *testing.T) {
	hostname := "db92"
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
