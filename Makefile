ifneq (,$(wildcard .env))
include .env
export
endif

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/firmware_practice?sslmode=disable
REDIS_ADDR ?= localhost:6379
SUBMISSION_QUEUE ?= bitforge:submission:queue
DEFAULT_USER_HANDLE ?= demo
C_COMPILER ?= gcc
COMPILE_TIMEOUT ?= 4s
RUN_TIMEOUT ?= 2s
QUEUE_POP_TIMEOUT ?= 4s

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f

backend:
	cd backend && \
	DATABASE_URL="$(DATABASE_URL)" \
	REDIS_ADDR="$(REDIS_ADDR)" \
	SUBMISSION_QUEUE="$(SUBMISSION_QUEUE)" \
	DEFAULT_USER_HANDLE="$(DEFAULT_USER_HANDLE)" \
	C_COMPILER="$(C_COMPILER)" \
	COMPILE_TIMEOUT="$(COMPILE_TIMEOUT)" \
	RUN_TIMEOUT="$(RUN_TIMEOUT)" \
	QUEUE_POP_TIMEOUT="$(QUEUE_POP_TIMEOUT)" \
	go run ./cmd/api

worker:
	cd backend && \
	DATABASE_URL="$(DATABASE_URL)" \
	REDIS_ADDR="$(REDIS_ADDR)" \
	SUBMISSION_QUEUE="$(SUBMISSION_QUEUE)" \
	C_COMPILER="$(C_COMPILER)" \
	COMPILE_TIMEOUT="$(COMPILE_TIMEOUT)" \
	RUN_TIMEOUT="$(RUN_TIMEOUT)" \
	QUEUE_POP_TIMEOUT="$(QUEUE_POP_TIMEOUT)" \
	go run ./cmd/worker

migrate:
	@set -e; \
	if (cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate); then \
		exit 0; \
	fi; \
	echo "Local Postgres is unavailable. Starting docker postgres and retrying migrate..."; \
	docker compose up -d postgres >/dev/null; \
	for i in 1 2 3 4 5 6 7 8 9 10; do \
		if (cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate); then \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "migrate failed after retries"; \
	exit 1

seed:
	@set -e; \
	if (cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed); then \
		exit 0; \
	fi; \
	echo "Local Postgres is unavailable. Starting docker postgres and retrying seed..."; \
	docker compose up -d postgres >/dev/null; \
	for i in 1 2 3 4 5 6 7 8 9 10; do \
		if (cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed); then \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "seed failed after retries"; \
	exit 1

frontend:
	cd frontend && npm run dev

backend-test:
	cd backend && go test ./...

frontend-lint:
	cd frontend && npm run lint
