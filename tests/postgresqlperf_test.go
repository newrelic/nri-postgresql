//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/tests/simulation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
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
					controller := NewSimulationController(container)
					// Start all simulations
					done := controller.StartAllSimulations(t)

					time.Sleep(30 * time.Second)

					stdout := runIntegration(t, container, tt.args...)
					// fmt.Println(stdout)
					samples := strings.Split(stdout, "\n")
					count := 0

					for idx, sample := range samples {
						sample = strings.TrimSpace(sample)
						if sample == "" {
							continue
						}

						// Validate sample type
						t.Run(fmt.Sprintf("Validating JSON schema for: %s", tt.expectedOrder[count]), func(t *testing.T) {
							// Validate JSON
							var j map[string]interface{}
							err := json.Unmarshal([]byte(sample), &j)
							assert.NoError(t, err, "Sample %d - Integration Output Is An Invalid JSONs", idx)

							sampleType := tt.expectedOrder[count]
							if !strings.Contains(sample, sampleType) {
								t.Errorf("Integration output does not contain: %s", tt.expectedOrder[count])
							}

							// Validate against schema
							schemaFileName := getSchemaFileName(sampleType)
							err = validateJSONSchema(schemaFileName, sample)
							assert.NoError(t, err, "Sample %d (%s) failed schema validation", idx, sampleType)
						})

						count++
					}
					fmt.Println(count)

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
					stdout := runIntegration(t, container, tt.args...)

					// Validate JSON format
					var j map[string]interface{}
					err := json.Unmarshal([]byte(stdout), &j)
					assert.NoError(t, err, "Integration Output Is An Invalid JSON")

					// Verify it's a PostgresqlInstanceSample
					assert.Contains(t, stdout, "PostgresqlInstanceSample",
						"Integration output does not contain PostgresqlInstanceSample")

					// Validate against schema
					err = validateJSONSchema("jsonschema-latest.json", stdout)
					assert.NoError(t, err, "Output failed schema validation")
				})
			}
		})
	}
}

// ---------------- HELPER FUNCTIONS ----------------

func openDB(targetContainer string) (*sqlx.DB, error) {
	// Use the container-specific port for database connections
	dbPort := getPortForContainer(targetContainer)

	connectionURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(*user, *psw),
		Host:   fmt.Sprintf("%s:%d", "localhost", dbPort),
		Path:   testDB,
	}

	query := url.Values{}
	query.Add("connect_timeout", "10")
	query.Add("sslmode", "disable")

	connectionURL.RawQuery = query.Encode()
	dsn := connectionURL.String()
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to db: %s", err) //nolint:all
	}
	return db, nil
}

func getPortForContainer(container string) int {
	switch container {
	case "postgresql-perf-latest":
		return 5432
	case "postgresql-perf-oldest":
		return 6432
	case "postgresql-noext":
		return 7432
	default:
		return 5432
	}
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
	// fmt.Printf("executing: docker %s", strings.Join(cmdLine, " "))

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
	// fmt.Println(stdout)

	return stdout, stderr, nil
}

func runIntegration(t *testing.T, targetContainer string, integration_args ...string) string {
	t.Helper()

	command := make([]string, 0)
	command = append(command, *binaryPath)

	if user != nil {
		command = append(command, "-username", *user)
	}
	if psw != nil {
		command = append(command, "-password", *psw)
	}

	// Always use port 5432 for integration runs
	command = append(command, "-port", "5432")

	if database != nil {
		command = append(command, "-database", *database)
	}
	if targetContainer != "" {
		command = append(command, "-hostname", targetContainer)
	}

	for _, arg := range integration_args {
		command = append(command, arg) //nolint:all
	}

	stdout, stderr, err := ExecInContainer(*container, command)
	if stderr != "" {
		log.Debug("Integration command Standard Error: ", stderr)
	}
	fmt.Println(stderr)
	require.NoError(t, err)

	return stdout
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
		return fmt.Errorf("Error loading JSON schema, error: %v", err) //nolint:all
	}

	if result.Valid() {
		return nil
	}
	fmt.Printf("Errors for JSON schema: '%s'\n", schemaURI)
	for _, desc := range result.Errors() {
		fmt.Printf("\t- %s\n", desc)
	}
	fmt.Printf("\n")
	return fmt.Errorf("The output of the integration doesn't have expected JSON format") //nolint:all
}

func getSchemaFileName(sampleType string) string {
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

// ---------------- DB SIMULATION FUNCTIONS ----------------

// SimulationController handles coordinating multiple database simulations
type SimulationController struct {
	targetContainer string
	envVars         []string
}

// NewSimulationController creates a new controller for database simulations
func NewSimulationController(targetContainer string, envVars ...string) *SimulationController {
	return &SimulationController{
		targetContainer: targetContainer,
		envVars:         envVars,
	}
}

// StartAllSimulations starts all simulation routines concurrently
func (sc *SimulationController) StartAllSimulations(t *testing.T) chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		// Create error channel to collect errors from goroutines
		errChan := make(chan error, 6)

		// Start all simulations in separate goroutines
		go func() {
			SimulateQueries(t, sc.targetContainer)
			errChan <- nil
		}()

		go func() {
			SimulateSlowQueries(t, sc.targetContainer)
			errChan <- nil
		}()

		for pclass := 1; pclass <= 3; pclass++ {
			go func() {
				SimulateWaitEvents(t, sc.targetContainer, pclass)
				errChan <- nil
			}()
		}

		go func() {
			SimulateBlockingSessions(t, sc.targetContainer)
			errChan <- nil
		}()

		// Wait for all goroutines to complete
		for i := 0; i < 6; i++ {
			if err := <-errChan; err != nil {
				log.Error("Error in simulation routine: %v", err)
				t.Error(err)
			}
		}
	}()

	return done
}

func ExecuteQuery(t *testing.T, query string, targetContainer string, delay int) {

	db, err := openDB(targetContainer)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(query)
	// fmt.Println(stderr)
	require.NoError(t, err)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

func SimulateQueries(t *testing.T, targetContainer string) {
	t.Helper()
	for _, query := range simulation.SimpleQueries() {
		ExecuteQuery(t, query, targetContainer, 100)
	}
}

func SimulateSlowQueries(t *testing.T, targetContainer string) {
	t.Helper()
	for _, query := range simulation.SlowQueries() {
		ExecuteQuery(t, query, targetContainer, 500)
	}
}

func SimulateWaitEvents(t *testing.T, targetContainer string, pclass int) {
	t.Helper()

	queries := simulation.WaitEventQueries(pclass)

	// Start the locking transaction in a goroutine
	go func() {
		ExecuteQuery(t, queries.LockingQuery, targetContainer, 100)
	}()

	// Wait for first transaction started
	time.Sleep(2 * time.Second)

	// Run the blocked transaction
	ExecuteQuery(t, queries.BlockedQuery, targetContainer, 100)

	time.Sleep(30 * time.Second)
}

func SimulateBlockingSessions(t *testing.T, targetContainer string) {
	t.Helper()

	queries := simulation.BlockingQueries()

	db, err := openDB(targetContainer)
	require.NoError(t, err)
	defer db.Close()

	// Start the first transaction that will hold the lock
	tx1, err := db.Begin()
	require.NoError(t, err)
	defer tx1.Rollback() //nolint:all

	// Execute the locking query
	_, err = tx1.Exec(queries.HoldLockQuery)
	require.NoError(t, err)

	// Start the blocking query in a separate goroutine
	go func() {
		time.Sleep(2 * time.Second) // Wait for a bit before trying to acquire lock

		tx2, err := db.Begin()
		if err != nil {
			t.Error(err)
			return
		}
		defer tx2.Rollback() //nolint:all

		// This query will block waiting for tx1's lock
		tx2.Exec(queries.BlockedQuery) //nolint:all
		// We don't check for errors here since this might timeout
	}()

	// Hold the lock for a few seconds, then release it
	time.Sleep(5 * time.Second)
	tx1.Commit() //nolint:all
}
