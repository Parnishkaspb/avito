mockdata:
    DB_DSN="postgres://user:password@localhost:5432/avito?sslmode=disable" \
   	go run ./cmd/mockdata

start_project:
	go run cmd/app/main.go --config=./config/config.yaml