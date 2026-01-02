package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"url-shortener/internal/database"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Очищаем таблицы перед тестами
	_, err = db.Exec("TRUNCATE TABLE link_visits, links RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	return db
}

func TestRedirectRoute(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Создаем тестовую ссылку
	queries := database.New(db)
	ctx := context.Background()
	link, _ := queries.CreateLink(ctx, database.CreateLinkParams{
		OriginalUrl: "https://example.com/target",
		ShortName:   "testlink",
	})

	gin.SetMode(gin.TestMode)
	router := setupRouter(db, "")

	// Тест редиректа
	req, _ := http.NewRequest("GET", "/r/testlink", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("Referer", "https://google.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Проверяем редирект
	assert.Equal(t, http.StatusFound, w.Code) // 302
	assert.Equal(t, "https://example.com/target", w.Header().Get("Location"))

	// Даем время на сохранение посещения (асинхронное)
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что посещение записано
	visits, err := queries.GetVisitsByLinkID(ctx, link.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(visits))
	assert.Equal(t, link.ID, visits[0].LinkID)
	assert.Equal(t, int32(302), visits[0].Status)
	assert.True(t, visits[0].UserAgent.Valid)
	assert.Equal(t, "TestAgent/1.0", visits[0].UserAgent.String)
}

func TestRedirectNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	gin.SetMode(gin.TestMode)
	router := setupRouter(db, "")

	req, _ := http.NewRequest("GET", "/r/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

