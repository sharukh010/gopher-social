include .env 
MIGRATIONS_PATH = ./cmd/migrate/migrations 

.PHONY:migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY:migrate-up 
migrate-up:
	@migrate -path $(MIGRATIONS_PATH) -database $(DB_MIGRATION_ADDR) up

.PHONY:migrate-down 
migrate-down:
	@migrate -path $(MIGRATIONS_PATH) -database $(DB_MIGRATION_ADDR) down $(filter-out $@,$(MAKECMDGOALS))
.PHONY:up-db
up-db: 
	docker compose up --build
.PHONY:down-db 
down-db: 
	docker compose down 
.PHONY:seed 
seed:
	@go run cmd/migrate/seed/main.go'

.PHONY: gen-docs 
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt 

.PHONY: run 
run:
	@make gen-docs && air