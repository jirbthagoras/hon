package shared

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	secretKey []byte
)

func getSecretKey() []byte {
	// taking the cached secret key
	if secretKey != nil {
		return secretKey
	}

	// if cache empty, take it from ENV and convert it to []byte
	secretKey = []byte(config.GetString("JWT_SECRET_KEY"))

	return secretKey
}

func TokenMiddleware(c *fiber.Ctx) error {
	// getting the token
	jwtToken, err := getTokenFromRequest(c)
	if err != nil {
		slog.Error("Error getting token from request", "err", err)
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	// validate the token
	token, err := ValidateToken(jwtToken)
	// because token is *jwt.Token, it can be vaidated with token.Valid
	if err != nil || !token.Valid {
		slog.Error("Error validating token", "err", err)
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	return c.Next()
}

func GenerateToken(email string, expiry time.Time) (string, error) {
	// Make a claim to register, using email as a subject cause why not?
	claims := jwt.RegisteredClaims{
		Subject:   email,
		Issuer:    "Hon",
		ExpiresAt: jwt.NewNumericDate(expiry),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	//  Generate a secret key
	secret := getSecretKey()

	//  Create and sign the token using the secret key
	//  Signing method using HS256 cause why not
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for validation
		return getSecretKey(), nil
	})
}

func getTokenFromRequest(c *fiber.Ctx) (string, error) {
	//  get the token
	token := c.Get("Authorization")

	// checks if the token empty or nah
	if token == "" {
		return token, fiber.NewError(fiber.StatusUnauthorized, "No Token")
	}

	return token, nil
}

func GetSubjectFromToken(c *fiber.Ctx) (string, error) {
	// get token with function
	token, err := getTokenFromRequest(c)
	if err != nil {
		return token, fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	// validate token and get the subject a.k.a email
	claims, err := ValidateToken(token)
	if err != nil {
		return token, fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	// exract the underlzying subject using type assertion
	return claims.Claims.(*jwt.RegisteredClaims).Subject, nil
}
