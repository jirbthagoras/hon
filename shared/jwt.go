package shared

import (
	"fmt"
	"log/slog"
	"strconv"
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
	_, _, err = ValidateToken(jwtToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Token Invalid")
	}

	return c.Next()
}

func GenerateToken(id int, expiry time.Time) (string, error) {
	// Make a claim to register, using email as a subject cause why not?
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(id),
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

func ValidateToken(tokenStr string) (*jwt.Token, *jwt.RegisteredClaims, error) {
	// Create a instance or new claims to make sure if the parsed claims are type of RegisteredClaims
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for validation
		return getSecretKey(), nil
	})

	// Checks if the token valid.
	if err != nil || !token.Valid {
		return nil, nil, fmt.Errorf("invalid token")
	}

	return token, claims, nil
}

func getTokenFromRequest(c *fiber.Ctx) (string, error) {
	//  get the token
	token := c.Get("Authorization")

	// checks if the token empty or nah
	if token == "" {
		return token, fiber.NewError(fiber.StatusUnauthorized, "Token Not Found")
	}

	return token, nil
}

func GetSubjectFromToken(c *fiber.Ctx) (int, error) {
	// get token with function
	token, err := getTokenFromRequest(c)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	// validate token and get the subject a.k.a email
	_, claims, err := ValidateToken(token)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	// converts from string to int
	id, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, err
	}

	return id, nil
}
