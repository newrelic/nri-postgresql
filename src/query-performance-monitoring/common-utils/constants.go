package commonutils

import "errors"

const MaxQueryThreshold = 30
const MaxIndividualQueryThreshold = 10
const PublishThreshold = 100
const RandomIntRange = 1000000
const TimeFormat = "20060102150405"
const VersionRegex = "PostgreSQL (\\d+)\\."

var ErrParseVersion = errors.New("unable to parse PostgreSQL version from string")
var ErrUnsupportedVersion = errors.New("unsupported PostgreSQL version")
var ErrUnExpectedError = errors.New("unexpected error")

var ErrVersionFetchError = errors.New("no rows returned from version query")
var ErrInvalidModelType = errors.New("invalid model type")
var ErrNotEligible = errors.New("not Eligible to fetch metrics")

const PostgresVersion12 = 12
const PostgresVersion11 = 11
const PostgresVersion13 = 13
const PostgresVersion14 = 14
const VersionIndex = 2
