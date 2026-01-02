package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"url-shortener/internal/database"
	"url-shortener/internal/models"

	"github.com/gin-gonic/gin"
)

// LinkVisitHandler обработчик для операций с посещениями
type LinkVisitHandler struct {
	queries *database.Queries
}

// NewLinkVisitHandler создает новый обработчик
func NewLinkVisitHandler(db *sql.DB) *LinkVisitHandler {
	return &LinkVisitHandler{
		queries: database.New(db),
	}
}

// formatLinkVisit форматирует запись посещения для ответа
func (h *LinkVisitHandler) formatLinkVisit(visit database.LinkVisit) models.LinkVisit {
	var userAgent, referer *string
	
	if visit.UserAgent.Valid {
		userAgent = &visit.UserAgent.String
	}
	
	if visit.Referer.Valid {
		referer = &visit.Referer.String
	}
	
	return models.LinkVisit{
		ID:        visit.ID,
		LinkID:    visit.LinkID,
		IP:        visit.Ip,
		UserAgent: userAgent,
		Referer:   referer,
		Status:    visit.Status,
		CreatedAt: visit.CreatedAt,
	}
}

// GetAllLinkVisits возвращает все посещения с поддержкой пагинации
func (h *LinkVisitHandler) GetAllLinkVisits(c *gin.Context) {
	rangeParam := c.Query("range")

	// Если параметр range не указан, возвращаем все посещения
	if rangeParam == "" {
		visits, err := h.queries.GetAllLinkVisits(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch link visits"})
			return
		}

		result := make([]models.LinkVisit, len(visits))
		for i, visit := range visits {
			result[i] = h.formatLinkVisit(visit)
		}

		c.JSON(http.StatusOK, result)
		return
	}

	// Парсим range параметр
	paginationRange, err := models.ParseRange(rangeParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid range parameter: %v", err)})
		return
	}

	// Получаем общее количество записей
	totalCount, err := h.queries.CountLinkVisits(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count link visits"})
		return
	}

	// Получаем посещения с пагинацией
	visits, err := h.queries.GetLinkVisitsWithPagination(c.Request.Context(), database.GetLinkVisitsWithPaginationParams{
		Limit:  paginationRange.Limit(),
		Offset: paginationRange.Offset(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch link visits"})
		return
	}

	// Форматируем результат
	result := make([]models.LinkVisit, len(visits))
	for i, visit := range visits {
		result[i] = h.formatLinkVisit(visit)
	}

	// Вычисляем фактический end
	actualEnd := paginationRange.Start + len(visits)

	// Добавляем заголовок Content-Range
	contentRange := fmt.Sprintf("link_visits %d-%d/%d", 
		paginationRange.Start, actualEnd-1, int(totalCount))
	c.Header("Content-Range", contentRange)

	c.JSON(http.StatusOK, result)
}
