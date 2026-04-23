# ================================
# DATABASE CONFIG
# ================================
# You can override this:
# make migrate-up DB_URL=postgres://user:pass@localhost:5432/dbname?sslmode=disable
DB_URL ?= postgres://appuser:apppass@localhost:5432/tasks_db?sslmode=disable

# Path where migration files are stored
MIGRATIONS_PATH ?= ./cmd/migrate/migrations


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
# PHONY TARGETS
# ================================
.PHONY: migrate-create migrate-up migrate-down migrate-up-one migrate-down-all migrate-fix migrate-version migrate-drop help


# ================================
# CREATE MIGRATION
# ================================
migrate-create:
ifndef NAME
	$(error NAME is required. Usage: make migrate-create NAME=create_tasks_table)
endif
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(NAME)


# ================================
# APPLY MIGRATIONS
# ================================
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" up

migrate-up-one:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" up 1


# ================================
# ROLLBACK MIGRATIONS
# ================================
migrate-down:
ifndef STEPS
	$(error STEPS is required. Usage: make migrate-down STEPS=1)
endif
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" down $(STEPS)

migrate-down-all:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" down -all


# ================================
# FIX / DEBUG
# ================================
migrate-fix:
ifndef VERSION
	$(error VERSION is required. Usage: make migrate-fix VERSION=1)
endif
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" force $(VERSION)

migrate-version:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" version


# ================================
# DANGEROUS (USE CAREFULLY)
# ================================
migrate-drop:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" drop -f


# ================================
# HELP
# ================================
help:
	@echo "Migration Commands:"
	@echo "  make migrate-create NAME=..."
	@echo "  make migrate-up"
	@echo "  make migrate-up-one"
	@echo "  make migrate-down STEPS=1"
	@echo "  make migrate-down-all"
	@echo "  make migrate-fix VERSION=..."
	@echo "  make migrate-version"
	@echo "  make migrate-drop"
