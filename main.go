package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
)

type Article struct {
	Id      string `db:"id" json:"id"`
	Title   string `db:"title" json:"title"`
	Content string `db:"content" json:"content"`
	Author  string `db:"author" json:"author"`
}

var app = pocketbase.New()

func main() {
	log.Println("server start")
	staticDir := http.Dir("./public")
	fileServer := http.FileServer(staticDir)
	http.Handle("/", fileServer)
	go func() {
		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("server start2")

	// serves static files from the provided public dir (if exists)
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS("./pb_public"), false))
		e.Router.GET("/hello", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello world!")
		}, apis.ActivityLogger(app))
		e.Router.GET("/article", getArticleList, apis.ActivityLogger(app))
		e.Router.POST("/article", postArticle, apis.ActivityLogger(app))
		e.Router.PUT("/article", updateArticle, apis.ActivityLogger(app))
		e.Router.DELETE("/article", deleteArticle, apis.ActivityLogger(app))
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
func getArticleList(c echo.Context) error {
	posts := []Article{}
	err := app.Dao().DB().
		Select("*").
		From("article").
		Limit(20).
		OrderBy("created DESC").
		All(&posts)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(&posts)
	return c.JSON(200, &posts)
}
func postArticle(c echo.Context) error {
	collection, err := app.Dao().FindCollectionByNameOrId("article")
	if err != nil {
		return err
	}

	record := models.NewRecord(collection)
	form := forms.NewRecordUpsert(app, record)

	data := apis.RequestInfo(c).Data
	form.LoadData(map[string]any{
		"title":   data["title"],
		"content": data["content"],
		"author":  data["author"],
	})
	if err := form.Submit(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, record)
}
func updateArticle(c echo.Context) error {
	data := Article{}
	if err := c.Bind(&data); err != nil {
		return apis.NewBadRequestError("Failed to read request data", err)
	}

	record, err := app.Dao().FindRecordById("article", data.Id)
	if err != nil {
		return err
	}
	form := forms.NewRecordUpsert(app, record)
	form.LoadData(map[string]any{
		"title":   data.Title,
		"content": data.Content,
		"author":  data.Author,
	})
	// validate and submit (internally it calls app.Dao().SaveRecord(record) in a transaction)
	if err := form.Submit(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, record)
}
func deleteArticle(c echo.Context) error {
	data := Article{}
	if err := c.Bind(&data); err != nil {
		return apis.NewBadRequestError("Failed to read request data", err)
	}

	record, err := app.Dao().FindRecordById("article", data.Id)
	if err != nil {
		return err
	}
	if err := app.Dao().DeleteRecord(record); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, nil)
}
