package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/dgrijalva/jwt-go"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
)

type customClaims struct {
	jwt.StandardClaims
	Sub string `json:"sub"`
}

func extractTokenFromHeader(header string) (string, error) {
	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid Authorization header format")
	}
	return parts[1], nil
}

func (c *customClaims) IsExpired() bool {
	return time.Now().Unix() > c.StandardClaims.ExpiresAt
}

func verifyToken(accessToken string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(accessToken, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err != nil {
		slog.Error("Error parsing token", "error", err)
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if token == nil || !token.Valid {
		slog.Warn("Token is nil or invalid")
		return nil, fmt.Errorf("token is not valid")
	}

	slog.Info("Token verified successfully")
	return token, nil
}

func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientToken, err := extractTokenFromHeader(r.Header.Get("Authorization"))
		if err != nil {
			slog.Warn("Invalid authorization header", "error", err)
			http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
			return
		}

		token, err := verifyToken(clientToken)
		if err != nil {
			slog.Warn("Unauthorized with token", "error", err)
			http.Error(w, fmt.Sprintf("Unauthorized with token: %v", err), http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*customClaims)
		if !ok {
			slog.Error("Invalid token claims")
			http.Error(w, "Invalid token claims: could not parse as CustomClaims", http.StatusBadRequest)
			return
		}

		if claims.IsExpired() {
			slog.Warn("Token has expired")
			http.Error(w, "Token has expired", http.StatusUnauthorized)
			return
		}

		userID := claims.Sub
		path := r.URL.Path
		if path == "/refreshToken" {
			slog.Info("Redirecting to refreshToken endpoint.")
			// Placeholder for refreshing logic
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		slog.Debug("User authenticated", "userID", userID)
		next(w, r.WithContext(ctx))
	}
}
