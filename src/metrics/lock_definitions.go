package metrics

import (
	"github.com/blang/semver"
	"github.com/newrelic/nri-postgresql/src/collection"
)

func generateLockDefinitions(databases collection.DatabaseList, version *semver.Version) []*QueryDefinition {
	queryDefinitions := make([]*QueryDefinition, 0, 1)
	if len(databases) == 0 {
		return queryDefinitions
	}

	queryDefinitions = append(queryDefinitions, lockDefinitions.insertDatabaseNames(databases))

	return queryDefinitions
}

var lockDefinitions = &QueryDefinition{
	query: `SELECT -- LOCKS_DEFINITION
                 database,
                 COALESCE(access_exclusive_lock, 0) AS access_exclusive_lock,
                 COALESCE(access_share_lock, 0) AS access_share_lock,
                 COALESCE(exclusive_lock, 0) AS exclusive_lock,
                 COALESCE(row_exclusive_lock, 0) AS row_exclusive_lock,
                 COALESCE(row_share_lock, 0) AS row_share_lock,
                 COALESCE(share_lock, 0) AS share_lock,
                 COALESCE(share_row_exclusive_lock, 0) AS share_row_exclusive_lock,
                 COALESCE(share_update_exclusive_lock, 0) AS share_update_exclusive_lock
            FROM public.crosstab(
                  $$SELECT psa.datname AS database,
                           lock.mode,
                           count(lock.mode)
                     FROM pg_locks AS lock
                LEFT JOIN pg_stat_activity AS psa ON lock.pid = psa.pid
                    WHERE psa.datname IN (%DATABASES%)
                 GROUP BY lock.database, lock.mode, psa.datname
                 ORDER BY database,mode$$,
                 $$VALUES ('AccessExclusiveLock'::text),
                          ('AccessShareLock'::text),
                          ('ExclusiveLock'::text),
                          ('RowExclusiveLock'::text),
                          ('RowShareLock'::text),
                          ('ShareLock'::text),
                          ('ShareRowExclusiveLock'::text),
                          ('ShareUpdateExclusiveLock'::text) $$
           ) AS data (
                 database text,
                 access_exclusive_lock numeric,
                 access_share_lock numeric,
                 exclusive_lock numeric,
                 row_exclusive_lock numeric,
                 row_share_lock numeric,
                 share_lock numeric,
                 share_row_exclusive_lock numeric,
                 share_update_exclusive_lock numeric
          );`,
	dataModels: []struct {
		databaseBase
		AccessExclusiveLock      *int `db:"access_exclusive_lock" metric_name:"db.locks.accessExclusiveLock" source_type:"gauge"`
		AccessShareLock          *int `db:"access_share_lock" metric_name:"db.locks.accessShareLock" source_type:"gauge"`
		ExclusiveLock            *int `db:"exclusive_lock" metric_name:"db.locks.exclusiveLock" source_type:"gauge"`
		RowExclusiveLock         *int `db:"row_exclusive_lock" metric_name:"db.locks.rowExclusiveLock" source_type:"gauge"`
		RowShareLock             *int `db:"row_share_lock" metric_name:"db.locks.rowShareLock" source_type:"gauge"`
		ShareLock                *int `db:"share_lock" metric_name:"db.locks.shareLock" source_type:"gauge"`
		ShareRowExclusiveLock    *int `db:"share_row_exclusive_lock" metric_name:"db.locks.shareRowExclusiveLock" source_type:"gauge"`
		ShareUpdateExclusiveLock *int `db:"share_update_exclusive_lock" metric_name:"db.locks.shareUpdateExclusiveLock" source_type:"gauge"`
	}{},
}
