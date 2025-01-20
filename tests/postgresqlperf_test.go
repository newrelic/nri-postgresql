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
)

func TestMain(m *testing.M) {
	flag.Parse()
	result := m.Run()
	os.Exit(result)
}

func TestIntegrationWithDatabaseLoadPerfEnabled(t *testing.T) {
	tests := []struct {
		name          string
		expectedOrder []string
		containers    []string
		args          []string
	}{
		{
			name: "Performance metrics collection test",
			expectedOrder: []string{
				"PostgresqlInstanceSample",
				"PostgresSlowQueries",
				"PostgresWaitEvents",
				"PostgresBlockingSessions",
				"PostgresIndividualQueries",
				"PostgresExecutionPlanMetrics",
			},
			containers: perfContainers,
			args:       []string{`-collection_list=all`, `-enable_query_monitoring=true`},
		},
		{
			name: "Performance metrics collection test without collection list",
			expectedOrder: []string{
				"PostgresqlInstanceSample",
			},
			containers: perfContainers,
			args:       []string{`-enable_query_monitoring=true`},
		},
		{
			name: "Performance metrics collection test without query monitoring enabled",
			expectedOrder: []string{
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
					count := 0

					for idx, sample := range samples {
						sample = strings.TrimSpace(sample)
						if sample == "" {
							continue
						}

						// Validate JSON
						var j map[string]interface{}
						err = json.Unmarshal([]byte(sample), &j)
						assert.NoError(t, err, "Sample %d - Integration Output Is An Invalid JSONs", idx)

						// Validate sample type
						t.Run(fmt.Sprintf("Validating JSON Schema For %s", tt.expectedOrder[count]), func(t *testing.T) {
							sampleType := tt.expectedOrder[count]
							if !strings.Contains(sample, sampleType) {
								t.Errorf("Integration output does not contain: %s", tt.expectedOrder[count])
							}

							// Validate against schema
							schemaFileName := simulation.GetSchemaFileName(sampleType)
							err = simulation.ValidateJSONSchema(schemaFileName, sample)
							assert.NoError(t, err, "Sample %d (%s) failed schema validation", idx, sampleType)
						})

						count++
					}

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
