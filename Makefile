include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: Show this help message
.PHONY: help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' Makefile | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #
.PHONY: run
run: 
	RSSAPP_DB_DSN=${RSSAPP_DB_DSN} air

# ==================================================================================== #
# TEST
# ==================================================================================== #
.PHONY: test
test:
	go test -v ./...

# ==================================================================================== #
# DATABASE
# ==================================================================================== #

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${RSSAPP_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo "Creating migration files for ${name}..."
	goose -s -dir ./migrations create ${name} sql

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo "Running up migrations..."
	goose -dir ./migrations postgres ${RSSAPP_DB_DSN} up

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo "Running down migrations..."
	goose -dir ./migrations postgres ${RSSAPP_DB_DSN} down

