//go:build integration

package tests

import (
	"encoding/json"
	"flag"
	"os"
	"testing"

	"github.com/newrelic/nri-postgresql/tests/simulation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCatalogDiscoveryMode(t *testing.T) {
	setup := `
		CREATE SCHEMA IF NOT EXISTS catalog_test;
		CREATE TABLE IF NOT EXISTS catalog_test.catalog_table (id int PRIMARY KEY, v text);
		CREATE TABLE IF NOT EXISTS catalog_test.catalog_ignored (id int PRIMARY KEY);
		CREATE TABLE IF NOT EXISTS catalog_test."catalog_owner's_table" (id int PRIMARY KEY);
		CREATE SCHEMA IF NOT EXISTS catalog_closed;
		REVOKE ALL ON SCHEMA catalog_closed FROM PUBLIC;
		CREATE TABLE IF NOT EXISTS catalog_closed.catalog_closed_table (id int PRIMARY KEY);
		CREATE TABLE IF NOT EXISTS catalog_test.catalog_parted (id int, d date) PARTITION BY RANGE (d);
		CREATE TABLE IF NOT EXISTS catalog_closed.catalog_partition PARTITION OF catalog_test.catalog_parted FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
		DO $$ BEGIN
			IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'catalog_monitor') THEN
				CREATE ROLE catalog_monitor LOGIN PASSWORD 'example';
			END IF;
		END $$;
		GRANT pg_monitor TO catalog_monitor;
		DO $$ BEGIN
			IF has_schema_privilege('catalog_monitor', 'catalog_closed', 'USAGE')
				OR EXISTS (SELECT FROM information_schema.role_table_grants WHERE grantee = 'catalog_monitor')
				OR EXISTS (SELECT FROM information_schema.role_column_grants WHERE grantee = 'catalog_monitor') THEN
				RAISE EXCEPTION 'catalog_monitor must not have table privileges or USAGE on catalog_closed';
			END IF;
		END $$;`
	_, stderr, err := simulation.ExecInContainer(serviceNamePostgresLatest, []string{"psql", "-U", *defaultUser, "-d", *defaultDB, "-v", "ON_ERROR_STOP=1", "-c", setup})
	require.NoError(t, err, stderr)

	monitorUser, monitorPassword := "catalog_monitor", "example"

	t.Run("information_schema mode does not see tables without privileges", func(t *testing.T) {
		stdout, stderr, err := simulation.RunIntegration(serviceNamePostgresLatest, integrationContainer, defaultBinaryPath, &monitorUser, &monitorPassword, defaultDB,
			`-collection_list=["demo"]`, `-metrics`)
		assert.NoError(t, err)
		assert.Empty(t, stderr)
		assert.NotContains(t, stdout, `"table:catalog_table"`)
		assert.Contains(t, stdout, `"pg-schema","Value":"pg_catalog"`)
	})

	t.Run("catalog mode sees user tables and skips system schemas", func(t *testing.T) {
		stdout, stderr, err := simulation.RunIntegration(serviceNamePostgresLatest, integrationContainer, defaultBinaryPath, &monitorUser, &monitorPassword, defaultDB,
			`-collection_list=ALL`, `-collection_discovery_mode=catalog`, `-collection_ignore_schema_list=["public"]`, `-collection_ignore_table_list=["catalog_ignored"]`, `-metrics`)
		assert.NoError(t, err)
		assert.Empty(t, stderr)
		assert.Contains(t, stdout, `"table:catalog_table"`)
		assert.Contains(t, stdout, `"index:catalog_table_pkey"`)
		assert.Contains(t, stdout, `"table:catalog_owner's_table"`)
		assert.Contains(t, stdout, `"table:catalog_closed_table"`)
		assert.Contains(t, stdout, `"index:catalog_closed_table_pkey"`)
		assert.Contains(t, stdout, `"table:catalog_partition"`)
		assert.NotNil(t, tableMetric(t, stdout, "catalog_closed_table", "table.totalSizeInBytes"))
		assert.NotContains(t, stdout, `"table:catalog_ignored"`)
		assert.NotContains(t, stdout, `"pg-schema","Value":"public"`)
		assert.NotContains(t, stdout, `"pg-schema","Value":"pg_catalog"`)
		assert.NotContains(t, stdout, `"pg-schema","Value":"information_schema"`)
	})
}

func tableMetric(t *testing.T, stdout, table, metric string) interface{} {
	t.Helper()
	var output struct {
		Data []struct {
			Entity struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"entity"`
			Metrics []map[string]interface{} `json:"metrics"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &output))
	for _, entity := range output.Data {
		if entity.Entity.Type != "pg-table" || entity.Entity.Name != table {
			continue
		}
		for _, metrics := range entity.Metrics {
			if value, ok := metrics[metric]; ok {
				return value
			}
		}
	}
	return nil
}
