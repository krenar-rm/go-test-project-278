package models

import (
	"fmt"
	"strconv"
	"strings"
)

// PaginationRange представляет диапазон для пагинации
type PaginationRange struct {
	Start int
	End   int
}

// ParseRange парсит параметр range вида "[0,10]"
func ParseRange(rangeStr string) (*PaginationRange, error) {
	// Убираем квадратные скобки
	rangeStr = strings.TrimSpace(rangeStr)
	rangeStr = strings.TrimPrefix(rangeStr, "[")
	rangeStr = strings.TrimSuffix(rangeStr, "]")

	// Разделяем по запятой
	parts := strings.Split(rangeStr, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format, expected [start,end]")
	}

	// Парсим начало
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start value: %w", err)
	}

	// Парсим конец
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end value: %w", err)
	}

	// Валидация
	if start < 0 {
		return nil, fmt.Errorf("start must be >= 0")
	}
	if end < start {
		return nil, fmt.Errorf("end must be >= start")
	}

	return &PaginationRange{
		Start: start,
		End:   end,
	}, nil
}

// Limit возвращает количество записей для выборки
func (r *PaginationRange) Limit() int32 {
	return int32(r.End - r.Start)
}

// Offset возвращает смещение для выборки
func (r *PaginationRange) Offset() int32 {
	return int32(r.Start)
}

// FormatContentRange форматирует заголовок Content-Range
// Пример: "links 0-9/42" означает записи с 0 по 9 из 42 всего
func FormatContentRange(start, end, total int) string {
	// Если записей нет
	if total == 0 {
		return "links */0"
	}

	// Если end больше total, корректируем
	if end > total {
		end = total
	}

	// Формат: "links start-end/total"
	// Внимание: end - это позиция последнего элемента (inclusive)
	return fmt.Sprintf("links %d-%d/%d", start, end-1, total)
}
