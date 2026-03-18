package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createJWT(userID string) (string, error) {
	signingKey := []byte(os.Getenv("SIGNING_KEY"))
	now := time.Now()

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": "auth-service",                 // Issuer
			"sub": userID,                         // Subject
			"iat": now.Unix(),                     // Issued at
			"exp": now.Add(24 * time.Hour).Unix(), // Expires at
		},
	)

	signedToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	return signedToken, nil
}

func validateJWT(t string) (string, error) {
	signingKey := []byte(os.Getenv("SIGNING_KEY"))

	parsedToken, err := jwt.ParseWithClaims(t, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return signingKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", status.Error(codes.Unauthenticated, "token expired")
		}
		return "", status.Error(codes.Unauthenticated, "unauthenticated")
	}

	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", status.Error(codes.Internal, "claims type assertion failed")
	}

	return claims.Subject, nil
}
