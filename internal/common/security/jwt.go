package security

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"school_enrollment_be/internal/config"
)

const (
	TokenTypeAdmin = "admin"
	TokenTypeUser  = "user"
)

type TokenClaims struct {
	ID        int64  `json:"id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type AdminJWTService interface {
	GenerateToken(id int64) (string, error)
	ParseToken(token string) (*TokenClaims, error)
}

type UserJWTService interface {
	GenerateToken(id int64) (string, error)
	ParseToken(token string) (*TokenClaims, error)
}

type jwtService struct {
	secret    []byte
	expiresIn time.Duration
	tokenType string
}

func NewAdminJWTService(cfg config.JWTConfig) AdminJWTService {
	return newJWTService(cfg, TokenTypeAdmin)
}

func NewUserJWTService(cfg config.JWTConfig) UserJWTService {
	return newJWTService(cfg, TokenTypeUser)
}

func newJWTService(cfg config.JWTConfig, tokenType string) *jwtService {
	return &jwtService{
		secret:    []byte(cfg.Secret),
		expiresIn: cfg.ExpiresIn,
		tokenType: tokenType,
	}
}

func (j *jwtService) GenerateToken(id int64) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("id must be greater than 0")
	}

	now := time.Now()
	claims := TokenClaims{
		ID:        id,
		TokenType: j.tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(id, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *jwtService) ParseToken(token string) (*TokenClaims, error) {
	claims := new(TokenClaims)

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.TokenType != j.tokenType {
		return nil, fmt.Errorf("token type %q is not allowed", claims.TokenType)
	}

	if claims.ID <= 0 {
		return nil, fmt.Errorf("invalid token id")
	}

	return claims, nil
}
