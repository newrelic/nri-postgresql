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
