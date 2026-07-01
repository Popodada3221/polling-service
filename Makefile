.PHONY: help run migrate-up migrate-down docker-up docker-down test build clean

# Переменные
BINARY_NAME=polling-service
MIGRATE_PATH=file://migrations
DB_URL=postgres://poll_user:poll_pass@localhost:5432/polling_db?sslmode=disable

help:
	@echo "Available commands:"
	@echo "  make run           - Run the application"
	@echo "  make build         - Build the application"
	@echo "  make test          - Run tests"
	@echo "  make migrate-up    - Run database migrations up"
	@echo "  make migrate-down  - Rollback database migrations"
	@echo "  make docker-up     - Start PostgreSQL in Docker"
	@echo "  make docker-down   - Stop PostgreSQL in Docker"
	@echo "  make clean         - Clean build artifacts"

# Запуск приложения
run:
	go run cmd/api/main.go

# Сборка приложения
build:
	go build -o bin/$(BINARY_NAME) cmd/api/main.go

# Запуск тестов
test:
	go test -v ./...

# Очистка
clean:
	rm -rf bin/
	go clean

# Docker
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-down-v:
	docker compose down -v

docker-logs:
	docker compose logs -f

# Миграции
migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

migrate-force:
	@echo "Usage: make migrate-force version=N"
	go run cmd/migrate/main.go force $(version)

migrate-status:
	@echo "Current migration status:"
	docker exec -it polling_db psql -U poll_user -d polling_db -c "SELECT * FROM schema_migrations;"

# Полный перезапуск (очистка + старт)
reset: docker-down-v docker-up migrate-up
	@echo "Database reset completed"

# Запуск с миграциями
start: docker-up migrate-up run

# Проверка кода
lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

# Проверка зависимостей
deps:
	go mod download
	go mod verify

# Все команды для CI/CD
ci: tidy fmt lint test build

# Показать все доступные команды
all: help