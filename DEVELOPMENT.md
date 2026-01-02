# Руководство по разработке

## Быстрый старт

### 1. Установка зависимостей

```bash
# Клонируйте репозиторий
git clone https://github.com/krenar-rm/go-test-project-278.git
cd go-test-project-278

# Установите Go зависимости
go mod download

# Установите инструменты (если еще не установлены)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### 2. Настройка базы данных

```bash
# Запустите PostgreSQL через Docker
make db-up

# Дождитесь запуска (3-5 секунд)
sleep 5

# Примените миграции
export DATABASE_URL="postgres://postgres:password@localhost:5432/urlshortener?sslmode=disable"
make db-migrate
```

### 3. Настройка переменных окружения

```bash
# Скопируйте пример конфигурации
cp env.example .env

# Отредактируйте .env файл
nano .env  # или используйте ваш любимый редактор
```

Минимальная конфигурация:
```
PORT=8080
BASE_URL=http://localhost:8080
DATABASE_URL=postgres://postgres:password@localhost:5432/urlshortener?sslmode=disable
ENV=development
```

### 4. Запуск приложения

```bash
# Запустите приложение
make run

# Или напрямую
go run main.go
```

Приложение будет доступно на `http://localhost:8080`

## Работа с базой данных

### Создание новой миграции

```bash
# Создайте файл миграции вручную в db/migrations/
# Формат: XXXXX_description.sql

# Пример:
# db/migrations/00002_add_clicks_counter.sql
```

Формат миграции:
```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE links ADD COLUMN clicks INTEGER DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links DROP COLUMN clicks;
-- +goose StatementEnd
```

### Применение миграций

```bash
# Применить все миграции
make db-migrate

# Откатить последнюю миграцию
make db-rollback

# Проверить статус миграций
goose -dir ./db/migrations postgres "$DATABASE_URL" status
```

### Работа с sqlc

Если вы изменили SQL queries или схему:

```bash
# 1. Обновите файлы в sql/queries/ или sql/schema/
# 2. Регенерируйте Go код
make sqlc-generate

# 3. Проверьте, что код компилируется
go build ./...
```

## Тестирование

### Запуск тестов

```bash
# Базовые тесты (без БД)
go test ./...

# Интеграционные тесты (требуется БД)
export TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/urlshortener?sslmode=disable"
go test -v ./internal/handlers/

# Тесты с race detector
go test -race ./...

# Тесты с покрытием
make test-coverage
```

### Тестирование API вручную

```bash
# Создать ссылку
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com","short_name":"test"}'

# Получить все ссылки
curl http://localhost:8080/api/links

# Получить ссылку по ID
curl http://localhost:8080/api/links/1

# Обновить ссылку
curl -X PUT http://localhost:8080/api/links/1 \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com/updated","short_name":"test2"}'

# Удалить ссылку
curl -X DELETE http://localhost:8080/api/links/1

# Проверить редирект
curl -L http://localhost:8080/r/test
```

## Линтинг и форматирование

```bash
# Запуск линтера
make lint

# Или напрямую
golangci-lint run

# Форматирование кода
go fmt ./...
goimports -w .
```

## Структура проекта

- `internal/database/` - сгенерированный sqlc код для работы с БД
- `internal/handlers/` - HTTP handlers для API endpoints
- `internal/models/` - бизнес-модели и структуры запросов/ответов
- `sql/queries/` - SQL queries для sqlc (CRUD операции)
- `sql/schema/` - SQL схема базы данных
- `db/migrations/` - миграции базы данных (goose)

## Полезные команды

```bash
# База данных
make db-up           # Запустить PostgreSQL
make db-down         # Остановить PostgreSQL  
make db-migrate      # Применить миграции
make db-rollback     # Откатить миграцию
make sqlc-generate   # Генерация Go кода из SQL

# Разработка
make run             # Запустить приложение
make build           # Собрать приложение
make clean           # Очистить временные файлы

# Тестирование
make test            # Запустить тесты
make test-race       # Тесты с race detector
make test-coverage   # Тесты с покрытием
make lint            # Запустить линтер

# Помощь
make help            # Показать все команды
```

## Отладка

### Подключение к базе данных

```bash
# Через psql
psql "$DATABASE_URL"

# Через Docker
docker exec -it urlshortener-db psql -U postgres -d urlshortener
```

### Просмотр логов PostgreSQL

```bash
docker logs urlshortener-db
```

### Проверка миграций

```bash
# Проверить текущую версию
goose -dir ./db/migrations postgres "$DATABASE_URL" version

# Показать статус всех миграций
goose -dir ./db/migrations postgres "$DATABASE_URL" status
```

## Частые проблемы

### База данных не запускается

```bash
# Проверьте, не занят ли порт 5432
lsof -i :5432

# Остановите и пересоздайте контейнер
make db-down
docker volume rm krenar_url_shorter_postgres_data
make db-up
```

### Ошибки компиляции после изменения SQL

```bash
# Регенерируйте код sqlc
make sqlc-generate

# Обновите зависимости
go mod tidy
```

### Тесты не проходят

```bash
# Убедитесь, что БД запущена
make db-up

# Примените миграции
make db-migrate

# Очистите данные перед тестами
export TEST_DATABASE_URL="$DATABASE_URL"
psql "$DATABASE_URL" -c "TRUNCATE TABLE links RESTART IDENTITY CASCADE;"
```

## CI/CD

GitHub Actions автоматически запускает:
- Линтер (golangci-lint)
- Тесты (go test)
- Сборку (go build)

При пуше в `main` или `master` ветку.

## Полезные ссылки

- [sqlc документация](https://docs.sqlc.dev/)
- [goose документация](https://github.com/pressly/goose)
- [Gin документация](https://gin-gonic.com/docs/)
- [PostgreSQL документация](https://www.postgresql.org/docs/)

