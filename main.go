package main

import (
	generated "code/db/generated"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// подключаемся с БД
	var err error

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
	// маршрут для получения всех записей
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
	// маршрут для добавления новой записи
	router.POST("/api/links", func(c *gin.Context) {
		var link generated.CreateLinkParams
		if err := c.ShouldBindJSON(&link); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := queries.CreateLink(c, link)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, id)
	})
	// маршрут для изменения записи
	router.PUT("/api/links/:id", func(c *gin.Context) {
		var updLink generated.UpdateLinkParams
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updLink.ID = id
		if err := c.ShouldBindJSON(&updLink); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res := queries.UpdateLink(c, updLink)
		if res != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})
	// маршрут для получения одной записи
	router.GET("/api/links/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		link, err := queries.GetLink(c, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "link not found",
			})
			return
		}
		c.JSON(http.StatusOK, link)
	})
	// маршрут для удаления записи
	router.DELETE("/api/links/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err = queries.DeleteLink(c, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "link has not been removed",
			})
			return
		}
		c.JSON(http.StatusNoContent, id)
	})
	// запускаем сервер на порту 8080
	router.Run(":8080")
}
