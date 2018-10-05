// Package connection contains the PGSQLConnection type and methods for manipulating and querying a PostgreSQL connection
package connection

import (
	"fmt"
	"net/url"

	"github.com/jmoiron/sqlx"
	// pq is required for postgreSQL driver but isn't used in code
	_ "github.com/lib/pq"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-postgresql/src/args"
)

// PGSQLConnection represents a wrapper around a PostgreSQL connection
type PGSQLConnection struct {
	connection *sqlx.DB
}

// Info holds all the information needed from the user to create a new connection
type Info interface {
	NewConnection(database string) (*PGSQLConnection, error)
}

type connectionInfo struct {
	Database               string
	Username               string
	Password               string
	Host                   string
	Port                   string
	Timeout                string
	EnableSSL              bool
	SSLCertLocation        string
	SSLRootCertLocation    string
	SSLKeyLocation         string
	TrustServerCertificate bool
}

// DefaultConnectionInfo takes an argument list and constructs a default connection out of it
func DefaultConnectionInfo(al *args.ArgumentList) Info {
	return &connectionInfo{
		Database:               "postgres",
		Username:               al.Username,
		Password:               al.Password,
		Host:                   al.Hostname,
		Port:                   al.Port,
		Timeout:                al.Timeout,
		EnableSSL:              al.EnableSSL,
		SSLCertLocation:        al.SSLCertLocation,
		SSLRootCertLocation:    al.SSLRootCertLocation,
		SSLKeyLocation:         al.SSLKeyLocation,
		TrustServerCertificate: al.TrustServerCertificate,
	}
}

// NewConnection creates a new PGSQLConnection from args
func (ci *connectionInfo) NewConnection(database string) (*PGSQLConnection, error) {
	ci.Database = database
	db, err := sqlx.Connect("postgres", createConnectionURL(ci))
	if err != nil {
		return nil, err
	}

	return &PGSQLConnection{
		connection: db,
	}, nil
}

// Close closes the PosgreSQL connection. If an error occurs
// it is logged as a warning.
func (p PGSQLConnection) Close() {
	if err := p.connection.Close(); err != nil {
		log.Warn("Unable to close PostgreSQL Connection: %s", err.Error())
	}
}

// Query runs a query and loads results into v
func (p PGSQLConnection) Query(v interface{}, query string) error {
	return p.connection.Select(v, query)
}

// createConnectionURL creates the connection string. A list of paramters
// can be found here https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
func createConnectionURL(ci *connectionInfo) string {
	connectionURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(ci.Username, ci.Password),
		Host:   fmt.Sprintf("%s:%s", ci.Host, ci.Port),
		Path:   ci.Database,
	}

	query := url.Values{}
	query.Add("connect_timeout", ci.Timeout)

	// SSL settings
	if ci.EnableSSL {
		addSSLQueries(query, ci)
	} else {
		query.Add("sslmode", "disable")
	}

	connectionURL.RawQuery = query.Encode()

	return connectionURL.String()
}

// addSSLQueries add SSL query parameters
func addSSLQueries(query url.Values, ci *connectionInfo) {
	if ci.TrustServerCertificate {
		query.Add("sslmode", "require")
	} else {
		query.Add("sslmode", "verify-full")
		query.Add("sslrootcert", ci.SSLRootCertLocation)
	}

	if ci.SSLCertLocation != "" && ci.SSLKeyLocation != "" {
		query.Add("sslcert", ci.SSLCertLocation)
		query.Add("sslkey", ci.SSLKeyLocation)
	}
}
