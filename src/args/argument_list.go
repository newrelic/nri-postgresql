// Package args contains the argument list, defined as a struct, along with a method that validates passed-in args
package args

import (
	"errors"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
)

// ArgumentList struct that holds all PostgreSQL arguments
type ArgumentList struct {
	sdkArgs.DefaultArgumentList
	Username               string `default:"" help:"The username for the PostgreSQL database"`
	Password               string `default:"" help:"The password for the specified username"`
	Hostname               string `default:"localhost" help:"The PostgreSQL hostname to connect to"`
	Port                   string `default:"5432" help:"The port to connect to the PostgreSQL database"`
	EnableSSL              bool   `default:"false" help:"If true will use SSL encryption, false will not use encryption"`
	TrustServerCertificate bool   `default:"false" help:"If true server certificate is not verified for SSL. If false certificate will be verified against supplied certificate"`
	SSLRootCertLocation    string `default:"" help:"Absolute path to PEM encoded root certificate file"`
	SSLCertLocation        string `default:"" help:"Absolute path to PEM encoded client cert file"`
	SSLKeyLocation         string `default:"" help:"Absolute path to PEM encoded client key file"`
	Timeout                string `default:"10" help:"Maximum wait for connection, in seconds. Set 0 for no timeout"`
}

// Validate validates PostgreSQl arguments
func (al ArgumentList) Validate() error {
	if al.Username == "" || al.Password == "" {
		return errors.New("invalid configuration: must specify a username and password")
	}

	if al.EnableSSL {
		if !al.TrustServerCertificate && al.SSLRootCertLocation == "" {
			return errors.New("invalid configuration: must specify a certificate file when using SSL and not trusting server certificate")
		} else if (al.SSLCertLocation != "" && al.SSLKeyLocation == "") || (al.SSLKeyLocation != "" && al.SSLCertLocation == "") {
			return errors.New("invalid configuration: must specify both a cert and key file")
		}
	}

	return nil
}
