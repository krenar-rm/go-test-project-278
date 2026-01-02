package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"url-shortener/internal/database"
	"url-shortener/internal/models"
	"url-shortener/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// LinkHandler обработчик для операций со ссылками
type LinkHandler struct {
	queries *database.Queries
	baseURL string
}

// NewLinkHandler создает новый обработчик
func NewLinkHandler(db *sql.DB) *LinkHandler {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &LinkHandler{
		queries: database.New(db),
		baseURL: baseURL,
	}
}

// formatLink форматирует ссылку для ответа
func (h *LinkHandler) formatLink(link database.Link) models.Link {
	return models.Link{
		ID:          link.ID,
		OriginalURL: link.OriginalUrl,
		ShortName:   link.ShortName,
		ShortURL:    fmt.Sprintf("%s/r/%s", h.baseURL, link.ShortName),
		CreatedAt:   link.CreatedAt,
	}
}

// GetAllLinks возвращает все ссылки с поддержкой пагинации
func (h *LinkHandler) GetAllLinks(c *gin.Context) {
	rangeParam := c.Query("range")

	// Если параметр range не указан, возвращаем все ссылки
	if rangeParam == "" {
		links, err := h.queries.GetAllLinks(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
			return
		}

		result := make([]models.Link, len(links))
		for i, link := range links {
			result[i] = h.formatLink(link)
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
	totalCount, err := h.queries.CountLinks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count links"})
		return
	}

	// Получаем ссылки с пагинацией
	links, err := h.queries.GetLinksWithPagination(c.Request.Context(), database.GetLinksWithPaginationParams{
		Limit:  paginationRange.Limit(),
		Offset: paginationRange.Offset(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
		return
	}

	// Форматируем результат
	result := make([]models.Link, len(links))
	for i, link := range links {
		result[i] = h.formatLink(link)
	}

	// Вычисляем фактический end (количество возвращенных записей)
	actualEnd := paginationRange.Start + len(links)

	// Добавляем заголовок Content-Range
	contentRange := models.FormatContentRange(paginationRange.Start, actualEnd, int(totalCount))
	c.Header("Content-Range", contentRange)

	c.JSON(http.StatusOK, result)
}

// GetLinkByID возвращает ссылку по ID
func (h *LinkHandler) GetLinkByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	link, err := h.queries.GetLinkByID(c.Request.Context(), int32(id))
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch link"})
		return
	}

	c.JSON(http.StatusOK, h.formatLink(link))
}

// CreateLink создает новую ссылку
func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req models.CreateLinkRequest
	
	// Проверяем корректность JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		// Если это ошибка валидации - возвращаем 422
		if _, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusUnprocessableEntity, utils.ErrorResponse{
				Errors: utils.FormatValidationErrors(err),
			})
			return
		}
		// Если это ошибка парсинга JSON - возвращаем 400
		c.JSON(http.StatusBadRequest, utils.SimpleErrorResponse{
			Error: "invalid request",
		})
		return
	}

	// Генерируем short_name если не указан
	shortName := req.ShortName
	if shortName == "" {
		var err error
		shortName, err = models.GenerateShortName()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short name"})
			return
		}
	}

	// Создаем ссылку
	link, err := h.queries.CreateLink(c.Request.Context(), database.CreateLinkParams{
		OriginalUrl: req.OriginalURL,
		ShortName:   shortName,
	})
	if err != nil {
		// Проверяем, не конфликт ли это по unique constraint
		if utils.IsDuplicateKeyError(err) {
			c.JSON(http.StatusUnprocessableEntity, utils.ErrorResponse{
				Errors: utils.FormatDuplicateKeyError(err, "short_name"),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create link"})
		return
	}

	c.JSON(http.StatusCreated, h.formatLink(link))
}

// UpdateLink обновляет существующую ссылку
func (h *LinkHandler) UpdateLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req models.UpdateLinkRequest
	
	// Проверяем корректность JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		// Если это ошибка валидации - возвращаем 422
		if _, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusUnprocessableEntity, utils.ErrorResponse{
				Errors: utils.FormatValidationErrors(err),
			})
			return
		}
		// Если это ошибка парсинга JSON - возвращаем 400
		c.JSON(http.StatusBadRequest, utils.SimpleErrorResponse{
			Error: "invalid request",
		})
		return
	}

	// Проверяем, существует ли ссылка
	_, err = h.queries.GetLinkByID(c.Request.Context(), int32(id))
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch link"})
		return
	}

	// Обновляем ссылку
	link, err := h.queries.UpdateLink(c.Request.Context(), database.UpdateLinkParams{
		ID:          int32(id),
		OriginalUrl: req.OriginalURL,
		ShortName:   req.ShortName,
	})
	if err != nil {
		// Проверяем конфликт по unique constraint
		if utils.IsDuplicateKeyError(err) {
			c.JSON(http.StatusUnprocessableEntity, utils.ErrorResponse{
				Errors: utils.FormatDuplicateKeyError(err, "short_name"),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update link"})
		return
	}

	c.JSON(http.StatusOK, h.formatLink(link))
}

// DeleteLink удаляет ссылку
func (h *LinkHandler) DeleteLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Проверяем, существует ли ссылка
	_, err = h.queries.GetLinkByID(c.Request.Context(), int32(id))
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch link"})
		return
	}

	// Удаляем ссылку
	err = h.queries.DeleteLink(c.Request.Context(), int32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete link"})
		return
	}

	c.Status(http.StatusNoContent)
}
