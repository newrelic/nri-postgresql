//nolint:all
package simulation

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/xeipuuv/gojsonschema"
)

// ExecInContainer executes a command in a specified container
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

// RunIntegration executes the integration binary with the provided arguments
func RunIntegration(targetContainer, integrationContainer, binaryPath string, username, password *string, database *string, args ...string) (string, string, error) {
	command := []string{binaryPath}

	if username != nil {
		command = append(command, "-username", *username)
	}
	if password != nil {
		command = append(command, "-password", *password)
	}

	// Always use port 5432 for integration runs
	command = append(command, "-port", "5432")

	if database != nil {
		command = append(command, "-database", *database)
	}
	if targetContainer != "" {
		command = append(command, "-hostname", targetContainer)
	}

	for _, arg := range args {
		command = append(command, arg)
	}

	stdout, stderr, err := ExecInContainer(integrationContainer, command)
	if stderr != "" {
		log.Debug("Integration command Standard Error: ", stderr)
	}

	return stdout, stderr, err
}

// ValidateJSONSchema validates a JSON string against a schema file
func ValidateJSONSchema(fileName string, input string) error {
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
		return fmt.Errorf("error loading JSON schema: %v", err)
	}

	if result.Valid() {
		return nil
	}

	fmt.Printf("Errors for JSON schema: '%s'\n", schemaURI)
	for _, desc := range result.Errors() {
		fmt.Printf("\t- %s\n", desc)
	}
	fmt.Println()

	return fmt.Errorf("the output of the integration doesn't have expected JSON format")
}

// GetSchemaFileName returns the appropriate schema filename for a given sample type
func GetSchemaFileName(sampleType string) string {
	schemaMap := map[string]string{
		"PostgresqlInstanceSample":     "jsonschema-latest.json",
		"PostgresSlowQueries":          "slow-queries-schema.json",
		"PostgresWaitEvents":           "wait-events-schema.json",
		"PostgresBlockingSessions":     "blocking-sessions-schema.json",
		"PostgresIndividualQueries":    "individual-queries-schema.json",
		"PostgresExecutionPlanMetrics": "execution-plan-schema.json",
	}
	return schemaMap[sampleType]
}
