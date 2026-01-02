# Руководство по пагинации

## Общая информация

API поддерживает пагинацию для списков ссылок через параметр `range` в query string.

## Формат запроса

```
GET /api/links?range=[start,end]
```

### Параметры

- `start` - индекс первой записи (начиная с 0)
- `end` - индекс после последней записи (exclusive)
- Диапазон является **включительным слева** и **исключающим справа**

### Примеры

| Range | Описание | Количество записей |
|-------|----------|-------------------|
| `[0,10]` | Записи с индексами 0-9 | 10 записей |
| `[5,10]` | Записи с индексами 5-9 | 5 записей |
| `[10,20]` | Записи с индексами 10-19 | 10 записей |
| `[0,1]` | Только первая запись | 1 запись |

## Формат ответа

### Заголовки

Ответ включает заголовок `Content-Range`:

```
Content-Range: links start-end/total
```

Где:
- `start` - индекс первой возвращенной записи
- `end` - индекс последней возвращенной записи (**inclusive**)
- `total` - общее количество записей в базе

### Пример

**Запрос:**
```bash
curl "http://localhost:8080/api/links?range=[0,10]" -v
```

**Ответ:**
```http
HTTP/1.1 200 OK
Content-Range: links 0-9/42
Content-Type: application/json

[
  {
    "id": 1,
    "original_url": "https://example.com/1",
    "short_name": "abc123",
    "short_url": "http://localhost:8080/r/abc123"
  },
  // ... 9 остальных записей
]
```

В этом примере:
- Запрошены записи с индексами 0-9 (10 записей)
- Возвращены записи с индексами 0-9
- Всего в базе 42 записи

## Граничные случаи

### Диапазон превышает количество записей

**Запрос:**
```bash
curl "http://localhost:8080/api/links?range=[10,100]"
```

**Ответ:**
```http
Content-Range: links 10-14/15
```

Если запрошено больше записей, чем есть, возвращается столько, сколько есть.

### Запрос без пагинации

**Запрос:**
```bash
curl "http://localhost:8080/api/links"
```

**Ответ:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

[
  // Все записи без ограничения
]
```

Заголовок `Content-Range` не включается.

### Пустой результат

**Запрос:**
```bash
curl "http://localhost:8080/api/links?range=[100,110]"
```

**Ответ:**
```http
HTTP/1.1 200 OK
Content-Range: links 100-100/15

[]
```

Если записей нет в указанном диапазоне, возвращается пустой массив.

## Ошибки валидации

### Неверный формат

```bash
curl "http://localhost:8080/api/links?range=[0]"
```

**Ответ:**
```http
HTTP/1.1 400 Bad Request

{
  "error": "Invalid range parameter: invalid range format, expected [start,end]"
}
```

### Неверные значения

```bash
curl "http://localhost:8080/api/links?range=[10,5]"
```

**Ответ:**
```http
HTTP/1.1 400 Bad Request

{
  "error": "Invalid range parameter: end must be >= start"
}
```

### Отрицательные значения

```bash
curl "http://localhost:8080/api/links?range=[-1,10]"
```

**Ответ:**
```http
HTTP/1.1 400 Bad Request

{
  "error": "Invalid range parameter: start must be >= 0"
}
```

## Примеры использования

### Загрузка первой страницы (10 элементов)

```bash
curl "http://localhost:8080/api/links?range=[0,10]"
```

### Загрузка второй страницы

```bash
curl "http://localhost:8080/api/links?range=[10,20]"
```

### Загрузка страницы по номеру

```javascript
function getPage(pageNumber, pageSize = 10) {
  const start = pageNumber * pageSize;
  const end = start + pageSize;
  return fetch(`/api/links?range=[${start},${end}]`);
}

// Первая страница (pageNumber = 0)
getPage(0);  // range=[0,10]

// Вторая страница (pageNumber = 1)
getPage(1);  // range=[10,20]

// Третья страница (pageNumber = 2)
getPage(2);  // range=[20,30]
```

### Извлечение общего количества записей

```javascript
fetch('/api/links?range=[0,1]')
  .then(response => {
    const contentRange = response.headers.get('Content-Range');
    // "links 0-0/42"
    const total = parseInt(contentRange.split('/')[1]);
    console.log(`Total records: ${total}`);
  });
```

## Реализация на клиенте

### React пример

```javascript
import { useState, useEffect } from 'react';

function LinksList() {
  const [links, setLinks] = useState([]);
  const [page, setPage] = useState(0);
  const [total, setTotal] = useState(0);
  const pageSize = 10;

  useEffect(() => {
    const start = page * pageSize;
    const end = start + pageSize;
    
    fetch(`/api/links?range=[${start},${end}]`)
      .then(response => {
        const contentRange = response.headers.get('Content-Range');
        if (contentRange) {
          const totalCount = parseInt(contentRange.split('/')[1]);
          setTotal(totalCount);
        }
        return response.json();
      })
      .then(data => setLinks(data));
  }, [page]);

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div>
      <ul>
        {links.map(link => (
          <li key={link.id}>{link.short_url}</li>
        ))}
      </ul>
      
      <div>
        <button 
          onClick={() => setPage(p => Math.max(0, p - 1))}
          disabled={page === 0}
        >
          Previous
        </button>
        
        <span>Page {page + 1} of {totalPages}</span>
        
        <button 
          onClick={() => setPage(p => p + 1)}
          disabled={page >= totalPages - 1}
        >
          Next
        </button>
      </div>
    </div>
  );
}
```

## Performance рекомендации

1. **Оптимальный размер страницы**: 10-50 записей
2. **Кэширование**: Кэшируйте результаты на клиенте
3. **Prefetching**: Предзагружайте следующую страницу
4. **Индексы БД**: Убедитесь, что есть индекс на поле сортировки

## Стандарты и RFC

Реализация основана на:
- [RFC 9110 - HTTP Semantics: Range Requests](https://www.rfc-editor.org/rfc/rfc9110.html#name-range)
- [MDN - Content-Range Header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Range)

