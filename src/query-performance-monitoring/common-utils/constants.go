package commonutils

import "errors"

// The maximum number of metrics to be published in a single batch
const PublishThreshold = 600
const RandomIntRange = 1000000
const TimeFormat = "20060102150405"

// The maximum number of individual queries that can be fetched in a single metrics, the value was chosen as the queries samples were with same query statements but with different parameters so 10 samples would be enough to check the execution plan
const MaxIndividualQueryCountThreshold = 10

var ErrUnsupportedVersion = errors.New("unsupported PostgreSQL version")
var ErrUnExpectedError = errors.New("unexpected error")

var ErrInvalidModelType = errors.New("invalid model type")
var ErrNotEligible = errors.New("not Eligible to fetch metrics")

const PostgresVersion12 = 12
const PostgresVersion11 = 11
const PostgresVersion13 = 13
const PostgresVersion14 = 14

const PgStatStatementExtension = "pg_stat_statements"
const PgStatMonitorExtension = "pg_stat_monitor"
const PgWaitSamplingExtension = "pg_wait_sampling"
