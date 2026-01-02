# URL Shortener

[![CI](https://github.com/krenar-rm/go-test-project-278/actions/workflows/ci.yml/badge.svg)](https://github.com/krenar-rm/go-test-project-278/actions/workflows/ci.yml)
[![Hexlet tests](https://github.com/krenar-rm/go-test-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/krenar-rm/go-test-project-278/actions)

Сервис для сокращения URL-адресов на Go + PostgreSQL + React Admin.

## 🌐 Демо

**Работающий сервис:** https://go-test-project-278.onrender.com/

## 🚀 Возможности

- REST API для управления ссылками
- Web UI (React Admin)
- Пагинация и валидация
- Редирект по коротким ссылкам
- PostgreSQL + sqlc + goose
- Docker + Caddy
- CI/CD + Sentry

## 🛠 Технологии

- **Backend:** Go 1.24, Gin, PostgreSQL, sqlc
- **Frontend:** React Admin
- **Deploy:** Docker, Caddy, Render

## 📦 Быстрый старт

```bash
# Установка
git clone https://github.com/krenar-rm/go-test-project-278.git
cd go-test-project-278
npm install

# Запуск PostgreSQL
make db-up

# Настройка переменных окружения
cp env.example .env
# Отредактируйте .env

# Запуск приложения
make dev
```

Приложение доступно:
- **Backend API:** http://localhost:8080
- **Frontend UI:** http://localhost:5173

## 📝 API Endpoints

```bash
GET  /ping                # Health check
GET  /api/links           # Список ссылок (с пагинацией)
POST /api/links           # Создать ссылку
GET  /api/links/:id       # Получить ссылку
PUT  /api/links/:id       # Обновить ссылку
DELETE /api/links/:id     # Удалить ссылку
GET  /r/:shortName        # Редирект
GET  /api/link_visits     # История посещений
```

## 🔧 Переменные окружения

См. `env.example` и `ENV_SETUP.md` для подробностей.

**Для Render:**
```bash
PORT=80
ENV=production
BASE_URL=https://your-app.onrender.com
DATABASE_URL=<postgresql-url>
```

## 📚 Документация

- `ENV_SETUP.md` - настройка переменных окружения
- `env.example` - шаблон конфигурации

## 🧪 Тестирование

```bash
go test -v ./...
golangci-lint run
```

## 📄 Лицензия

Проект создан в рамках обучения на [Hexlet](https://hexlet.io/).
