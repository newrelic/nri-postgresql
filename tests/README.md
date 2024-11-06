# Integration tests

Steps to update the integration tests for the latest supported version:

1. Update the postgres image in the `postgres-latest-supported` of the [docker compose](./docker-compose.yml).
2. Execute the integration tests
    * If the JSON-schema validation fails:
        - Check the inventory, some server settings might have been removed.
        - Check the number of entities: the number of internal tables and or indexes may vary (metrics failures).
        - Check the release notes ([Postgres 16 example](https://www.postgresql.org/docs/release/16.0/))
3. Once the failures are understood (if any), update the corresponding JSON-schema files, you may need to generate it
   using the integration output, specially if there is any metric failure.

# Testing pgbouncer upgrades

Steps to test breaking metrics changes that happen when new pgbouncer versions are released:

1. Update the `db` image and `pgbouncer` images in [docker compose](./docker-compose-pgbouncer.yml).
2. Use the command `docker compose -f ./docker-compose-pgbouncer.yml up` to get the environment running
3. Run the integration with `go run ./src/main.go -pgbouncer -username {USERNAME} -password {PASSWORD} -p 5432 -pretty > pgbouncer_output.json`
    * If the terminal logs errors:
        - Check which query is failing
        - Explore pgbouncer release notes for changes to the `STATS` and `POOLS` tables
        - Add or remove metrics to make the query succeed
        - Modify tests to check for the new metrics in the latest versions
    * No errors:
        - Take a look at `pgbouncer_output.json` and look at the pgbouncer entities and check if metrics are reported correctly.
        - If metrics are incorrectly reported, go back and look at where queries might be failing.