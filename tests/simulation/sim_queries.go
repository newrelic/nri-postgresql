//nolint:all
package simulation

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/stretchr/testify/require"
)

const (
	defaultUser = "dbuser"
	defaultPass = "dbpassword"
	testDB      = "titanic"
)

func openDB(targetContainer string) (*sqlx.DB, error) {
	dbPort := getPortForContainer(targetContainer)

	connectionURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(defaultUser, defaultPass),
		Host:   fmt.Sprintf("%s:%d", "localhost", dbPort),
		Path:   testDB,
	}

	query := url.Values{}
	query.Add("connect_timeout", "10")
	query.Add("sslmode", "disable")

	connectionURL.RawQuery = query.Encode()
	dsn := connectionURL.String()
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to db: %s", err)
	}
	return db, nil
}

func getPortForContainer(container string) int {
	switch container {
	case "postgresql-perf-latest":
		return 5432
	case "postgresql-perf-oldest":
		return 6432
	case "postgresql-noext":
		return 7432
	default:
		return 5432
	}
}

// SimulationController handles coordinating multiple database simulations
type SimulationController struct {
	targetContainer string
	envVars         []string
}

// NewSimulationController creates a new controller for database simulations
func NewSimulationController(targetContainer string, envVars ...string) *SimulationController {
	return &SimulationController{
		targetContainer: targetContainer,
		envVars:         envVars,
	}
}

// StartAllSimulations starts all simulation routines concurrently
func (sc *SimulationController) StartAllSimulations(t *testing.T) chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		// Create error channel to collect errors from goroutines
		errChan := make(chan error, 6)

		// Start all simulations in separate goroutines
		go func() {
			SimulateQueries(t, sc.targetContainer)
			errChan <- nil
		}()

		go func() {
			SimulateSlowQueries(t, sc.targetContainer)
			errChan <- nil
		}()

		for pclass := 1; pclass <= 3; pclass++ {
			go func() {
				SimulateWaitEvents(t, sc.targetContainer, pclass)
				errChan <- nil
			}()
		}

		go func() {
			SimulateBlockingSessions(t, sc.targetContainer)
			errChan <- nil
		}()

		// Wait for all goroutines to complete
		for i := 0; i < 6; i++ {
			if err := <-errChan; err != nil {
				log.Error("Error in simulation routine: %v", err)
				t.Error(err)
			}
		}
	}()

	return done
}

func ExecuteQuery(t *testing.T, query string, targetContainer string, delay int) {
	db, err := openDB(targetContainer)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(query)
	require.NoError(t, err)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

func SimulateQueries(t *testing.T, targetContainer string) {
	t.Helper()
	for _, query := range SimpleQueries() {
		ExecuteQuery(t, query, targetContainer, 100)
	}
}

func SimulateSlowQueries(t *testing.T, targetContainer string) {
	t.Helper()
	for _, query := range SlowQueries() {
		ExecuteQuery(t, query, targetContainer, 500)
	}
}

func SimulateWaitEvents(t *testing.T, targetContainer string, pclass int) {
	t.Helper()

	queries := WaitEventQueries(pclass)

	// Start the locking transaction in a goroutine
	go func() {
		ExecuteQuery(t, queries.LockingQuery, targetContainer, 100)
	}()

	// Wait for first transaction started
	time.Sleep(2 * time.Second)

	// Run the blocked transaction
	ExecuteQuery(t, queries.BlockedQuery, targetContainer, 100)

	time.Sleep(30 * time.Second)
}

func SimulateBlockingSessions(t *testing.T, targetContainer string) {
	t.Helper()

	queries := BlockingQueries()

	db, err := openDB(targetContainer)
	require.NoError(t, err)
	defer db.Close()

	// Start the first transaction that will hold the lock
	tx1, err := db.Begin()
	require.NoError(t, err)
	defer tx1.Rollback()

	// Execute the locking query
	_, err = tx1.Exec(queries.HoldLockQuery)
	require.NoError(t, err)

	// Start the blocking query in a separate goroutine
	go func() {
		time.Sleep(2 * time.Second) // Wait for a bit before trying to acquire lock

		tx2, err := db.Begin()
		if err != nil {
			t.Error(err)
			return
		}
		defer tx2.Rollback()

		// This query will block waiting for tx1's lock
		tx2.Exec(queries.BlockedQuery)
		// We don't check for errors here since this might timeout
	}()

	// Hold the lock for a few seconds, then release it
	time.Sleep(5 * time.Second)
	tx1.Commit()
}

func SimpleQueries() []string {
	return []string{
		// Basic queries that will generate typical workload
		"SELECT COUNT(*) FROM passenger WHERE survived = 1",
		"SELECT class, COUNT(*) FROM passenger GROUP BY class",
		"SELECT * FROM passenger WHERE fare > 100 ORDER BY fare DESC LIMIT 10",
		"SELECT sex, AVG(age) as avg_age FROM passenger GROUP BY sex",
		"SELECT * FROM passenger WHERE name LIKE '%John%'",
	}
}

func SlowQueries() []string {
	return []string{
		// Age-based survival analysis
		`WITH age_groups AS (
            SELECT 
                CASE 
                    WHEN age < 18 THEN 'child'
                    WHEN age < 50 THEN 'adult'
                    ELSE 'elderly'
                END as age_group,
                survived
            FROM passenger
        )
        SELECT 
            age_group,
            COUNT(*) as total,
            SUM(survived::int) as survived_count,
            ROUND(AVG(survived::int) * 100, 2) as survival_rate
        FROM age_groups
        GROUP BY age_group
        ORDER BY survival_rate DESC`,

		// Multiple self-joins analysis
		`SELECT DISTINCT p1.name, p1.class, p1.fare
         FROM passenger p1
         JOIN passenger p2 ON p1.fare = p2.fare AND p1.passengerid != p2.passengerid
         JOIN passenger p3 ON p2.class = p3.class AND p2.passengerid != p3.passengerid
         WHERE p1.survived = 1
         ORDER BY p1.fare DESC`,

		// Subquery with expensive sort
		`SELECT *, 
            (SELECT COUNT(*) 
             FROM passenger p2 
             WHERE p2.fare > p1.fare) as more_expensive_tickets
         FROM passenger p1
         ORDER BY more_expensive_tickets DESC`,

		// Complex aggregation with JSON
		`SELECT 
            p1.class,
            p1.survived,
            COUNT(*) as group_size,
            AVG(p1.age) as avg_age,
            STRING_AGG(DISTINCT p1.name, ', ' ORDER BY p1.name) as passenger_names,
            (
                SELECT JSON_AGG(
                    JSON_BUILD_OBJECT(
                        'name', p2.name,
                        'fare', p2.fare
                    )
                )
                FROM passenger p2
                WHERE p2.class = p1.class
                AND p2.survived = p1.survived
            ) as similar_fare_passengers
        FROM passenger p1
        GROUP BY p1.class, p1.survived
        ORDER BY p1.class, p1.survived`,

		// Statistical analysis with percentiles
		`SELECT 
            p1.class,
            p1.sex,
            PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY p1.age) as median_age,
            PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY p1.fare) as median_fare,
            COUNT(*) FILTER (WHERE p1.survived = 1)::float / COUNT(*) as survival_rate,
            (
                SELECT array_agg(DISTINCT p2.embarked)
                FROM passenger p2
                WHERE p2.class = p1.class AND p2.sex = p1.sex
            ) as embarkation_points
        FROM passenger p1
        GROUP BY p1.class, p1.sex
        HAVING COUNT(*) > 10
        ORDER BY p1.class, p1.sex`,

		// Decile analysis with window functions
		`WITH fare_ranks AS (
            SELECT 
                *,
                NTILE(10) OVER (ORDER BY fare) as fare_decile
            FROM passenger
        ),
        age_ranks AS (
            SELECT 
                *,
                NTILE(10) OVER (ORDER BY age) as age_decile
            FROM passenger
            WHERE age IS NOT NULL
        )
        SELECT 
            fr.fare_decile,
            ar.age_decile,
            COUNT(*) as passenger_count,
            SUM(fr.survived::int) as survivors,
            AVG(fr.fare) as avg_fare,
            AVG(ar.age) as avg_age,
            array_agg(DISTINCT fr.class) as class_distribution
        FROM fare_ranks fr
        JOIN age_ranks ar ON fr.passengerid = ar.passengerid
        GROUP BY fr.fare_decile, ar.age_decile
        ORDER BY fr.fare_decile, ar.age_decile`,
	}
}

func BlockingQueries() struct {
	HoldLockQuery string
	BlockedQuery  string
} {
	return struct {
		HoldLockQuery string
		BlockedQuery  string
	}{
		HoldLockQuery: `
BEGIN;
SELECT * FROM passenger WHERE passengerid = 100 FOR UPDATE;
`,
		BlockedQuery: `
BEGIN;
SELECT * FROM passenger WHERE passengerid = 100 FOR UPDATE;
`,
	}
}

func WaitEventQueries(pclass int) struct {
	LockingQuery string
	BlockedQuery string
} {
	return struct {
		LockingQuery string
		BlockedQuery string
	}{
		LockingQuery: fmt.Sprintf(`
BEGIN;
UPDATE passenger
SET fare = fare * 1.01 
WHERE pclass = %d;
SELECT pg_sleep(30);
COMMIT;`, pclass),

		BlockedQuery: fmt.Sprintf(`
BEGIN;
UPDATE passenger
SET fare = fare * 0.99 
WHERE pclass = %d;
COMMIT;`, pclass),
	}
}
