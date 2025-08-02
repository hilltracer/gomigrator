-- +gomigrator Up

CREATE TABLE test_table1 (
    id SERIAL PRIMARY KEY,
    name TEXT
);

-- +gomigrator Down

DROP TABLE IF EXISTS test_table1;
