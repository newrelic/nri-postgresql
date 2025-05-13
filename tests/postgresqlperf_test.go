//go:build integration

package tests

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/newrelic/nri-postgresql/tests/simulation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	oldestSupportedPerf  = "postgresql-perf-oldest"
	latestSupportedPerf  = "postgresql-perf-latest"
	unsupportedPerf      = "postgresql-noext"
	perfContainers       = []string{oldestSupportedPerf, latestSupportedPerf}
	nonPerfContainers    = []string{unsupportedPerf}
	integrationContainer = "nri_postgresql"

	defaultBinPath = "/nri-postgresql"
	defaultUser    = "dbuser"
	defaultPass    = "dbpassword"
	defaultPort    = 5432
	defaultDB      = "demo"
	testDB         = "titanic"

	// cli flags
	container  = flag.String("container", integrationContainer, "container where the integration is installed")
	binaryPath = flag.String("bin", defaultBinPath, "Integration binary path")
	user       = flag.String("user", defaultUser, "Postgresql user name")
	psw        = flag.String("psw", defaultPass, "Postgresql user password")
	port       = flag.Int("port", defaultPort, "Postgresql port")
	database   = flag.String("database", defaultDB, "Postgresql database")

	allSampleTypes = []string{
		"PostgresqlInstanceSample",
		"PostgresSlowQueries",
		"PostgresWaitEvents",
		"PostgresBlockingSessions",
		"PostgresIndividualQueries",
		"PostgresExecutionPlanMetrics",
	}
)

func TestMain(m *testing.M) {
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}

func TestIntegrationWithDatabaseLoadPerfEnabled(t *testing.T) {
	tests := []struct {
		name                string
		expectedSampleTypes []string
		containers          []string
		args                []string
	}{
		{
			name:                "Performance metrics collection test",
			expectedSampleTypes: allSampleTypes,
			containers:          perfContainers,
			args:                []string{`-collection_list=all`, `-enable_query_monitoring=true`},
		},
		{
			name: "Performance metrics collection test without collection list",
			expectedSampleTypes: []string{
				"PostgresqlInstanceSample",
			},
			containers: perfContainers,
			args:       []string{`-enable_query_monitoring=true`},
		},
		{
			name: "Performance metrics collection test without query monitoring enabled",
			expectedSampleTypes: []string{
				"PostgresqlInstanceSample",
			},
			containers: perfContainers,
			args:       []string{`-collection_list=all`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, container := range tt.containers {
				t.Run(container, func(t *testing.T) {
					// Create simulation controller
					controller := simulation.NewSimulationController(container)
					// Start all simulations
					done := controller.StartAllSimulations(t)

					time.Sleep(30 * time.Second)

					stdout, stderr, err := simulation.RunIntegration(container, integrationContainer, *binaryPath, user, psw, database, tt.args...)
					if stderr != "" {
						fmt.Println(stderr)
					}
					assert.NoError(t, err, "Running Integration Failed")

					samples := strings.Split(stdout, "\n")
					foundSampleTypes := make(map[string]bool)

					for _, sample := range samples {
						sample = strings.TrimSpace(sample)
						if sample == "" {
							continue
						}

						sampleType := getSampleType(sample, allSampleTypes)
						require.NotEmpty(t, sampleType, "No sample type found in JSON output: %s", sample)
						require.Contains(t, tt.expectedSampleTypes, sampleType, "Found unexpected sample type %q", sampleType)
						foundSampleTypes[sampleType] = true

						validateSample(t, sample, sampleType)
					}
					samplesFound := getFoundSampleTypes(foundSampleTypes)
					require.ElementsMatch(t, tt.expectedSampleTypes, samplesFound, "Not all expected sample types where found expected %v, found %v", tt.expectedSampleTypes, samplesFound)
					// Wait for all simulations to complete
					<-done
				})
			}
		})
	}
}

func TestIntegrationUnsupportedDatabase(t *testing.T) {
	tests := []struct {
		name       string
		containers []string
		args       []string
	}{
		{
			name:       "Performance metrics collection with unsupported database - perf enabled",
			containers: nonPerfContainers,
			args:       []string{`-collection_list=all`, `-enable_query_monitoring=true`},
		},
		{
			name:       "Performance metrics collection with unsupported database - perf disabled",
			containers: nonPerfContainers,
			args:       []string{`-collection_list=all`, `-enable_query_monitoring=false`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, container := range tt.containers {
				t.Run(container, func(t *testing.T) {
					stdout, stderr, err := simulation.RunIntegration(container, integrationContainer, *binaryPath, user, psw, database, tt.args...)
					if stderr != "" {
						fmt.Println(stderr)
					}
					assert.NoError(t, err, "Running Integration Failed")

					// Validate JSON format
					var j map[string]interface{}
					err = json.Unmarshal([]byte(stdout), &j)
					assert.NoError(t, err, "Integration Output Is An Invalid JSON")

					// Verify it's a PostgresqlInstanceSample
					assert.Contains(t, stdout, "PostgresqlInstanceSample",
						"Integration output does not contain PostgresqlInstanceSample")

					// Validate against schema
					err = simulation.ValidateJSONSchema("jsonschema-latest.json", stdout)
					assert.NoError(t, err, "Output failed schema validation")
				})
			}
		})
	}
}

func getSampleType(sample string, sampleTypes []string) string {
	for _, sType := range sampleTypes {
		if strings.Contains(sample, sType) {
			return sType
		}
	}
	return ""
}

func validateSample(t *testing.T, sample string, sampleType string) {
	t.Helper()
	// Validate JSON format
	var j map[string]interface{}
	err := json.Unmarshal([]byte(sample), &j)
	require.NoError(t, err, "Got an invalid JSON as output: %s", sample)

	// Validate schema
	t.Run(fmt.Sprintf("Validating JSON schema for sample: %s", sampleType), func(t *testing.T) {
		schemaFile := simulation.GetSchemaFileName(sampleType)
		require.NotEmpty(t, schemaFile, "Schema file not found for sample type: %s", sampleType)

		err = simulation.ValidateJSONSchema(schemaFile, sample)
		assert.NoError(t, err, "Sample failed schema validation for type: %s", sampleType)
	})
}

func getFoundSampleTypes(foundSampleTypes map[string]bool) []string {
	foundSampleTypesKeys := make([]string, 0)
	for key, _ := range foundSampleTypes {
		foundSampleTypesKeys = append(foundSampleTypesKeys, key)
	}
	return foundSampleTypesKeys
}
