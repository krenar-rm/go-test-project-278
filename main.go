package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"url-shortener/internal/database"
	"url-shortener/internal/handlers"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	envDevelopment = "development"
)

func main() {
	// Загружаем .env файл для локальной разработки (игнорируем ошибку в production)
	_ = godotenv.Load()

	// Инициализация Sentry
	sentryDSN := os.Getenv("SENTRY_DSN")
	if sentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDSN,
			TracesSampleRate: 1.0,
			Environment:      getEnvironment(),
		})
		if err != nil {
			log.Printf("Sentry initialization failed: %v", err)
		} else {
			log.Println("Sentry initialized successfully")
			defer sentry.Flush(2 * time.Second)
		}
	} else {
		log.Println("SENTRY_DSN not set, Sentry is disabled")
	}

	// Подключение к базе данных
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	log.Println("Database connected successfully")

	// Создаем роутер
	router := setupRouter(db, sentryDSN)

	// Порт для приложения
	// В Docker всегда используем 8080 (Caddy проксирует с 80 на 8080)
	// Локально можно переопределить через APP_PORT
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// Запускаем приложение
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initDB инициализирует подключение к базе данных
func initDB() (*sql.DB, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("Warning: DATABASE_URL not set, using default")
		databaseURL = "postgres://postgres:password@localhost:5432/urlshortener?sslmode=disable"
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// setupRouter настраивает routes и middleware
func setupRouter(db *sql.DB, sentryDSN string) *gin.Engine {
	router := gin.Default()

	// Настройка trusted proxies для корректного получения IP за Cloudflare/Caddy
	router.TrustedPlatform = gin.PlatformCloudflare

	// Настройка CORS
	corsConfig := cors.DefaultConfig()
	if getEnvironment() == envDevelopment {
		// В режиме разработки разрешаем localhost:5173 (фронтенд)
		corsConfig.AllowOrigins = []string{"http://localhost:5173"}
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
		corsConfig.AllowCredentials = true
	} else {
		// В продакшене используем BASE_URL или домен из переменной окружения
		allowedOrigin := os.Getenv("FRONTEND_URL")
		if allowedOrigin == "" {
			allowedOrigin = os.Getenv("BASE_URL")
		}
		if allowedOrigin != "" {
			corsConfig.AllowOrigins = []string{allowedOrigin}
		}
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	}
	router.Use(cors.New(corsConfig))

	// Подключаем Sentry middleware (если DSN установлен)
	if sentryDSN != "" {
		router.Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
		}))
	}

	// Создаем handler для ссылок и посещений
	linkHandler := handlers.NewLinkHandler(db)
	linkVisitHandler := handlers.NewLinkVisitHandler(db)

	// Health check
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// Тестовые маршруты для Sentry (только для development)
	if getEnvironment() == envDevelopment {
		router.GET("/sentry-test", func(c *gin.Context) {
			sentry.CaptureMessage("Test error from /sentry-test endpoint")
			c.JSON(200, gin.H{"message": "Test error sent to Sentry"})
		})

		router.GET("/sentry-panic", func(c *gin.Context) {
			panic("Test panic for Sentry!")
		})
	}

	// API routes для CRUD операций
	api := router.Group("/api")
	{
		// Ссылки
		api.GET("/links", linkHandler.GetAllLinks)
		api.POST("/links", linkHandler.CreateLink)
		api.GET("/links/:id", linkHandler.GetLinkByID)
		api.PUT("/links/:id", linkHandler.UpdateLink)
		api.DELETE("/links/:id", linkHandler.DeleteLink)

		// Посещения
		api.GET("/link_visits", linkVisitHandler.GetAllLinkVisits)
	}

	// Редирект по короткой ссылке с записью аналитики
	router.GET("/r/:shortName", func(c *gin.Context) {
		shortName := c.Param("shortName")

		queries := database.New(db)
		link, err := queries.GetLinkByShortName(c.Request.Context(), shortName)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short link not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch link"})
			return
		}

		// Определяем код редиректа (используем 302 Found для временного редиректа)
		redirectStatus := http.StatusFound // 302

		// Получаем информацию о клиенте
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		referer := c.Request.Referer()

		// Сохраняем посещение в БД (асинхронно, чтобы не замедлять редирект)
		go func() {
			ctx := c.Request.Context()
			_, _ = queries.CreateLinkVisit(ctx, database.CreateLinkVisitParams{
				LinkID: link.ID,
				Ip:     clientIP,
				UserAgent: sql.NullString{
					String: userAgent,
					Valid:  userAgent != "",
				},
				Referer: sql.NullString{
					String: referer,
					Valid:  referer != "",
				},
				Status: int32(http.StatusFound), // #nosec G115
			})
		}()

		c.Redirect(redirectStatus, link.OriginalUrl)
	})

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "The requested resource was not found",
			"path":    c.Request.URL.Path,
		})
	})

	return router
}

// getEnvironment возвращает окружение приложения
func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		return envDevelopment
	}
	return env
}
