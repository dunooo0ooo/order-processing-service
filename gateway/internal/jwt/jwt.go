package jwtutil

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func Sign(secret, role string, ttl time.Duration) (string, error) {
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func Parse(secret, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
