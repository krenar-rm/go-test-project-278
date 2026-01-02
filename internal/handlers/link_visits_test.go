package handlers

import (
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

func setupTestDBForVisits(t *testing.T) *sql.DB {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Очищаем таблицы перед тестами (link_visits сначала из-за foreign key)
	_, err = db.Exec("TRUNCATE TABLE link_visits, links RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	return db
}

func TestGetAllLinkVisits(t *testing.T) {
	db := setupTestDBForVisits(t)
	defer db.Close()

	// Создаем тестовую ссылку
	queries := database.New(db)
	ctx := context.Background()
	link, _ := queries.CreateLink(ctx, database.CreateLinkParams{
		OriginalUrl: "https://example.com/test",
		ShortName:   "testlink",
	})

	// Создаем несколько посещений
	for i := 1; i <= 3; i++ {
		_, _ = queries.CreateLinkVisit(ctx, database.CreateLinkVisitParams{
			LinkID: link.ID,
			Ip:     fmt.Sprintf("192.168.1.%d", i),
			UserAgent: sql.NullString{
				String: fmt.Sprintf("TestAgent/%d", i),
				Valid:  true,
			},
			Referer: sql.NullString{
				String: "https://google.com",
				Valid:  true,
			},
			Status: 302,
		})
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkVisitHandler(db)
	router.GET("/api/link_visits", handler.GetAllLinkVisits)

	req, _ := http.NewRequest("GET", "/api/link_visits", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.LinkVisit
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(response), 3)
	assert.Equal(t, link.ID, response[0].LinkID)
}

func TestGetLinkVisitsWithPagination(t *testing.T) {
	db := setupTestDBForVisits(t)
	defer db.Close()

	// Создаем тестовую ссылку
	queries := database.New(db)
	ctx := context.Background()
	link, _ := queries.CreateLink(ctx, database.CreateLinkParams{
		OriginalUrl: "https://example.com/test",
		ShortName:   "testlink",
	})

	// Создаем 15 посещений
	for i := 1; i <= 15; i++ {
		_, _ = queries.CreateLinkVisit(ctx, database.CreateLinkVisitParams{
			LinkID: link.ID,
			Ip:     fmt.Sprintf("192.168.1.%d", i),
			UserAgent: sql.NullString{
				String: fmt.Sprintf("TestAgent/%d", i),
				Valid:  true,
			},
			Status: 302,
		})
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewLinkVisitHandler(db)
	router.GET("/api/link_visits", handler.GetAllLinkVisits)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/link_visits?range=%s", tt.rangeParam)
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response []models.LinkVisit
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(response))

				if tt.checkContentRange {
					contentRange := w.Header().Get("Content-Range")
					expectedContentRange := fmt.Sprintf("link_visits %d-%d/%d",
						tt.expectedStart, tt.expectedEnd-1, tt.expectedTotal)
					assert.Equal(t, expectedContentRange, contentRange)
				}
			}
		})
	}
}
