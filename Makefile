run:
	go run cmd/main.go

migrate:
	migrate -path=migrations -database=postgres://user:pass@localhost:5432/dbname up
