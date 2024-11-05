package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	middleware "github.com/hemantsharma1498/segwise-assignment/pkg/auth-middleware"
)

// GenerateJWT generates a new JWT token for the authenticated user
func GenerateJWT(userID int, email string) (string, error) {
	// Set token expiration time
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create the claims with user details and expiration time
	claims := &middleware.Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create the token with the claims and sign it using the secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(middleware.JwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
