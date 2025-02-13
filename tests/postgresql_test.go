//go:build integration

package tests

import (
	"flag"
	"os"
	"testing"

	"github.com/newrelic/nri-postgresql/tests/simulation"
	"github.com/stretchr/testify/assert"
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
	defaultBinaryPath         = "/nri-postgresql"
	integrationContainer      = "nri-postgresql"
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
			stdout, stderr, err := simulation.RunIntegration(tc.Hostname, integrationContainer, defaultBinaryPath, defaultUser, defaultPassword, defaultDB, args...)
			assert.Empty(t, stderr)
			assert.NoError(t, err)
			assert.NotEmpty(t, stdout)
			err = simulation.ValidateJSONSchema(tc.Schema, stdout)
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

	_, stderr, err := simulation.RunIntegration(serviceNamePostgresLatest, integrationContainer, defaultBinaryPath, defaultUser, defaultPassword, defaultDB)
	assert.Error(t, err)
	assert.Contains(t, stderr, "invalid configuration: must specify a username and password")
}

func TestIgnoringDB(t *testing.T) {
	args := []string{
		`-collection_list=all`,
		`-collection_ignore_database_list=["demo"]`,
	}
	stdout, stderr, err := simulation.RunIntegration(serviceNamePostgresLatest, integrationContainer, defaultBinaryPath, defaultUser, defaultPassword, defaultDB, args...)
	assert.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, `"database:postgres"`)
	assert.NotContains(t, stdout, `"database:demo"`)
}
