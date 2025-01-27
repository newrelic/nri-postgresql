package commonutils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/newrelic/nri-postgresql/src/collection"
)

// re is a regular expression that matches single-quoted strings, numbers, or double-quoted strings
var re = regexp.MustCompile(`'[^']*'|\d+|".*?"`)

func GetQuotedStringFromArray(array []string) string {
	var quotedNames = make([]string, 0)
	for _, name := range array {
		quotedNames = append(quotedNames, fmt.Sprintf("'%s'", name))
	}
	return strings.Join(quotedNames, ",")
}

func GetDatabaseListInString(dbMap collection.DatabaseList) string {
	var databaseNames = make([]string, 0)
	for dbName := range dbMap {
		databaseNames = append(databaseNames, dbName)
	}
	if len(databaseNames) == 0 {
		return ""
	}
	return GetQuotedStringFromArray(databaseNames)
}

func AnonymizeQueryText(query string) string {
	anonymizedQuery := re.ReplaceAllString(query, "?")
	return anonymizedQuery
}

// This function is used to generate a unique plan ID for a query
func GeneratePlanID(queryID string) *string {
	randomInt, err := rand.Int(rand.Reader, big.NewInt(RandomIntRange))
	if err != nil {
		return nil
	}
	currentTime := time.Now().Format(TimeFormat)
	result := fmt.Sprintf("%s-%d-%s", queryID, randomInt.Int64(), currentTime)
	return &result
}
