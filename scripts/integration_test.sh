#!/usr/bin/env bash
set -euo pipefail

# Start Postgres container using Docker Compose
docker-compose up -d db

# Wait for Postgres to be ready (using the health check)
echo "Waiting for Postgres to be healthy..."
for i in {1..10}; do
  if docker-compose exec db pg_isready -U postgres -d postgres; then
    break
  fi
  sleep 2
done

# 1. Test `up` command applies all migrations
echo "Running gomigrator up..."
./bin/gomigrator --log-level debug --dir test_migrations "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable" up

# Verify that both tables were created
echo "Verifying tables after up..."
docker-compose exec -T db psql -U postgres -d postgres -c "\dt" | grep -q "test_table1" 
docker-compose exec -T db psql -U postgres -d postgres -c "\dt" | grep -q "test_table2" 
echo "Both test_table1 and test_table2 exist."

# Verify that the database version is 2 (both migrations applied)
DB_VERSION=$(./bin/gomigrator --dir test_migrations "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable" dbversion)
echo "Current DB version after up: $DB_VERSION"
if [ "$DB_VERSION" -ne 2 ]; then
  echo "Error: Expected DB version 2, got $DB_VERSION"
  docker-compose down -v
  exit 1
fi

# 2. Test `down` rolls back the last migration
echo "Running gomigrator down (rollback last migration)..."
./bin/gomigrator --log-level debug --dir test_migrations "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable" down

# Verify that the second table was dropped but the first table still exists
echo "Verifying tables after one down..."
! docker-compose exec -T db psql -U postgres -d postgres -c "\dt" | grep -q "test_table2"  || { 
  echo "Error: test_table2 should have been dropped, but still exists"; exit 1; }
docker-compose exec -T db psql -U postgres -d postgres -c "\dt" | grep -q "test_table1"
echo "test_table2 is dropped, test_table1 still exists."

# Verify DB version is now 1 (only first migration applied)
DB_VERSION=$(./bin/gomigrator --dir test_migrations "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable" dbversion)
echo "Current DB version after one down: $DB_VERSION"
if [ "$DB_VERSION" -ne 1 ]; then
  echo "Error: Expected DB version 1 after rollback, got $DB_VERSION"
  docker-compose down -v
  exit 1
fi

# 3. Test concurrent `gomigrator up` processes (locking behavior)
echo "Testing concurrent gomigrator up processes to check locking..."
# Start one migrator process in the background (on a fresh DB state)
docker-compose down -v   # reset DB (drop volumes)
docker-compose up -d db  # fresh DB container
# Wait for DB to be ready again
for i in {1..10}; do
  if docker-compose exec db pg_isready -U postgres -d postgres; then
    break
  fi
  sleep 2
done
echo "Starting first migrator (with up)..."
./bin/gomigrator --dir test_migrations "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable" up &
pid1=$!
sleep 1  # give the first migrator a head start (it will pause in migration 002)
echo "Starting second migrator (with up)..."
./bin/gomigrator --dir test_migrations "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable" up &
pid2=$!
wait $pid1
status1=$?
wait $pid2
status2=$?
if [[ $status1 -ne 0 || $status2 -ne 0 ]]; then
  echo "Error: One of the concurrent migrator processes failed (locking issue)"
  docker-compose down -v
  exit 1
fi
echo "Both migrator processes finished successfully."

# Verify that all migrations were applied exactly once
DB_VERSION=$(./bin/gomigrator --dir test_migrations "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable" dbversion)
echo "Final DB version after concurrent up: $DB_VERSION"
if [ "$DB_VERSION" -ne 2 ]; then
  echo "Error: Expected DB version 2 after concurrent up, got $DB_VERSION"
  docker-compose down -v
  exit 1
fi

# Cleanup: stop and remove containers
docker-compose down -v
echo "Integration tests completed successfully."
