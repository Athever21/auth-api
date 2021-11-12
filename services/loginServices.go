package services

import (
	"auth-api/helpers"
	"auth-api/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/matthewhartstonge/argon2"
)

func LoginUser(e echo.Context) error {
	user_body := make(map[string]string)

	err := json.NewDecoder(e.Request().Body).Decode(&user_body)
	if err != nil {
		return e.JSON(400, &echo.Map{"error": "Invalid body request"})
	}
	if check := CheckBody(user_body); !check {
		return e.JSON(400, &echo.Map{"error": "Missing username or paassword field"})
	}

	user, _ := models.FindByUsername(user_body["username"])

	if user == nil {
		return e.JSON(404, &echo.Map{"error": "User not found"})
	}

	ok, err := argon2.VerifyEncoded([]byte(user_body["password"]), user.Password)

	if err != nil {
		fmt.Println(err)
		return e.JSON(500, &echo.Map{"error": "Internal server error"})
	}
	if !ok {
		return e.JSON(403, &echo.Map{"error": "Invalid password"})
	}

	refresh_cookie := createRefreshCookie(helpers.CreateRefreshToken(user.ID.String()))
	e.SetCookie(refresh_cookie)
	if err != nil {
		return e.JSON(500, &echo.Map{"error": "Internal server error"})
	}
	user.Password = nil
	return e.JSON(200, &echo.Map{"token": helpers.CreateLoginToken(user.ID.String()), "user": user})
}

func RefreshToken(e echo.Context) error {
	refreshCookie, err := e.Cookie("refresh_token")

	if err != nil {
		return e.JSON(403, &echo.Map{"error": "Refresh cookie not present"})
	}

	userId, err := helpers.GetIdFromToken(refreshCookie.Value, true)

	if err != nil {
		return e.JSON(403, &echo.Map{"error": "Invalid token"})
	}

	u := models.FindUserByIdString(userId)

	if u.Username == "" {
		return e.JSON(404, &echo.Map{"error": "User not found"})
	}

	return e.JSON(200, &echo.Map{"token": helpers.CreateLoginToken(userId), "user": u})
}

func createRefreshCookie(value string) *http.Cookie {
	cookie := new(http.Cookie)
	cookie.HttpOnly = true
	cookie.Expires = time.Now().Add(time.Hour * 24 * 7)
	cookie.Name = "refresh_token"
	cookie.Value = value

	return cookie
}
