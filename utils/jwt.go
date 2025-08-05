package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateToken(userID uint) (string, string, error) {
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	at, err := accessToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	rt, err := refreshToken.SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET")))
	if err != nil {
		return "", "", err
	}

	return at, rt, nil
}
