package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"email-blaze/pkg/domainVerifier"
)

type User struct {
	ID       int
	Email    string
	Domain   string
	Password string
}

func GenerateToken(user *User, secret string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"domain":  user.Domain,
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"exp":     now.Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(secret))
}

func RefreshToken(tokenString string, secret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		now := time.Now()
		claims["iat"] = now.Unix()
		claims["nbf"] = now.Unix()
		claims["exp"] = now.Add(time.Hour * 24).Unix()

		newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return newToken.SignedString([]byte(secret))
	}

	return "", fmt.Errorf("invalid token")
}

func VerifyToken(tokenString string, secret string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("invalid token")
}

func VerifyEmail(email string) (bool, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false, errors.New("invalid email")
	}

	domain := parts[1]
	results, err := domainVerifier.VerifyDomain(domain)
	if err != nil {
		return false, err
	}

	if v, ok := results["MX"]; !ok || v != "Valid" {
		return false, nil
	}

	return true, nil
}

