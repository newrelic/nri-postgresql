package globalvariables

import (
	"github.com/newrelic/nri-postgresql/src/args"
)

type GlobalVariables struct {
	Version        uint64
	DatabaseString string
	Arguments      args.ArgumentList
}

func SetGlobalVariables(args args.ArgumentList, version uint64, databaseString string) *GlobalVariables {
	return &GlobalVariables{
		Version:        version,
		DatabaseString: databaseString,
		Arguments:      args,
	}
}
