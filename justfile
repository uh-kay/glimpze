migrate-create name:
	goose create -s {{ name }} sql

migrate-up:
	goose up

migrate-down:
	goose reset

seed:
	@go run cmd/seed/main.go

gen-docs:
	@swag init -g ./api/main.go -d cmd,store && swag fmt

test:
	@go test ./...

golangci-lint:
    golangci-lint run

dev:
    docker compose up -d db && air

db:
    docker exec -it glimpze-db psql -U root -d glimpze
