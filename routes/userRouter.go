package routes

import (
	"auth-api/services"

	"github.com/labstack/echo/v4"
)

func UserRouter(app *echo.Echo) {
	users := app.Group("/users")
	users.GET("", services.GetAllUsers)
	users.POST("", services.CreateUser)

	users.POST("/login", services.LoginUser)
	users.POST("/refresh_token", services.RefreshToken)

	usersId := app.Group("/users/:id")
	usersId.Use(services.UsersMiddleware)
	usersId.GET("", services.GetUser)
}
