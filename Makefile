.PHONY: run test test-race lint build clean help db-up db-down db-migrate db-rollback sqlc-generate dev frontend backend install

# Установка зависимостей
install:
	go mod download
	npm install

# Запуск приложения (только backend)
run:
	go run main.go

# Запуск backend
backend:
	go run main.go

# Запуск frontend
frontend:
	npm run frontend

# Запуск frontend и backend одновременно
dev:
	npm run dev

# Запуск тестов
test:
	go test -v ./...

# Запуск тестов с проверкой race conditions
test-race:
	go test -v -race ./...

# Запуск тестов с покрытием
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Запуск линтера
lint:
	golangci-lint run

# Сборка приложения
build:
	go build -o url-shortener main.go

# Очистка
clean:
	rm -f url-shortener coverage.out coverage.html

# Запуск PostgreSQL через Docker
db-up:
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3

# Остановка PostgreSQL
db-down:
	docker-compose -f docker-compose.dev.yml down

# Применить миграции
db-migrate:
	goose -dir ./db/migrations postgres "$(DATABASE_URL)" up

# Откатить последнюю миграцию
db-rollback:
	goose -dir ./db/migrations postgres "$(DATABASE_URL)" down

# Генерация кода из SQL с помощью sqlc
sqlc-generate:
	sqlc generate

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make install        - Установить все зависимости (Go + Node.js)"
	@echo "  make dev            - Запустить frontend и backend одновременно"
	@echo "  make run            - Запустить только backend"
	@echo "  make backend        - Запустить только backend"
	@echo "  make frontend       - Запустить только frontend"
	@echo "  make test           - Запустить тесты"
	@echo "  make test-race      - Запустить тесты с проверкой race conditions"
	@echo "  make test-coverage  - Запустить тесты с покрытием"
	@echo "  make lint           - Запустить линтер"
	@echo "  make build          - Собрать приложение"
	@echo "  make clean          - Очистить сгенерированные файлы"
	@echo ""
	@echo "Database команды:"
	@echo "  make db-up          - Запустить PostgreSQL в Docker"
	@echo "  make db-down        - Остановить PostgreSQL"
	@echo "  make db-migrate     - Применить миграции"
	@echo "  make db-rollback    - Откатить последнюю миграцию"
	@echo "  make sqlc-generate  - Генерация кода из SQL"
