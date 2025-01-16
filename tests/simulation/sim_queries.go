package simulation

import "fmt"

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
