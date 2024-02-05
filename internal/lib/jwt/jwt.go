package jwt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"blog-api/internal/domain/models"

	"github.com/go-chi/jwtauth/v5"
	"github.com/golang-jwt/jwt/v5"
)

func NewToken(user models.User, duration time.Duration, secret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func CheckClaim(ctx context.Context, claim, expectedClaim string) (bool, error) {
	const op = "CheckClaim"

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	c := claims[claim]

	switch c.(type) {
	case float64:
		claim, ok := c.(float64)
		if !ok {
			return false, fmt.Errorf("%s: %w", op, err)
		}

		expClaim, err := strconv.ParseFloat(expectedClaim, 64)
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}

		if claim != expClaim {
			return true, fmt.Errorf("%s: %w", op, err)
		}
	case string:
		claim, ok := c.(string)
		if !ok {
			return false, fmt.Errorf("%s: %w", op, err)
		}

		if claim != expectedClaim {
			return false, fmt.Errorf("%s: the claims are not satisfied: %w", err)
		}
	}

	return true, nil
}
