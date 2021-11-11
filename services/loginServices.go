package services

import (
	"auth-api/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/matthewhartstonge/argon2"
)

type customClaims struct {
	UserId string `json:"asd"`
	jwt.StandardClaims
}

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

	refresh_cookie := createRefreshCookie(createRefreshToken(user.ID.String()))
	e.SetCookie(refresh_cookie)
	if err != nil {
		return e.JSON(500, &echo.Map{"error": "Internal server error"})
	}
	user.Password = nil
	return e.JSON(200, &echo.Map{"token": createLoginToken(user.ID.String()), "user": user})
}

func RefreshToken(e echo.Context) error {
	refreshCookie, err := e.Cookie("refresh_token")

	if err != nil {
		return e.JSON(403, &echo.Map{"error": "Refresh cookie not present"})
	}

	token, err := jwt.ParseWithClaims(
		refreshCookie.Value,
		&customClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("jwt_refresh_token")), nil
		},
	)

	if err != nil || !token.Valid {
		return e.JSON(403, &echo.Map{"error": "Invalid refresh cookie"})
	}

	claims, ok := token.Claims.(*customClaims)

	if !ok {
		return e.JSON(403, &echo.Map{"error": "Couldn't parse claims"})
	}

	userId := getIdFromClaims(claims.UserId)

	u := models.FindUserByIdString(userId)

	if u.Username == "" {
		return e.JSON(404, &echo.Map{"error": "User not found"})
	}

	return e.JSON(200, &echo.Map{"token": createLoginToken(claims.UserId), "user": u})
}

func createRefreshCookie(value string) *http.Cookie {
	cookie := new(http.Cookie)
	cookie.HttpOnly = true
	cookie.Expires = time.Now().Add(time.Hour * 24 * 7)
	cookie.Name = "refresh_token"
	cookie.Value = value

	return cookie
}

func createLoginToken(id string) string {
	signedToken, _ := createToken(id).SignedString([]byte(os.Getenv("jwt_token")))
	return signedToken
}

func createRefreshToken(id string) string {
	signedToken, _ := createToken(id).SignedString([]byte(os.Getenv("jwt_refresh_token")))
	return signedToken
}

func createToken(id string) *jwt.Token {
	claims := customClaims{
		UserId: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: int64(time.Minute * 5),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
}

func getIdFromClaims(id string) string {
	return strings.Split(id, "\"")[1]
}
