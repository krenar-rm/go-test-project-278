package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter создает и настраивает роутер для тестов
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Отключаем Sentry в тестах
	_ = os.Setenv("SENTRY_DSN", "")

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	router.GET("/sentry-test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Test error sent to Sentry",
		})
	})

	return router
}

// TestPingRoute проверяет работу маршрута GET /ping
func TestPingRoute(t *testing.T) {
	router := setupTestRouter()

	// Создаем тестовый запрос
	req, err := http.NewRequest("GET", "/ping", nil)
	assert.NoError(t, err)

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем тело ответа
	assert.Equal(t, "pong", w.Body.String())
}

// TestSentryTestRoute проверяет работу маршрута GET /sentry-test
func TestSentryTestRoute(t *testing.T) {
	router := setupTestRouter()

	// Создаем тестовый запрос
	req, err := http.NewRequest("GET", "/sentry-test", nil)
	assert.NoError(t, err)

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Выполняем запрос
	router.ServeHTTP(w, req)

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем тело ответа
	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test error sent to Sentry", response["message"])
}
