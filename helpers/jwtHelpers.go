package helpers

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type customClaims struct {
	UserId string
	jwt.StandardClaims
}

func CreateLoginToken(id string) string {
	signedToken, _ := createToken(id, false).SignedString([]byte(os.Getenv("jwt_token")))
	return signedToken
}

func CreateRefreshToken(id string) string {
	signedToken, _ := createToken(id, true).SignedString([]byte(os.Getenv("jwt_refresh_token")))
	return signedToken
}

func createToken(id string, refresh bool) *jwt.Token {
	claims := customClaims{
		UserId: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: int64(time.Now().UTC().Unix() + 60*5),
			IssuedAt:  time.Now().UTC().Unix(),
		},
	}

	if refresh {
		claims.StandardClaims.ExpiresAt = int64(time.Now().UTC().Unix() + 60*60*24*7)
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
}

func GetIdFromToken(tokenString string, refresh bool) (string, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&customClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if refresh {
				return []byte(os.Getenv("jwt_refresh_token")), nil
			}
			return []byte(os.Getenv("jwt_token")), nil
		},
	)

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims := token.Claims.(*customClaims)
	return getIdFromClaims(claims.UserId), nil
}

func getIdFromClaims(id string) string {
	arr := strings.Split(id, "\"")

	if len(arr) == 3 {
		return arr[1]
	}

	return arr[0]
}
