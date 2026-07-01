.PHONY: help run build test clean docker-up docker-down docker-down-v docker-logs \
        migrate-up migrate-down migrate-force migrate-status migrate-create \
        reset start lint fmt tidy deps ci dev

# ============================================================================
# Переменные
# ============================================================================



# Загружаем переменные из .env
include .env
export $(shell sed 's/=.*//' .env)

# Бинарный файл
BINARY_NAME=polling-service
BIN_DIR=bin

# Миграции
MIGRATE_PATH=file://migrations
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Docker команда (новый синтаксис)
DOCKER_COMPOSE = docker compose

# Go команды
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get

# ============================================================================
# Определение Docker команды (поддержка старого и нового синтаксиса)
# ============================================================================

DOCKER_COMPOSE := $(shell \
	if docker compose version > /dev/null 2>&1; then \
		echo "docker compose"; \
	else \
		echo "docker-compose"; \
	fi \
)
# ============================================================================
# Основные команды
# ============================================================================

help:
	@echo "📋 Доступные команды:"
	@echo ""
	@echo "🚀 Запуск и сборка:"
	@echo "  make run           - Запустить приложение"
	@echo "  make build         - Собрать бинарник в $(BIN_DIR)/"
	@echo "  make dev           - Запустить с горячей перезагрузкой (требуется air)"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  make docker-up     - Поднять все сервисы (PostgreSQL + Redis)"
	@echo "  make docker-down   - Остановить все сервисы"
	@echo "  make docker-down-v - Остановить и удалить тома с данными"
	@echo "  make docker-logs   - Показать логи контейнеров"
	@echo "  make docker-ps     - Показать статус контейнеров"
	@echo "  make docker-restart - Перезапустить контейнеры"
	@echo "  make docker-build  - Пересобрать образы"
	@echo ""
	@echo "🗄️ Миграции:"
	@echo "  make migrate-up    - Применить все миграции"
	@echo "  make migrate-down  - Откатить последнюю миграцию"
	@echo "  make migrate-down-all - Откатить все миграции"
	@echo "  make migrate-force version=N - Форсировать версию N"
	@echo "  make migrate-status - Показать статус миграций"
	@echo "  make migrate-create name=NAME - Создать новую миграцию"
	@echo ""
	@echo "🧹 Утилиты:"
	@echo "  make clean         - Очистить собранные файлы"
	@echo "  make tidy          - Обновить go.mod"
	@echo "  make fmt           - Отформатировать код"
	@echo "  make lint          - Запустить линтер"
	@echo "  make test          - Запустить тесты"
	@echo "  make deps          - Загрузить зависимости"
	@echo ""
	@echo "🔄 Полный цикл:"
	@echo "  make start         - Запустить всё (Docker + миграции + приложение)"
	@echo "  make reset         - Полностью пересоздать БД"
	@echo "  make ci            - Проверки для CI/CD"

# ============================================================================
# Запуск и сборка
# ============================================================================

run:
	$(GORUN) cmd/api/main.go

build:
	mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) cmd/api/main.go
	@echo "✅ Бинарник собран: $(BIN_DIR)/$(BINARY_NAME)"

dev:
	@echo "🔄 Запуск с горячей перезагрузкой..."
	@which air > /dev/null || (echo "❌ Установите air: go install github.com/air-verse/air@latest" && exit 1)
	air

# ============================================================================
# Docker
# ============================================================================

docker-up:
	$(DOCKER_COMPOSE) up -d
	@echo "✅ Контейнеры запущены"

docker-down:
	$(DOCKER_COMPOSE) down
	@echo "✅ Контейнеры остановлены"

docker-down-v:
	$(DOCKER_COMPOSE) down -v
	@echo "✅ Контейнеры остановлены и тома удалены"

docker-logs:
	$(DOCKER_COMPOSE) logs -f

docker-ps:
	$(DOCKER_COMPOSE) ps

docker-restart: docker-down docker-up

docker-build:
	$(DOCKER_COMPOSE) build --no-cache
	@echo "✅ Образы пересобраны"

docker-shell:
	@echo "🐚 Подключаемся к PostgreSQL..."
	docker exec -it polling_db psql -U $(DB_USER) -d $(DB_NAME)

redis-shell:
	@echo "🐚 Подключаемся к Redis..."
	docker exec -it polling_redis redis-cli

# ============================================================================
# Миграции
# ============================================================================

migrate-up:
	@echo "🔄 Применяем миграции..."
	$(GORUN) cmd/migrate/main.go up
	@echo "✅ Миграции применены"

migrate-down:
	@echo "🔄 Откатываем миграцию..."
	$(GORUN) cmd/migrate/main.go down
	@echo "✅ Миграция откачена"

migrate-down-all:
	@echo "🔄 Откатываем все миграции..."
	$(GORUN) cmd/migrate/main.go down -all
	@echo "✅ Все миграции откачены"

migrate-force:
	@if [ -z "$(version)" ]; then \
		echo "❌ Укажите версию: make migrate-force version=N"; \
		exit 1; \
	fi
	@echo "🔄 Форсируем версию $(version)..."
	$(GORUN) cmd/migrate/main.go force $(version)
	@echo "✅ Версия $(version) форсирована"

migrate-status:
	@echo "📊 Статус миграций:"
	docker exec -it polling_db psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT * FROM schema_migrations;"

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "❌ Укажите имя миграции: make migrate-create name=NAME"; \
		exit 1; \
	fi
	@echo "🔄 Создаем миграцию: $(name)"
	migrate create -ext sql -dir migrations -seq $(name)
	@echo "✅ Миграция создана"

migrate-fix:
	@echo "🔧 Исправляем dirty состояние..."
	docker exec -it polling_db psql -U $(DB_USER) -d $(DB_NAME) -c "UPDATE schema_migrations SET dirty = false;"
	@echo "✅ Dirty флаг сброшен"

# ============================================================================
# Утилиты
# ============================================================================

clean:
	rm -rf $(BIN_DIR)/
	rm -rf tmp/
	go clean
	@echo "✅ Проект очищен"

tidy:
	$(GOMOD) tidy
	@echo "✅ go.mod обновлен"

fmt:
	go fmt ./...
	@echo "✅ Код отформатирован"

lint:
	@which golangci-lint > /dev/null || (echo "❌ Установите golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...
	@echo "✅ Линтер прошел успешно"

test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "✅ Тесты пройдены"

deps:
	$(GOMOD) download
	$(GOMOD) verify
	@echo "✅ Зависимости загружены"

coverage:
	go tool cover -html=coverage.out
	@echo "📊 Отчет покрытия открыт в браузере"

# ============================================================================
# Полный цикл
# ============================================================================

start: docker-up migrate-up run

reset: docker-down-v docker-up migrate-up
	@echo "✅ Полный сброс выполнен"

ci: tidy fmt lint test build
	@echo "✅ CI проверки пройдены"

# ============================================================================
# Быстрые команды (алиасы)
# ============================================================================

up: docker-up
down: docker-down
logs: docker-logs
ps: docker-ps
status: migrate-status
migrate: migrate-up