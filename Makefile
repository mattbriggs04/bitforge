ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: dev api backend migrate seed reset-db frontend backend-test frontend-lint

DATABASE_URL ?= file:bitforge.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)
DEFAULT_USER_HANDLE ?= demo
COMPILE_TIMEOUT ?= 20s
RUN_TIMEOUT ?= 3s

dev:
	@trap 'kill 0' INT TERM EXIT; \
	$(MAKE) api & \
	$(MAKE) frontend & \
	$(MAKE) worker & \
	wait

api:
	cd backend && \
	DATABASE_URL="$(DATABASE_URL)" \
	DEFAULT_USER_HANDLE="$(DEFAULT_USER_HANDLE)" \
	COMPILE_TIMEOUT="$(COMPILE_TIMEOUT)" \
	RUN_TIMEOUT="$(RUN_TIMEOUT)" \
	go run ./cmd/api

backend: api

migrate:
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate

seed:
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed

reset-db:
	rm -f backend/bitforge.db backend/bitforge.db-shm backend/bitforge.db-wal
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate
	cd backend && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed

frontend:
	cd frontend && npm run dev

backend-test:
	cd backend && go test ./...

frontend-lint:
	cd frontend && npm run lint

.PHONY: worker sandbox-image generation-prompt
worker:
	cd backend && DATABASE_URL="$(DATABASE_URL)" COMPILE_TIMEOUT="$(COMPILE_TIMEOUT)" RUN_TIMEOUT="$(RUN_TIMEOUT)" go run ./cmd/judge-worker

sandbox-image:
	docker build -t bitforge-runner:v1 sandbox

generation-prompt:
	cd backend && go run ./cmd/problems prompt -config ../generation/request.json -out ../generation/next-question.txt
