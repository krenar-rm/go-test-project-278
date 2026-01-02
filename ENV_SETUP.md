# 🔧 Настройка переменных окружения на Render

## Обязательные переменные

### 1. `PORT`
```
PORT=80
```
⚠️ **ВАЖНО:** На Render используйте `80` - это порт для **Caddy** (не для Go приложения)!  
Go приложение внутри контейнера всегда использует порт `8080`.  
Caddy слушает на `PORT` (80) и проксирует на `localhost:8080`.

### 2. `DATABASE_URL`
```
DATABASE_URL=<your-internal-database-url>
```
📌 **Как получить:**
1. Создайте PostgreSQL базу на Render: Dashboard → "New +" → "PostgreSQL"
2. Скопируйте **Internal Database URL** (НЕ External!)
3. Формат: `postgres://user:password@host/database`

### 3. `BASE_URL`
```
BASE_URL=https://your-app-name.onrender.com
```
📌 Замените `your-app-name` на реальное имя вашего сервиса на Render.

### 4. `ENV`
```
ENV=production
```
📌 Устанавливает режим production (отключает CORS для localhost, убирает debug endpoints).

---

## Опциональные переменные

### 5. `SENTRY_DSN` (рекомендуется)
```
SENTRY_DSN=https://xxxxx@o123456.ingest.sentry.io/789012
```
📌 **Как получить:**
1. Создайте проект на [Sentry.io](https://sentry.io/)
2. Скопируйте DSN из настроек проекта
3. Если не хотите использовать Sentry - оставьте пустым

### 6. `FRONTEND_URL` (опционально)
```
FRONTEND_URL=https://your-app-name.onrender.com
```
📌 Для настройки CORS в production. Обычно совпадает с `BASE_URL`.  
Если не указано - используется значение `BASE_URL`.

---

## Как добавить переменные на Render

### Через Web UI:
1. Откройте ваш сервис в [Render Dashboard](https://dashboard.render.com/)
2. Перейдите в **Environment** в левом меню
3. Нажмите **Add Environment Variable**
4. Добавьте каждую переменную:
   - **Key**: название переменной (например, `PORT`)
   - **Value**: значение (например, `80`)
5. Нажмите **Save Changes**
6. Render автоматически пересоберет приложение

### Через render.yaml (Infrastructure as Code):
Создайте файл `render.yaml` в корне проекта:

```yaml
services:
  - type: web
    name: url-shortener
    runtime: docker
    plan: free
    envVars:
      - key: PORT
        value: 80
      - key: ENV
        value: production
      - key: BASE_URL
        value: https://your-app-name.onrender.com
      - key: DATABASE_URL
        fromDatabase:
          name: urlshortener-db
          property: connectionString
      - key: SENTRY_DSN
        sync: false  # Добавьте вручную через UI

databases:
  - name: urlshortener-db
    databaseName: urlshortener
    plan: free
```

---

## Проверка настроек

После деплоя проверьте:

1. **Логи запуска** (в Render Dashboard → Logs):
   ```
   [run.sh] Starting service
   [run.sh] Running DB migrations
   [run.sh] Starting Caddy
   [run.sh] Starting Go app
   Database connected successfully
   Starting server on port 8080
   ```

2. **Проверка API:**
   ```bash
   curl https://your-app-name.onrender.com/ping
   # Ответ: pong
   ```

3. **Проверка UI:**
   Откройте `https://your-app-name.onrender.com` в браузере

---

## Локальная разработка

Для локальной разработки создайте файл `.env`:

```bash
cp env.example .env
```

Отредактируйте `.env`:

```bash
PORT=8080
BASE_URL=http://localhost:8080
DATABASE_URL=postgres://postgres:password@localhost:5432/urlshortener?sslmode=disable
ENV=development
SENTRY_DSN=  # Оставьте пустым или укажите свой DSN
```

Запустите приложение:
```bash
make dev  # Запускает frontend + backend
```

---

## Troubleshooting

### ❌ Ошибка: "Failed to connect to database"
**Решение:** Проверьте правильность `DATABASE_URL`. Убедитесь, что используете **Internal URL**, а не External.

### ❌ Ошибка: "bind: permission denied"
**Решение:** На Render используйте `PORT=80`, а не `8080`.

### ❌ Ошибка: "CORS policy: No 'Access-Control-Allow-Origin'"
**Решение:**
- Development: убедитесь, что `ENV=development`
- Production: проверьте `FRONTEND_URL` или `BASE_URL`

### ❌ Frontend не отображается
**Решение:** 
1. Проверьте, что `package-lock.json` добавлен в git
2. Проверьте логи сборки в Render
3. Убедитесь, что `PORT=80` (для Caddy)

---

## Дополнительная информация

- 📖 [Документация Render](https://render.com/docs)
- 📖 [Документация Sentry](https://docs.sentry.io/)
- 📖 [PostgreSQL на Render](https://render.com/docs/databases)

