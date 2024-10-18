package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	user := &User{
		ID:     1,
		Email:  "test@example.com",
		Domain: "example.com",
	}
	secret := "test_secret"

	token, err := GenerateToken(user, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the generated token
	claims, err := VerifyToken(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, float64(user.ID), (*claims)["user_id"])
	assert.Equal(t, user.Email, (*claims)["email"])
	assert.Equal(t, user.Domain, (*claims)["domain"])
}

func TestRefreshToken(t *testing.T) {
	user := &User{
		ID:     1,
		Email:  "test@example.com",
		Domain: "example.com",
	}
	secret := "test_secret"

	// Generate an initial token
	initialToken, _ := GenerateToken(user, secret)

	// Refresh the token
	refreshedToken, err := RefreshToken(initialToken, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshedToken)
	assert.NotEqual(t, initialToken, refreshedToken)

	// Verify the refreshed token
	claims, err := VerifyToken(refreshedToken, secret)
	assert.NoError(t, err)
	assert.Equal(t, float64(user.ID), (*claims)["user_id"])
	assert.Equal(t, user.Email, (*claims)["email"])
	assert.Equal(t, user.Domain, (*claims)["domain"])
}

func TestVerifyToken(t *testing.T) {
	user := &User{
		ID:     1,
		Email:  "test@example.com",
		Domain: "example.com",
	}
	secret := "test_secret"

	// Generate a valid token
	validToken, _ := GenerateToken(user, secret)

	// Test with valid token
	claims, err := VerifyToken(validToken, secret)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// Test with invalid token
	invalidToken := "invalid.token.string"
	claims, err = VerifyToken(invalidToken, secret)
	assert.Error(t, err)
	assert.Nil(t, claims)

	// Test with expired token
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	expiredTokenString, _ := expiredToken.SignedString([]byte(secret))
	claims, err = VerifyToken(expiredTokenString, secret)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestVerifyEmail(t *testing.T) {
	// Test cases
	testCases := []struct {
		email    string
		expected bool
		hasError bool
	}{
		{"valid@example.com", true, false},
		{"invalid@example", false, true},
		{"notanemail", false, true},
	}

	for _, tc := range testCases {
		result, err := VerifyEmail(tc.email)
		if tc.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, tc.expected, result)
	}
}
