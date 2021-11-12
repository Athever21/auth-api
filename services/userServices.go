package services

import (
	"auth-api/helpers"
	"auth-api/models"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/matthewhartstonge/argon2"
)

func GetAllUsers(e echo.Context) error {
	page := e.QueryParam("page")
	n, _ := strconv.Atoi(page)
	users, err := models.GetAllUser(n)
	if err != nil {
		fmt.Println(err)
		return e.JSON(500, &echo.Map{"error": "Internal Server Error"})
	}

	return e.JSON(200, users)
}

func CheckBody(body map[string]string) bool {
	_, ok := body["username"]
	if !ok {
		return false
	}
	_, ok = body["password"]
	return ok
}

func CreateUser(e echo.Context) error {
	user_body := make(map[string]string)
	argon := argon2.DefaultConfig()

	err := json.NewDecoder(e.Request().Body).Decode(&user_body)
	if err != nil {
		return e.JSON(400, &echo.Map{"error": "Invalid body request"})
	}
	if check := CheckBody(user_body); !check {
		return e.JSON(400, &echo.Map{"error": "Missing username or paassword field"})
	}

	withUsername, _ := models.FindByUsername(user_body["username"])

	if withUsername != nil {
		return e.JSON(400, &echo.Map{"error": "Username already in use"})
	}

	hash, err := argon.HashEncoded([]byte(user_body["password"]))

	if err != nil {
		return e.JSON(500, &echo.Map{"error": "Internal server error"})
	}

	user := models.User{Username: user_body["username"], Password: hash, CreatedAt: time.Now()}

	u, err := models.CreateUser(user)

	if err != nil {
		return e.JSON(500, &echo.Map{"error": "Internal server error"})
	}

	return e.JSON(200, u)
}

func UsersMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(e echo.Context) error {
		id := e.Param("id")
		user := models.FindUserByIdString(id)

		if user.Username == "" {
			return e.JSON(404, &echo.Map{"error": "User not found"})
		}

		e.Set("userFromId", user)
		if e.Request().Method == "GET" {
			return next(e)
		}

		auth := e.Request().Header["Authorization"]

		if len(auth) == 0 {
			return e.JSON(401, &echo.Map{"error": "Unauthorized"})
		}

		if valid := strings.HasPrefix(auth[0], "Bearer "); !valid {
			return e.JSON(401, &echo.Map{"error": "Unauthorized"})
		}

		userId, err := helpers.GetIdFromToken(auth[0][7:], false)

		if err != nil {
			return e.JSON(401, &echo.Map{"error": "Invalid token"})
		}

		u := models.FindUserByIdString(userId)

		if u.Username == "" {
			return e.JSON(404, &echo.Map{"error": "User Not Found"})
		}

		if u.ID != user.ID {
			return e.JSON(403, &echo.Map{"error": "Forbidden"})
		}

		e.Set("authUser", u)

		return next(e)
	}
}

func GetUser(e echo.Context) error {
	return e.JSON(200, e.Get("userFromId"))
}

func ChangeUser(e echo.Context) error {
	userAuth := e.Get("authUser").(*models.User)

	var body map[string]string

	err := json.NewDecoder(e.Request().Body).Decode(&body)
	if err != nil {
		return e.JSON(400, &echo.Map{"error": "Invalid body request"})
	}

	user, err := models.UpdateUser(userAuth, body)

	if err != nil {
		return e.JSON(400, &echo.Map{"error": err.Error()})
	}

	user.Password = nil
	return e.JSON(200, user)
}

func DeleteUser(e echo.Context) error {
	userAuth := e.Get("authUser").(*models.User)
	err := models.DeleteUser(userAuth)

	if err != nil {
		return e.JSON(400, &echo.Map{"error": "Internal server error"})
	}

	return e.JSON(400, &echo.Map{"success": "true"})
}
