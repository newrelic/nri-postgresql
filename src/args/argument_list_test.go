package args

import (
	"testing"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name      string
		arg       *ArgumentList
		wantError bool
	}{
		{
			"No Errors",
			&ArgumentList{
				Username:       "user",
				Password:       "password",
				Hostname:       "localhost",
				Port:           "90",
				CollectionList: "{}",
			},
			false,
		},
		{
			"No Username",
			&ArgumentList{
				Username:       "",
				Password:       "password",
				Hostname:       "localhost",
				Port:           "90",
				CollectionList: "{}",
			},
			true,
		},
		{
			"No Password",
			&ArgumentList{
				Username:       "user",
				Hostname:       "localhost",
				Port:           "90",
				CollectionList: "{}",
			},
			true,
		},
		{
			"SSL and No Server Certificate",
			&ArgumentList{
				Username:               "user",
				Password:               "password",
				Hostname:               "localhost",
				Port:                   "90",
				EnableSSL:              true,
				TrustServerCertificate: false,
				SSLRootCertLocation:    "",
				CollectionList:         "{}",
			},
			true,
		},
		{
			"Missing Key file with Cert file",
			&ArgumentList{
				Username:               "user",
				Password:               "password",
				Hostname:               "localhost",
				Port:                   "90",
				EnableSSL:              true,
				TrustServerCertificate: true,
				SSLKeyLocation:         "",
				SSLCertLocation:        "my.crt",
				CollectionList:         "{}",
			},
			false,
		},
		{
			"Missing Cert file with Key file",
			&ArgumentList{
				Username:               "user",
				Password:               "password",
				Hostname:               "localhost",
				Port:                   "90",
				EnableSSL:              true,
				TrustServerCertificate: true,
				SSLKeyLocation:         "my.key",
				SSLCertLocation:        "",
				CollectionList:         "{}",
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.arg.Validate()
		if tc.wantError && err == nil {
			t.Errorf("Test Case %s Failed: Expected error", tc.name)
		} else if !tc.wantError && err != nil {
			t.Errorf("Test Case %s Failed: Unexpected error: %v", tc.name, err)
		}
	}
}
