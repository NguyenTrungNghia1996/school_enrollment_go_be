package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"school_enrollment_be/internal/config"
)

type JWTManager interface {
	GenerateToken(subject string, claims jwt.MapClaims) (string, error)
	ParseToken(token string) (*jwt.Token, error)
}

type jwtManager struct {
	secret    []byte
	expiresIn time.Duration
}

func NewJWTManager(cfg config.JWTConfig) JWTManager {
	return &jwtManager{
		secret:    []byte(cfg.Secret),
		expiresIn: cfg.ExpiresIn,
	}
}

func (j *jwtManager) GenerateToken(subject string, claims jwt.MapClaims) (string, error) {
	if claims == nil {
		claims = jwt.MapClaims{}
	}

	now := time.Now()
	claims["sub"] = subject
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(j.expiresIn).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *jwtManager) ParseToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
}
