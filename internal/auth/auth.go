package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)



func HashPassword(password string) (string, error){
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}
func CheckPasswordHash(password, hash string) (bool, error){
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration)(string, error){
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256, 
		jwt.RegisteredClaims{
			Issuer: "chirpy", 
			IssuedAt: jwt.NewNumericDate(time.Now()), 
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)), 
			Subject: userId.String(),
		},
	)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("Error: Trying to sign token.\n%w", err)
	}
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error){
	claims := jwt.MapClaims{}
	token, err  := jwt.ParseWithClaims(
		tokenString, 
		claims, 
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256{
				return nil , jwt.ErrTokenSignatureInvalid
			}
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("Invalid tokenString\n%w", err) 
	}
	userIdString, err := token.Claims.GetSubject()
	if err != nil { 
		return uuid.Nil, fmt.Errorf("userId not found in token string.") 
	}
	userId, err :=  uuid.Parse(userIdString)
	if err != nil { 
		return uuid.Nil, fmt.Errorf("Invalid userId in tokenString.") 
	}
	return userId, nil
}
func GetBearerToken(headers http.Header) (string, error){
	authString := headers.Get("Authorization")
	if authString == ""{
		return "", errors.New("No Authorization header found in request.") 
	}
	authString = strings.TrimPrefix(authString, "Bearer ")
	authString = strings.TrimSpace(authString)
	if authString == ""{
		return "", errors.New("No Authorization token found header.") 
	}
	return authString, nil
}
func MakeRefreshToken() (string, error){
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	tokenString := hex.EncodeToString(tokenBytes)
	return tokenString, nil
}

func GetAPIKey(headers http.Header) (string, error){
	apiKeyString := headers.Get("Authorization")
	if apiKeyString == ""{
		return "", errors.New("No Authorization header found in request.") 
	}
	apiKeyString= strings.TrimPrefix(apiKeyString, "ApiKey ")
	apiKeyString= strings.TrimSpace(apiKeyString)
	if apiKeyString == ""{
		return "", errors.New("No Api Key found Authorization header.") 
	}
	return apiKeyString , nil
}

