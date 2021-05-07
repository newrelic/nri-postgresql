// +build integration

package tests

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

const (
	containerName = "nri-postgresql"
	schema        = "jsonschema.json"
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
	hostname := "db"
	envVars := []string{
		fmt.Sprintf("HOSTNAME=%s", hostname),
		"USERNAME=postgres",
		"PASSWORD=example",
		"DATABASE=demo",
	}
	stdout, _, err := executeDockerCompose(containerName, envVars)
	assert.Nil(t, err)
	assert.NotEmpty(t, stdout)
	err = validateJSONSchema(schema, stdout)
	assert.Nil(t, err)
}

func TestMissingRequiredVars(t *testing.T) {
	hostname := "db"
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
