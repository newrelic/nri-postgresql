-- Connect to titanic database
\c titanic;

-- Import the titanic.sql file that was downloaded during Docker build
\i /docker-entrypoint-initdb.d/titanic.sql;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO dbuser;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO dbuser;

-- Analyze tables for better query planning
ANALYZE VERBOSE;
