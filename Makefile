# ================================
# STRICT SHELL (PRODUCTION SAFETY)
# ================================
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c


# ================================
# USAGE GUIDE
# ================================
# Create a new migration:
#   make migrate-create NAME=create_tasks_table
#
# Run all migrations:
#   make migrate-up
#
# Run only one migration:
#   make migrate-up-one
#
# Rollback N migrations:
#   make migrate-down STEPS=1
#
# Rollback all migrations:
#   make migrate-down-all
#
# Fix dirty migration state:
#   make migrate-fix VERSION=1
#
# Check current migration version:
#   make migrate-version
#
# Drop all tables (DANGEROUS):
#   make migrate-drop
#
# Show help:
#   make help


# ================================
# CONFIG
# ================================
# DB_URL is expected from environment (.envrc or export)
# Example:
#   export DB_URL=postgres://user:pass@localhost:5432/dbname?sslmode=disable
#
# Or override:
#   make migrate-up DB_URL=postgres://...

MIGRATIONS_PATH ?= ./cmd/migrate/migrations


# ================================
# PHONY TARGETS
# ================================
.PHONY: migrate-create migrate-up migrate-up-one migrate-up-all \
        migrate-down migrate-down-all \
        migrate-fix migrate-version migrate-drop \
        help check-db print-env


# ================================
# INTERNAL CHECKS
# ================================
check-db:
ifndef DB_URL
	$(error DB_URL is not set. Use .envrc or export DB_URL=...)
endif


# ================================
# CREATE MIGRATION
# ================================
migrate-create:
ifndef NAME
	$(error NAME is required. Usage: make migrate-create NAME=create_tasks_table)
endif
	@echo "Creating migration: $(NAME)"
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(NAME)


# ================================
# APPLY MIGRATIONS
# ================================
migrate-up: check-db
	@echo "Running migrations UP on $(DB_URL)"
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" up

# alias (optional but useful)
migrate-up-all: migrate-up

migrate-up-one: check-db
	@echo "Running 1 migration UP on $(DB_URL)"
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" up 1


# ================================
# ROLLBACK MIGRATIONS
# ================================
migrate-down: check-db
ifndef STEPS
	$(error STEPS is required. Usage: make migrate-down STEPS=1)
endif
	@echo "Rolling back $(STEPS) migrations on $(DB_URL)"
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" down $(STEPS)

migrate-down-all: check-db
	@echo "Rolling back ALL migrations on $(DB_URL)"
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" down -all


# ================================
# FIX / DEBUG
# ================================
migrate-fix: check-db
ifndef VERSION
	$(error VERSION is required. Usage: make migrate-fix VERSION=1)
endif
	@echo "Forcing migration version to $(VERSION)"
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" force $(VERSION)

migrate-version: check-db
	@echo "Current migration version:"
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" version


# ================================
# DANGEROUS (USE CAREFULLY)
# ================================
migrate-drop: check-db
	@echo "⚠️  Dropping entire database schema on $(DB_URL)"
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" drop -f


# ================================
# DEBUG HELPERS
# ================================
print-env:
	@echo "DB_URL=$(DB_URL)"
	@echo "MIGRATIONS_PATH=$(MIGRATIONS_PATH)"


# ================================
# HELP
# ================================
help:
	@echo "=== Migration Commands ==="
	@echo "  make migrate-create NAME=..."
	@echo "  make migrate-up"
	@echo "  make migrate-up-one"
	@echo "  make migrate-up-all"
	@echo "  make migrate-down STEPS=1"
	@echo "  make migrate-down-all"
	@echo "  make migrate-fix VERSION=..."
	@echo "  make migrate-version"
	@echo "  make migrate-drop"
	@echo ""
	@echo "=== Debug ==="
	@echo "  make print-env"