DB_DSN=postgres://user:password@localhost:5432/avito?sslmode=disable

# Генерация мок-данных в базе
mockdata:
	@echo "Генерация тестовых данных..."
	DB_DSN=$(DB_DSN) go run ./cmd/mockdata

# Сборка сервиса
build:
	@echo "Сборка сервиса..."
	go build -o bin/app ./cmd/app

# Запуск сервиса напрямую без сборки
start:
	@echo "Запуск сервиса..."
	go run ./cmd/app/main.go --config=./config/config.yaml

# Очистка бинарников
clean:
	@echo "Очистка..."
	rm -rf bin
