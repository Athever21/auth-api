package main

import (
	"auth-api/db"
	"auth-api/routes"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	godotenv.Load()

	app := echo.New()

	routes.UserRouter(app)

	defer db.CloseDb()
	log.Fatal(app.Start(fmt.Sprintf(":%s", os.Getenv("PORT"))))
}
