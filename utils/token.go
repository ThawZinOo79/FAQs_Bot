package utils

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func ValidateToken(tokenString string) bool {
	tokenString = extractToken(tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return false
	}

	return true
}

func extractToken(header string) string {
	// support "Bearer <token>"
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return header
}
