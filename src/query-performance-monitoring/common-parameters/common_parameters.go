package commonparameters

import (
	"strings"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
)

// The maximum number records that can be fetched in a single metrics
const MaxQueryCountThreshold = 30

// DefaultQueryMonitoringCountThreshold is the default threshold for the number of queries to monitor.
const DefaultQueryMonitoringCountThreshold = 20

// DefaultQueryResponseTimeThreshold is the default threshold for the response time of a query.
const DefaultQueryResponseTimeThreshold = 500

type CommonParameters struct {
	Version                              uint64
	Databases                            string
	QueryMonitoringCountThreshold        int
	QueryMonitoringResponseTimeThreshold int
	Host                                 string
	Port                                 string
	IsRds                                bool
}

func SetCommonParameters(args args.ArgumentList, version uint64, databases string) *CommonParameters {
	return &CommonParameters{
		Version:                              version,
		Databases:                            databases, // comma separated database names
		QueryMonitoringCountThreshold:        validateAndGetQueryMonitoringCountThreshold(args),
		QueryMonitoringResponseTimeThreshold: validateAndGetQueryMonitoringResponseTimeThreshold(args),
		Host:                                 args.Hostname,
		Port:                                 args.Port,
		IsRds:                                isRdsHost(args.Hostname),
	}
}

// A reliable way to identify hostnames associated with Amazon Relational Database Service (RDS) instances. Amazon uses this domain suffix for RDS endpoints across various database engines.
func isRdsHost(hostname string) bool {
	return strings.Contains(hostname, ".rds.amazonaws.com")
}

func validateAndGetQueryMonitoringResponseTimeThreshold(args args.ArgumentList) int {
	if args.QueryMonitoringResponseTimeThreshold < 0 {
		log.Warn("QueryResponseTimeThreshold should be greater than or equal to 0 but the input is %d, setting value to default which is %d", args.QueryMonitoringResponseTimeThreshold, DefaultQueryResponseTimeThreshold)
		return DefaultQueryResponseTimeThreshold
	}
	return args.QueryMonitoringResponseTimeThreshold
}

func validateAndGetQueryMonitoringCountThreshold(args args.ArgumentList) int {
	if args.QueryMonitoringCountThreshold < 0 {
		log.Warn("QueryCountThreshold should be greater than 0 but the input is %d, setting value to default which is %d", args.QueryMonitoringCountThreshold, DefaultQueryMonitoringCountThreshold)
		return DefaultQueryMonitoringCountThreshold
	}
	if args.QueryMonitoringCountThreshold > MaxQueryCountThreshold {
		log.Warn("QueryCountThreshold should be less than or equal to max limit but the input is %d, setting value to max limit which is %d", args.QueryMonitoringCountThreshold, MaxQueryCountThreshold)
		return MaxQueryCountThreshold
	}
	return args.QueryMonitoringCountThreshold
}
