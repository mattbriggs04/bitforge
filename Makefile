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
	cd backend && \
	DATABASE_URL="$(DATABASE_URL)" \
	go run ./cmd/migrate

seed:
	cd backend && \
	DATABASE_URL="$(DATABASE_URL)" \
	go run ./cmd/seed

frontend:
	cd frontend && npm run dev

backend-test:
	cd backend && go test ./...

frontend-lint:
	cd frontend && npm run lint
