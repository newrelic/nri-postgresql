package connection

import (
	"errors"
	"fmt"
	"testing"

	"github.com/newrelic/nri-postgresql/internal/args"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func Test_PGConnection_Close(t *testing.T) {
	conn, mock := CreateMockSQL(t)

	mock.ExpectClose().WillReturnError(errors.New("error"))
	conn.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("close expectation was not met: %s", err.Error())
	}
}

func Test_PGSQLConnection_Query(t *testing.T) {
	conn, mock := CreateMockSQL(t)

	// Temp data structure to store data into
	temp := []struct {
		One int `db:"one"`
		Two int `db:"two"`
	}{}

	// dummy query to run
	query := "select one, two from everywhere"

	rows := sqlmock.NewRows([]string{"one", "two"}).AddRow(1, 2)
	mock.ExpectQuery(query).WillReturnRows(rows)

	if err := conn.Query(&temp, query); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		t.FailNow()
	}

	if length := len(temp); length != 1 {
		t.Errorf("Expected 1 element got %d", length)
		t.FailNow()
	}

	if temp[0].One != 1 || temp[0].Two != 2 {
		t.Error("Query did not marshal correctly")
	}
}

func Test_PGSQLConnection_HaveExtensionInSchema(t *testing.T) {
	conn, mock := CreateMockSQL(t)

	exentionRows := sqlmock.NewRows([]string{
		"schema",
		"extension",
	}).AddRow("schema1", "extension1")
	mock.ExpectQuery(".*EXTENSIONS_LIST.*").WillReturnRows(exentionRows)

	result := conn.HaveExtensionInSchema("extension1", "schema1")
	assert.Equal(t, true, result)
}

func Test_PGSQLConnection_HaveExtensionInSchema_WithMissingExtension(t *testing.T) {
	conn, mock := CreateMockSQL(t)

	exentionRows := sqlmock.NewRows([]string{
		"schema",
		"extension",
	}).AddRow("schema1", "extension1")
	mock.ExpectQuery(".*EXTENSIONS_LIST.*").WillReturnRows(exentionRows)

	result := conn.HaveExtensionInSchema("missing", "schema1")
	assert.Equal(t, false, result)
}

func Test_PGSQLConnection_HaveExtensionInSchema_WithMissingSchema(t *testing.T) {
	conn, mock := CreateMockSQL(t)

	exentionRows := sqlmock.NewRows([]string{
		"schema",
		"extension",
	}).AddRow("schema1", "extension1")
	mock.ExpectQuery(".*EXTENSIONS_LIST.*").WillReturnRows(exentionRows)

	result := conn.HaveExtensionInSchema("extension1", "missing")
	assert.Equal(t, false, result)
}

func Test_PGSQLConnection_HaveExtensionInSchema_WithFailedQuery(t *testing.T) {
	conn, mock := CreateMockSQL(t)
	mock.ExpectQuery(".*EXTENSIONS_LIST.*").WillReturnError(fmt.Errorf("error"))

	result := conn.HaveExtensionInSchema("extension1", "missing")
	assert.Equal(t, false, result)
}

func Test_createConnectionURL(t *testing.T) {
	testCases := []struct {
		name string
		arg  *args.ArgumentList
		want string
	}{
		{
			"Base",
			&args.ArgumentList{
				Username:  "user",
				Password:  "pass",
				Hostname:  "localhost",
				EnableSSL: false,
				Port:      "5432",
				Timeout:   "30",
			},
			"postgres://user:pass@localhost:5432/postgres?connect_timeout=30&sslmode=disable",
		},
		{
			"SSL No Verify",
			&args.ArgumentList{
				Username:               "user",
				Password:               "pass",
				Hostname:               "localhost",
				EnableSSL:              true,
				TrustServerCertificate: true,
				SSLCertLocation:        "/path/cert.crt",
				SSLKeyLocation:         "/path/key.key",
				Port:                   "5432",
				Timeout:                "30",
			},
			"postgres://user:pass@localhost:5432/postgres?connect_timeout=30&sslcert=%2Fpath%2Fcert.crt&sslkey=%2Fpath%2Fkey.key&sslmode=require",
		},
		{
			"SSL Verify",
			&args.ArgumentList{
				Username:               "user",
				Password:               "pass",
				Hostname:               "localhost",
				EnableSSL:              true,
				TrustServerCertificate: false,
				SSLRootCertLocation:    "/path/to/my/cert.pem",
				SSLCertLocation:        "/path/cert.crt",
				SSLKeyLocation:         "/path/key.key",
				Port:                   "5432",
				Timeout:                "30",
			},
			"postgres://user:pass@localhost:5432/postgres?connect_timeout=30&sslcert=%2Fpath%2Fcert.crt&sslkey=%2Fpath%2Fkey.key&sslmode=verify-full&sslrootcert=%2Fpath%2Fto%2Fmy%2Fcert.pem",
		},
		{
			"SSL No Verify with Client Cert",
			&args.ArgumentList{
				Username:               "user",
				Password:               "pass",
				Hostname:               "localhost",
				EnableSSL:              true,
				TrustServerCertificate: true,
				SSLCertLocation:        "/path/cert.crt",
				SSLKeyLocation:         "/path/key.key",
				Port:                   "5432",
				Timeout:                "30",
			},
			"postgres://user:pass@localhost:5432/postgres?connect_timeout=30&sslcert=%2Fpath%2Fcert.crt&sslkey=%2Fpath%2Fkey.key&sslmode=require",
		},
		{
			"SSL Verify",
			&args.ArgumentList{
				Username:               "user",
				Password:               "pass",
				Hostname:               "localhost",
				EnableSSL:              true,
				TrustServerCertificate: false,
				SSLRootCertLocation:    "/path/to/my/cert.pem",
				SSLCertLocation:        "/path/cert.crt",
				SSLKeyLocation:         "/path/key.key",
				Port:                   "5432",
				Timeout:                "30",
			},
			"postgres://user:pass@localhost:5432/postgres?connect_timeout=30&sslcert=%2Fpath%2Fcert.crt&sslkey=%2Fpath%2Fkey.key&sslmode=verify-full&sslrootcert=%2Fpath%2Fto%2Fmy%2Fcert.pem",
		},
	}

	for _, tc := range testCases {
		if out := createConnectionURL(DefaultConnectionInfo(tc.arg).(*connectionInfo), "postgres"); out != tc.want {
			t.Errorf("Test Case %s Failed: Expected '%s' got '%s'", tc.name, tc.want, out)
		}
	}
}
