package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"url-shortener/internal/database"
	"url-shortener/internal/models"

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

	// Очищаем таблицу перед тестами
	_, err = db.Exec("TRUNCATE TABLE links RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	return db
}

func TestCreateLink(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkHandler(db)
	router.POST("/api/links", handler.CreateLink)

	tests := []struct {
		name       string
		payload    models.CreateLinkRequest
		wantStatus int
		wantError  bool
	}{
		{
			name: "Create link with custom short name",
			payload: models.CreateLinkRequest{
				OriginalURL: "https://example.com/long-url",
				ShortName:   "test123",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "Create link with auto-generated short name",
			payload: models.CreateLinkRequest{
				OriginalURL: "https://example.com/another-long-url",
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "Create link with duplicate short name",
			payload: models.CreateLinkRequest{
				OriginalURL: "https://example.com/duplicate",
				ShortName:   "test123",
			},
			wantStatus: http.StatusConflict,
			wantError:  true,
		},
		{
			name: "Create link with invalid URL",
			payload: models.CreateLinkRequest{
				OriginalURL: "not-a-url",
				ShortName:   "invalid",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/links", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				var response models.Link
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotZero(t, response.ID)
				assert.Equal(t, tt.payload.OriginalURL, response.OriginalURL)
				if tt.payload.ShortName != "" {
					assert.Equal(t, tt.payload.ShortName, response.ShortName)
				} else {
					assert.NotEmpty(t, response.ShortName)
				}
				assert.Contains(t, response.ShortURL, response.ShortName)
			}
		})
	}
}

func TestGetAllLinks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Создаем тестовые ссылки
	queries := database.New(db)
	ctx := context.Background()
	_, _ = queries.CreateLink(ctx, database.CreateLinkParams{
		OriginalUrl: "https://example.com/1",
		ShortName:   "link1",
	})
	_, _ = queries.CreateLink(ctx, database.CreateLinkParams{
		OriginalUrl: "https://example.com/2",
		ShortName:   "link2",
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkHandler(db)
	router.GET("/api/links", handler.GetAllLinks)

	req, _ := http.NewRequest("GET", "/api/links", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Link
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(response), 2)
}

func TestGetAllLinksWithPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Создаем 15 тестовых ссылок
	queries := database.New(db)
	ctx := context.Background()
	for i := 1; i <= 15; i++ {
		_, _ = queries.CreateLink(ctx, database.CreateLinkParams{
			OriginalUrl: fmt.Sprintf("https://example.com/%d", i),
			ShortName:   fmt.Sprintf("link%d", i),
		})
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkHandler(db)
	router.GET("/api/links", handler.GetAllLinks)

	tests := []struct {
		name              string
		rangeParam        string
		expectedStatus    int
		expectedCount     int
		expectedStart     int
		expectedEnd       int
		expectedTotal     int
		checkContentRange bool
	}{
		{
			name:              "First 10 items",
			rangeParam:        "[0,10]",
			expectedStatus:    http.StatusOK,
			expectedCount:     10,
			expectedStart:     0,
			expectedEnd:       10,
			expectedTotal:     15,
			checkContentRange: true,
		},
		{
			name:              "Items 5-10",
			rangeParam:        "[5,10]",
			expectedStatus:    http.StatusOK,
			expectedCount:     5,
			expectedStart:     5,
			expectedEnd:       10,
			expectedTotal:     15,
			checkContentRange: true,
		},
		{
			name:              "Items 10-20 (exceeds total)",
			rangeParam:        "[10,20]",
			expectedStatus:    http.StatusOK,
			expectedCount:     5,
			expectedStart:     10,
			expectedEnd:       15,
			expectedTotal:     15,
			checkContentRange: true,
		},
		{
			name:           "Invalid range format",
			rangeParam:     "[0]",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid range values",
			rangeParam:     "[10,5]",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Negative start",
			rangeParam:     "[-1,10]",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/links?range=%s", tt.rangeParam)
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []models.Link
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(response))

				if tt.checkContentRange {
					contentRange := w.Header().Get("Content-Range")
					expectedContentRange := fmt.Sprintf("links %d-%d/%d",
						tt.expectedStart, tt.expectedEnd-1, tt.expectedTotal)
					assert.Equal(t, expectedContentRange, contentRange)
				}
			}
		})
	}
}

func TestGetLinkByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Создаем тестовую ссылку
	queries := database.New(db)
	ctx := context.Background()
	link, _ := queries.CreateLink(ctx, database.CreateLinkParams{
		OriginalUrl: "https://example.com/test",
		ShortName:   "testlink",
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkHandler(db)
	router.GET("/api/links/:id", handler.GetLinkByID)

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "Get existing link",
			id:         "1",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "Get non-existent link",
			id:         "999",
			wantStatus: http.StatusNotFound,
			wantError:  true,
		},
		{
			name:       "Get link with invalid ID",
			id:         "invalid",
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/links/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if !tt.wantError {
				var response models.Link
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, link.ID, response.ID)
			}
		})
	}
}

func TestUpdateLink(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Создаем тестовую ссылку
	queries := database.New(db)
	ctx := context.Background()
	_, _ = queries.CreateLink(ctx, database.CreateLinkParams{
		OriginalUrl: "https://example.com/original",
		ShortName:   "orig",
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkHandler(db)
	router.PUT("/api/links/:id", handler.UpdateLink)

	updateReq := models.UpdateLinkRequest{
		OriginalURL: "https://example.com/updated",
		ShortName:   "updated",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/links/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Link
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, updateReq.OriginalURL, response.OriginalURL)
	assert.Equal(t, updateReq.ShortName, response.ShortName)
}

func TestDeleteLink(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Создаем тестовую ссылку
	queries := database.New(db)
	ctx := context.Background()
	_, _ = queries.CreateLink(ctx, database.CreateLinkParams{
		OriginalUrl: "https://example.com/to-delete",
		ShortName:   "todel",
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkHandler(db)
	router.DELETE("/api/links/:id", handler.DeleteLink)

	req, _ := http.NewRequest("DELETE", "/api/links/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Проверяем, что ссылка удалена
	_, err := queries.GetLinkByID(ctx, 1)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}
