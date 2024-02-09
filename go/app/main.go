package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
	dbPath = "../db/mercari.sqlite3"
)

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

type Item struct {
	Name          string `json:"name"`
	Category      string `json:"category"`
	ImageFilename string `json:"image_name"`
}

type Items struct {
	Items []*Item `json:"items"`
}

// getItems returns all items in the database
// curl -X GET http://localhost:9000/items
func getItems(c echo.Context) error {
	// get a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer db.Close()

	// invoke SQL to collect all of items
	rows, err := db.Query("SELECT name, category, image_name FROM items")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	// map the items to returned variable
	items := Items{Items: []*Item{}}
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Name, &item.Category, &item.ImageFilename)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		items.Items = append(items.Items, &item)
	}

	// return them as JSON
	return c.JSON(http.StatusOK, items)
}

// addItem adds an item to the database
// curl -X POST -F "name=shoes" http://localhost:9000/items
func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	c.Logger().Infof("Receive item: %s", name)

	message := fmt.Sprintf("item received: %s", name)

	// get a connection to the SQLite3 database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer db.Close()

	// invoke SQL to collect all of items
	stmt, err := db.Prepare("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, "unknown", "default.jpg")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.GET("/items", getItems)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
