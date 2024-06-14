package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)

var SecretKey string

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	SecretKey = os.Getenv("JWT_SECRET_KEY")
}

func generateToken(userID string, duration time.Duration) (string, error) {
	if userID == "" {
		return "", fmt.Errorf("userID cannot be empty")
	}
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(duration).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func GenerateAccessToken(userID string) (string, error) {
	return generateToken(userID, time.Hour)
}

func generateRefreshToken(userID string) (string, error) {
	return generateToken(userID, 6*30*24*time.Hour) // 6 months beacusae redundency.
}

func GenerateTokens(userID string) (string, string, error) {
	accessToken, err := GenerateAccessToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("error generating access token: %w", err)
	}
	refreshToken, err := generateRefreshToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("error generating refresh token: %w", err)
	}
	return accessToken, refreshToken, nil
}
