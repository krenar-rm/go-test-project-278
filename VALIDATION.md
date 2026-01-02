# Валидация API

## Обзор

API реализует полную валидацию входящих данных с единым форматом ошибок.

## Коды ответов

| Код | Описание | Когда возвращается |
|-----|----------|-------------------|
| 400 | Bad Request | Невалидный JSON |
| 422 | Unprocessable Entity | Ошибки валидации полей |

## Формат ответов об ошибках

### 400 Bad Request - Невалидный JSON

Возвращается, если тело запроса содержит некорректный JSON.

**Пример запроса:**
```bash
POST /api/links
Content-Type: application/json

{invalid json}
```

**Ответ:**
```json
{
  "error": "invalid request"
}
```

### 422 Unprocessable Entity - Ошибки валидации

Возвращается, если данные не прошли валидацию.

**Формат:**
```json
{
  "errors": {
    "field_name": "error message"
  }
}
```

**Пример - невалидный URL:**
```json
{
  "errors": {
    "original_url": "must be a valid URL"
  }
}
```

**Пример - несколько ошибок:**
```json
{
  "errors": {
    "original_url": "OriginalURL is required",
    "short_name": "must be at least 3 characters"
  }
}
```

## Правила валидации

### POST /api/links

#### original_url

- **Тип:** string
- **Обязательное:** Да
- **Правило:** Должен быть валидным URL (RFC 3986)

**Примеры валидных значений:**
```
https://example.com
http://example.com/path
https://example.com:8080/path?query=value
```

**Примеры ошибок:**
```json
// Поле отсутствует
{
  "errors": {
    "original_url": "OriginalURL is required"
  }
}

// Невалидный URL
{
  "errors": {
    "original_url": "must be a valid URL"
  }
}
```

#### short_name

- **Тип:** string
- **Обязательное:** Нет (генерируется автоматически если не указан)
- **Правила:**
  - Длина: 3-32 символа
  - Только буквы и цифры (alphanumeric)
  - Уникальность

**Примеры валидных значений:**
```
abc123
hexlet
mylink
```

**Примеры ошибок:**
```json
// Слишком короткий (< 3)
{
  "errors": {
    "short_name": "must be at least 3 characters"
  }
}

// Слишком длинный (> 32)
{
  "errors": {
    "short_name": "must be at most 32 characters"
  }
}

// Недопустимые символы
{
  "errors": {
    "short_name": "must contain only alphanumeric characters"
  }
}

// Уже существует
{
  "errors": {
    "short_name": "short name already in use"
  }
}
```

### PUT /api/links/:id

#### original_url

- **Тип:** string
- **Обязательное:** Да
- **Правило:** Должен быть валидным URL (RFC 3986)

Те же правила, что и для POST.

#### short_name

- **Тип:** string
- **Обязательное:** Да (в отличие от POST)
- **Правила:**
  - Длина: 3-32 символа
  - Только буквы и цифры (alphanumeric)
  - Уникальность

Те же правила, что и для POST, но поле обязательное.

## Примеры использования

### 1. Успешное создание

**Запрос:**
```bash
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://hexlet.io",
    "short_name": "hexlet"
  }'
```

**Ответ: 201 Created**
```json
{
  "id": 1,
  "original_url": "https://hexlet.io",
  "short_name": "hexlet",
  "short_url": "http://localhost:8080/r/hexlet",
  "created_at": "2026-01-01T12:00:00Z"
}
```

### 2. Автогенерация short_name

**Запрос:**
```bash
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://example.com"
  }'
```

**Ответ: 201 Created**
```json
{
  "id": 2,
  "original_url": "https://example.com",
  "short_name": "aB3xY9Zk",
  "short_url": "http://localhost:8080/r/aB3xY9Zk",
  "created_at": "2026-01-01T12:01:00Z"
}
```

### 3. Ошибка валидации - невалидный URL

**Запрос:**
```bash
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "not-a-url"
  }'
```

**Ответ: 422 Unprocessable Entity**
```json
{
  "errors": {
    "original_url": "must be a valid URL"
  }
}
```

### 4. Ошибка валидации - слишком короткий short_name

**Запрос:**
```bash
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://example.com",
    "short_name": "ab"
  }'
```

**Ответ: 422 Unprocessable Entity**
```json
{
  "errors": {
    "short_name": "must be at least 3 characters"
  }
}
```

### 5. Ошибка валидации - недопустимые символы

**Запрос:**
```bash
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://example.com",
    "short_name": "test-123"
  }'
```

**Ответ: 422 Unprocessable Entity**
```json
{
  "errors": {
    "short_name": "must contain only alphanumeric characters"
  }
}
```

### 6. Конфликт уникальности

**Запрос:**
```bash
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://different.com",
    "short_name": "hexlet"
  }'
```

**Ответ: 422 Unprocessable Entity**
```json
{
  "errors": {
    "short_name": "short name already in use"
  }
}
```

### 7. Невалидный JSON

**Запрос:**
```bash
curl -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{invalid json}'
```

**Ответ: 400 Bad Request**
```json
{
  "error": "invalid request"
}
```

### 8. Обновление ссылки

**Запрос:**
```bash
curl -X PUT http://localhost:8080/api/links/1 \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://hexlet.io/updated",
    "short_name": "hexlet2"
  }'
```

**Ответ: 200 OK**
```json
{
  "id": 1,
  "original_url": "https://hexlet.io/updated",
  "short_name": "hexlet2",
  "short_url": "http://localhost:8080/r/hexlet2",
  "created_at": "2026-01-01T12:00:00Z"
}
```

## Сообщения об ошибках

### Все возможные сообщения

| Поле | Тег | Сообщение |
|------|-----|-----------|
| original_url | required | `OriginalURL is required` |
| original_url | url | `must be a valid URL` |
| short_name | required | `ShortName is required` |
| short_name | min | `must be at least 3 characters` |
| short_name | max | `must be at most 32 characters` |
| short_name | alphanum | `must contain only alphanumeric characters` |
| short_name | duplicate | `short name already in use` |

## Тестирование

### Запуск автоматических тестов

```bash
# Убедитесь, что TEST_DATABASE_URL установлен
export TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/urlshortener_test?sslmode=disable"

# Запустите тесты
go test -v ./internal/handlers/
```

### Ручное тестирование

Используйте скрипт `test_validation.sh`:

```bash
# Запустите приложение
make run

# В другом терминале запустите тесты
./test_validation.sh
```

## Технические детали

### Используемые библиотеки

- **gin-gonic/gin** - веб-фреймворк
- **go-playground/validator/v10** - валидация (встроена в Gin)

### Архитектура

```
HTTP Request
    ↓
JSON Parsing (400 if fails)
    ↓
Validator Tags (422 if fails)
    ↓
Business Logic
    ↓
Database Constraints (422 if conflicts)
    ↓
Success Response
```

### Кастомные валидаторы

Проект использует встроенные валидаторы Gin:
- `required` - поле обязательно
- `url` - валидный URL по RFC 3986
- `min=N` - минимальная длина
- `max=N` - максимальная длина
- `alphanum` - только буквы и цифры
- `omitempty` - пропустить если пусто

### Обработка ошибок дублирования

PostgreSQL возвращает код ошибки `23505` при нарушении UNIQUE constraint.
Приложение перехватывает эту ошибку и преобразует в понятное сообщение.

## Best Practices

1. **Всегда проверяйте Content-Type**
   ```bash
   -H "Content-Type: application/json"
   ```

2. **Обрабатывайте 422 отдельно от 400**
   - 400 - проблема с форматом запроса
   - 422 - проблема с данными

3. **Читайте поле `errors`**
   Оно содержит детальную информацию о каждой ошибке валидации.

4. **Учитывайте автогенерацию**
   `short_name` опционален при создании - будет сгенерирован автоматически.

## Миграция со старого формата

### Было (до валидации)

```json
// 400 Bad Request
{
  "error": "Key: 'CreateLinkRequest.OriginalURL' Error:Field validation for 'OriginalURL' failed on the 'url' tag"
}

// 409 Conflict
{
  "error": "Short name already exists"
}
```

### Стало (после валидации)

```json
// 422 Unprocessable Entity
{
  "errors": {
    "original_url": "must be a valid URL"
  }
}

// 422 Unprocessable Entity
{
  "errors": {
    "short_name": "short name already in use"
  }
}
```

## FAQ

### Q: Почему 422 вместо 400 для ошибок валидации?

**A:** По стандарту HTTP:
- `400 Bad Request` - запрос некорректен синтаксически (например, невалидный JSON)
- `422 Unprocessable Entity` - запрос корректен, но данные не прошли бизнес-валидацию

### Q: Можно ли использовать специальные символы в short_name?

**A:** Нет, только буквы и цифры (a-z, A-Z, 0-9).

### Q: Что если не указать short_name?

**A:** Он будет сгенерирован автоматически (8 случайных символов).

### Q: Можно ли обновить ссылку без указания short_name?

**A:** Нет, при обновлении (PUT) оба поля обязательны.

### Q: Как проверить, занят ли short_name?

**A:** Попробуйте создать ссылку. Если занят - получите ошибку 422.

---

**Версия:** 1.0  
**Дата:** 2026-01-01  
**Статус:** ✅ Реализовано и протестировано

