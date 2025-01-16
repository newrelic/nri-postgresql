//go:build integration

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

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

var (
	defaultPassword = flag.String("password", "example", "Default password for postgres")
	defaultUser     = flag.String("username", "postgres", "Default username for postgres")
	defaultDB       = flag.String("database", "demo", "Default database name")
	container       = flag.String("container", "nri-postgresql", "Container name for the integration")
)

const (
	// docker compose service names
	serviceNamePostgres96     = "postgres-9-6"
	serviceNamePostgresLatest = "postgres-latest-supported"
)

func TestMain(m *testing.M) {
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}

func TestSuccessConnection(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name       string
		Hostname   string
		Schema     string
		ExtraFlags []string
	}{
		{
			Name:     "Testing Metrics and inventory for Postgres v9.6.x",
			Hostname: serviceNamePostgres96,
			Schema:   "jsonschema-latest.json",
		},
		{
			Name:     "Testing Metrics and inventory for latest Postgres supported version",
			Hostname: serviceNamePostgresLatest,
			Schema:   "jsonschema-latest.json",
		},
		{
			Name:       "Inventory only for latest Postgres supported version",
			Hostname:   serviceNamePostgresLatest,
			Schema:     "jsonschema-inventory-latest.json",
			ExtraFlags: []string{`-inventory=true`},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			args := append([]string{`-collection_list=all`}, tc.ExtraFlags...)
			stdout, stderr, err := runIntegration(t, tc.Hostname, args...)
			assert.Empty(t, stderr)
			assert.NoError(t, err)
			assert.NotEmpty(t, stdout)
			err = validateJSONSchema(tc.Schema, stdout)
			assert.NoError(t, err)
		})
	}
}

func TestMissingRequiredVars(t *testing.T) {
	// Temporarily set username and password to nil to test missing credentials
	origUser, origPsw := defaultUser, defaultPassword
	defaultUser, defaultPassword = nil, nil
	defer func() {
		defaultUser, defaultPassword = origUser, origPsw
	}()

	_, stderr, err := runIntegration(t, serviceNamePostgresLatest)
	assert.Error(t, err)
	assert.Contains(t, stderr, "invalid configuration: must specify a username and password")
}

func TestIgnoringDB(t *testing.T) {
	args := []string{
		`-collection_list=all`,
		`-collection_ignore_database_list=["demo"]`,
	}
	stdout, stderr, err := runIntegration(t, serviceNamePostgresLatest, args...)
	assert.NoError(t, err)
	assert.Empty(t, stderr)
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

func ExecInContainer(container string, command []string, envVars ...string) (string, string, error) {
	cmdLine := make([]string, 0, 3+len(command))
	cmdLine = append(cmdLine, "exec", "-i")

	for _, envVar := range envVars {
		cmdLine = append(cmdLine, "-e", envVar)
	}

	cmdLine = append(cmdLine, container)
	cmdLine = append(cmdLine, command...)

	log.Debug("executing: docker %s", strings.Join(cmdLine, " "))

	cmd := exec.Command("docker", cmdLine...)

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout := outbuf.String()
	stderr := errbuf.String()

	if err != nil {
		return stdout, stderr, err
	}

	return stdout, stderr, nil
}

func runIntegration(t *testing.T, targetContainer string, args ...string) (string, string, error) {
	t.Helper()

	command := []string{"/nri-postgresql"}

	if defaultUser != nil {
		command = append(command, "-username", *defaultUser)
	}
	if defaultPassword != nil {
		command = append(command, "-password", *defaultPassword)
	}

	// Always use port 5432 for integration runs
	command = append(command, "-port", "5432")

	if defaultDB != nil {
		command = append(command, "-database", *defaultDB)
	}
	if targetContainer != "" {
		command = append(command, "-hostname", targetContainer)
	}

	command = append(command, args...)

	stdout, stderr, err := ExecInContainer(*container, command)
	if stderr != "" {
		log.Debug("Integration command Standard Error: ", stderr)
	}

	return stdout, stderr, err
}
