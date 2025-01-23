package globalvariables

import (
	"github.com/newrelic/nri-postgresql/src/args"
)

type GlobalVariables struct {
	QueryResponseTimeThreshold int
	QueryCountThreshold        int
	Version                    uint64
	DatabaseString             string
	Hostname                   string
	Port                       string
	Arguments                  args.ArgumentList
}

func SetGlobalVariables(args args.ArgumentList, version uint64, databaseString string) *GlobalVariables {
	return &GlobalVariables{
		QueryResponseTimeThreshold: args.QueryResponseTimeThreshold,
		QueryCountThreshold:        args.QueryCountThreshold,
		Version:                    version,
		DatabaseString:             databaseString,
		Hostname:                   args.Hostname,
		Port:                       args.Port,
		Arguments:                  args,
	}
}
