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
	"url-shortener/internal/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setupTestDBForValidation(t *testing.T) *sql.DB {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Очищаем таблицы
	_, err = db.Exec("TRUNCATE TABLE link_visits, links RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	return db
}

func TestCreateLinkValidation(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkHandler(db)
	router.POST("/api/links", handler.CreateLink)

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		checkError     func(*testing.T, map[string]interface{})
	}{
		{
			name:           "Invalid JSON",
			payload:        `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			checkError: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, "invalid request", resp["error"])
			},
		},
		{
			name:           "Missing original_url",
			payload:        `{"short_name": "test"}`,
			expectedStatus: http.StatusUnprocessableEntity,
			checkError: func(t *testing.T, resp map[string]interface{}) {
				errors := resp["errors"].(map[string]interface{})
				assert.Contains(t, errors, "original_url")
			},
		},
		{
			name:           "Invalid URL",
			payload:        `{"original_url": "not-a-url"}`,
			expectedStatus: http.StatusUnprocessableEntity,
			checkError: func(t *testing.T, resp map[string]interface{}) {
				errors := resp["errors"].(map[string]interface{})
				assert.Equal(t, "must be a valid URL", errors["original_url"])
			},
		},
		{
			name:           "Short name too short",
			payload:        `{"original_url": "https://example.com", "short_name": "ab"}`,
			expectedStatus: http.StatusUnprocessableEntity,
			checkError: func(t *testing.T, resp map[string]interface{}) {
				errors := resp["errors"].(map[string]interface{})
				assert.Contains(t, errors["short_name"], "at least 3 characters")
			},
		},
		{
			name:           "Short name too long",
			payload:        `{"original_url": "https://example.com", "short_name": "this_is_a_very_long_short_name_that_exceeds_32_characters"}`,
			expectedStatus: http.StatusUnprocessableEntity,
			checkError: func(t *testing.T, resp map[string]interface{}) {
				errors := resp["errors"].(map[string]interface{})
				assert.Contains(t, errors["short_name"], "at most 32 characters")
			},
		},
		{
			name:           "Valid request with alphanumeric",
			payload:        `{"original_url": "https://example.com", "short_name": "test123"}`,
			expectedStatus: http.StatusCreated,
			checkError:     nil,
		},
		{
			name:           "Valid request with dashes and underscores",
			payload:        `{"original_url": "https://example.com", "short_name": "test-123_abc"}`,
			expectedStatus: http.StatusCreated,
			checkError:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/links", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkError != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				tt.checkError(t, response)
			}
		})
	}
}

func TestCreateLinkDuplicateShortName(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Создаем существующую ссылку
	queries := database.New(db)
	ctx := context.Background()
	_, _ = queries.CreateLink(ctx, database.CreateLinkParams{
		OriginalUrl: "https://example.com/1",
		ShortName:   "duplicate",
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkHandler(db)
	router.POST("/api/links", handler.CreateLink)

	payload := `{"original_url": "https://example.com/2", "short_name": "duplicate"}`
	req, _ := http.NewRequest("POST", "/api/links", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "short name already in use", response.Errors["short_name"])
}

func TestUpdateLinkValidation(t *testing.T) {
	db := setupTestDBForValidation(t)
	defer db.Close()

	// Создаем тестовую ссылку
	queries := database.New(db)
	ctx := context.Background()
	link, _ := queries.CreateLink(ctx, database.CreateLinkParams{
		OriginalUrl: "https://example.com/original",
		ShortName:   "original",
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkHandler(db)
	router.PUT("/api/links/:id", handler.UpdateLink)

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		checkError     func(*testing.T, map[string]interface{})
	}{
		{
			name:           "Invalid JSON",
			payload:        `{invalid}`,
			expectedStatus: http.StatusBadRequest,
			checkError: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, "invalid request", resp["error"])
			},
		},
		{
			name:           "Missing original_url",
			payload:        `{"short_name": "updated"}`,
			expectedStatus: http.StatusUnprocessableEntity,
			checkError: func(t *testing.T, resp map[string]interface{}) {
				errors := resp["errors"].(map[string]interface{})
				assert.Contains(t, errors, "original_url")
			},
		},
		{
			name:           "Missing short_name",
			payload:        `{"original_url": "https://example.com"}`,
			expectedStatus: http.StatusUnprocessableEntity,
			checkError: func(t *testing.T, resp map[string]interface{}) {
				errors := resp["errors"].(map[string]interface{})
				assert.Contains(t, errors, "short_name")
			},
		},
		{
			name:           "Invalid URL",
			payload:        `{"original_url": "invalid", "short_name": "updated"}`,
			expectedStatus: http.StatusUnprocessableEntity,
			checkError: func(t *testing.T, resp map[string]interface{}) {
				errors := resp["errors"].(map[string]interface{})
				assert.Equal(t, "must be a valid URL", errors["original_url"])
			},
		},
		{
			name:           "Valid update",
			payload:        `{"original_url": "https://example.com/updated", "short_name": "updated"}`,
			expectedStatus: http.StatusOK,
			checkError:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/links/%d", link.ID)
			req, _ := http.NewRequest("PUT", url, bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkError != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				tt.checkError(t, response)
			}
		})
	}
}

