package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Note - if you're concerned about how strings and temporary byte slices hang
// around in memory until garbage collected - IMO there seems to be little
// practical benefit to making code more error prone and obscure by trying to
// deal with everything as byte slices that you then zero-out after use, when
// golang and the http library deal with everything as strings anyway...

func HashPassword(password string) (string, error) {
	// Use DefaultCost (passing '0' to get this)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	return string(hash), err
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject: userID.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	// This API is a little weird, and the documentation is pretty awful.
	// It looks like you have to pass in a stack-local 'claims' to parse into ...
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	// ... but then you can still access it via the returned token
	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	if id == "" {
		return uuid.UUID{}, errors.New("subject claim is missing")
	}

	return uuid.Parse(id)
}

var bearerRegex = regexp.MustCompile(`^Bearer\s+([A-Za-z0-9-._~+/]+=*)$`)
func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", errors.New("Missing or empty Authorization header")
	}
	
	matches := bearerRegex.FindStringSubmatch(auth)
	if len(matches) < 2 {
		// Don't leak the actual token in logs
		return "", errors.New("Malformed Authorization header")
	}

	return matches[1], nil
}

func MakeRefreshToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Missing Authorization header")
	}
	const prefix = "ApiKey "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("Authorization header must start with 'ApiKey '")
	}
	apiKey := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	if apiKey == "" {
		return "", errors.New("API key is empty")
	}
	return apiKey, nil
}