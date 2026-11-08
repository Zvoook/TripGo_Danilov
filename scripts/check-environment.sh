#!/usr/bin/env bash
# Run from Ubuntu / VS Code Remote WSL. Does not modify application tables.
set -euo pipefail
cd "$(dirname "$0")/.."
for command in go gcc make docker tripgoctl psql gopls dlv; do
  command -v "$command" >/dev/null
done
tripgoctl doctor
go version
go mod verify
go build ./...
go test -race ./...
go vet ./...
test -f .env || { echo 'Run tripgoctl environment start first.' >&2; exit 1; }
set -a
source .env
set +a
: "${DATABASE_URL:?Missing DATABASE_URL}"
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -c 'select 1 as connection_ok;'

# An isolated schema tests goose up/down without touching lab migrations.
schema="tripgo_envcheck_$(date +%s)_$$"
work=$(mktemp -d)
cleanup() {
  psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -c "DROP SCHEMA IF EXISTS $schema CASCADE;" >/dev/null
  rm -f "$work/00001_environment_check.sql"
  rmdir "$work"
}
trap cleanup EXIT
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -c "CREATE SCHEMA $schema;" >/dev/null
cat > "$work/00001_environment_check.sql" <<SQL
-- +goose Up
CREATE TABLE $schema.probe (id integer PRIMARY KEY);
INSERT INTO $schema.probe VALUES (1);
-- +goose Down
DROP TABLE $schema.probe;
SQL
go tool goose -dir "$work" -table "$schema.goose_db_version" postgres "$DATABASE_URL" up
test "$(psql "$DATABASE_URL" -X -At -v ON_ERROR_STOP=1 -c "SELECT count(*) FROM $schema.probe;")" = 1
go tool goose -dir "$work" -table "$schema.goose_db_version" postgres "$DATABASE_URL" down
test "$(psql "$DATABASE_URL" -X -At -v ON_ERROR_STOP=1 -c "SELECT to_regclass('$schema.probe') IS NULL;")" = t
echo 'Environment checks passed (Go build/race toolchain/vet, PostgreSQL, goose up/down).'
