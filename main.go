package main

import (
	generated "code/db/generated"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// создание маршрутизатора Gin
func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// подключаем монитор просмотра ошибок
	router.Use(sentrygin.New(sentrygin.Options{}))
	// подключаем инструмент восстановления сбоев
	router.Use(gin.Recovery())
	// подключаем логгер
	router.Use(gin.Logger())
	return router
}

// получение всех записей
func listLinks(db *generated.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		var paginParams generated.ListLinksParams
		// получаем параметры для пагинации
		rangeLinks := c.DefaultQuery("range", "[0,20]")
		// задаём параметры по умолчанию
		limit := 20
		offset := 0
		// задаём регулярное выражение для поиска всех чисел
		re := regexp.MustCompile(`\d+`)
		// получаем лимит записей на странице и сдвиг для вывода записей
		numRange := re.FindAllString(rangeLinks, -1)
		// проверяем корректность ввода данных
		if len(numRange) != 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "the range must be specified by two numbers, example: [1,4]"})
			return
		}
		idx0, _ := strconv.Atoi(numRange[0])
		idx1, _ := strconv.Atoi(numRange[1])
		// проверка на положительные значения
		if idx0 < 0 || idx1 < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "the range value must be positive"})
			return
		}
		// если первый индекс меньше второго
		if idx0 < idx1 {
			limit = idx1 - idx0
			offset = idx0
		}
		// если индексы равны
		if idx0 == idx1 {
			limit = 1
			offset = idx0
		}
		if idx0 > idx1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "range values ​​are specified incorrectly"})
			return
		}
		// ограничение максимального числа записей на странице
		if limit > 20 {
			limit = 20
		}
		paginParams.Limit = int32(limit)
		paginParams.Offset = int32(offset)
		links, err := db.ListLinks(c, paginParams)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "links not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		count, err := db.CounterLinks(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to count the number of records"})
			return
		}
		headerVal := fmt.Sprintf("links: %d-%d/%d", idx0, idx1, count)
		c.JSON(http.StatusOK, links)
		c.Header("Content-Range", headerVal)
	}
}

// создание новой записи
func createLink(db *generated.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		var link generated.CreateLinkParams
		if err := c.ShouldBindJSON(&link); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := db.CreateLink(c, link)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, id)
	}
}

// обновление записи
func updateLink(db *generated.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		res := db.UpdateLink(c, updLink)
		if res != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	}
}

// получение одной записи
func getLinkFromId(db *generated.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		link, err := db.GetLink(c, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "link not found",
			})
			return
		}
		c.JSON(http.StatusOK, link)
	}
}

// удаление записи
func deleteLink(db *generated.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// проверяем наличие записи
		_, err = db.GetLink(c, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "the link does not exist",
			})
			return
		}
		// удаляем ссылку
		err = db.DeleteLink(c, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "error deleting links",
			})
			return
		}
		c.JSON(http.StatusNoContent, id)
	}
}

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
		Dsn: "https://b81aac0d2c97f7e747a4fb8aeb0d72ea@o4511376473587712.ingest.de.sentry.io/4511376481189968",
	})
	if errSentry != nil {
		log.Fatalf("sentry initialization failed: %s", errSentry)
	}
	defer sentry.Flush(2 * time.Second)

	// создаём маршрутизатор
	r := setupRouter()
	// регистрируем маршруты
	r.GET("api/links", listLinks(queries))
	r.GET("api/links/:id", getLinkFromId(queries))
	r.POST("api/links", createLink(queries))
	r.PUT("api/links/:id", updateLink(queries))
	r.DELETE("api/links/:id", deleteLink(queries))
	// запускаем сервер на порту 8080
	r.Run(":8080")
}
