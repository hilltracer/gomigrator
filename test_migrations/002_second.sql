-- +gomigrator Up

CREATE TABLE test_table2 (
    id SERIAL PRIMARY KEY,
    data TEXT
);
-- Add an artificial delay to simulate a long-running migration (for lock testing)
SELECT pg_sleep(5);

-- +gomigrator Down

DROP TABLE IF EXISTS test_table2;
