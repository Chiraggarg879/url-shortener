package main

import (
	"log"
	"os"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/chiraggarg879/url-shortener/routes"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

)

func setupRoutes(app *fiber.App, db *sql.DB) {
	app.Get("/:url", func(c *fiber.Ctx) error {
		return routes.ResolveURL(c, db)
	})

	app.Post("/api/v1", func(c *fiber.Ctx) error {
		return routes.ShortenURL(c, db)
	})
}

func initializeconfig()*sql.DB{
	dsn := "appuser:12345678@tcp(127.0.0.1:3306)/URL_SHORTENER"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatal("Error connecting to MySQL:", err)
	}

	fmt.Println("Connected to MySQL successfully!")

	return db
}

func main(){
	err := godotenv.Load()

	if(err != nil){
		fmt.Println(err)
	}
	db := initializeconfig()
	defer db.Close()

	app := fiber.New()

	app.Use(logger.New())
	setupRoutes(app,db)
	log.Fatal(app.Listen(os.Getenv("APP_PORT"))) 	
}