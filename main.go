package main

import (
	generated "code/db/generated"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// подключаемся с БД
	var err error
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL не установлена")
	}

	// Инициализация пула соединений
	conn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка подключения к БД: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	queries := generated.New(conn)
	// подключаем мониторинг ошибок
	errSentry := sentry.Init(sentry.ClientOptions{
		Dsn:            "https://b81aac0d2c97f7e747a4fb8aeb0d72ea@o4511376473587712.ingest.de.sentry.io/4511376481189968",
		Debug:          true,
		SendDefaultPII: true,
	})
	if errSentry != nil {
		log.Fatalf("sentry.Init: %s", errSentry)
	}
	defer sentry.Flush(2 * time.Second)
	// создаём маршрутизатор Gin
	router := gin.Default()
	// Once it's done, you can attach the handler as one of your middleware
	router.Use(sentrygin.New(sentrygin.Options{}))
	// подключаем инструмент восстановления сбоев
	router.Use(gin.Recovery())
	// подулючаем логгер
	router.Use(gin.Logger())
	// создаём точку входа
	router.GET("/api/links", func(c *gin.Context) {
		links, err := queries.ListLinks(c)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "links not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		c.JSON(http.StatusOK, links)
	})
	// запускаем сервер на порту 8080
	router.Run(":8080")
}
