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

// NewConnection creates a new PGSQLConnection from args
func NewConnection(args *args.ArgumentList) (*PGSQLConnection, error) {
	db, err := sqlx.Connect("postgres", createConnectionURL(args))
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
func createConnectionURL(args *args.ArgumentList) string {
	connectionURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(args.Username, args.Password),
		Host:   fmt.Sprintf("%s:%s", args.Hostname, args.Port),
		Path:   "postgres",
	}

	query := url.Values{}
	query.Add("connect_timeout", args.Timeout)

	// SSL settings
	if args.EnableSSL {
		addSSLQueries(query, args)
	} else {
		query.Add("sslmode", "disable")
	}

	connectionURL.RawQuery = query.Encode()

	return connectionURL.String()
}

// addSSLQueries add SSL query parameters
func addSSLQueries(query url.Values, args *args.ArgumentList) {
	if args.TrustServerCertificate {
		query.Add("sslmode", "require")
	} else {
		query.Add("sslmode", "verify-full")
		query.Add("sslrootcert", args.SSLRootCertLocation)
	}

	if args.SSLCertLocation != "" && args.SSLKeyLocation != "" {
		query.Add("sslcert", args.SSLCertLocation)
		query.Add("sslkey", args.SSLKeyLocation)
	}
}
