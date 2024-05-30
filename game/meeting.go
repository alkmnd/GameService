package game

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type JWTGenerator struct {
	// Api ключ.
	zoomSDKKey string
	// Подпись.
	zoomSDKSecret string
}

// NewJWTGenerator создает экземпляр структуры JWTGenerator.
func NewJWTGenerator(zoomSDKKey string, zoomSDKSecret string) *JWTGenerator {
	return &JWTGenerator{
		zoomSDKKey:    zoomSDKKey,
		zoomSDKSecret: zoomSDKSecret,
	}
}

// GenerateJWTForMeeting генерирует JWT токен для конференции.
func (generator *JWTGenerator) GenerateJWTForMeeting(meetingNumber string, role int) (string, error) {
	// Определение утверждения токена.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"appKey":   generator.zoomSDKKey,
		"sdkKey":   generator.zoomSDKKey,
		"mn":       meetingNumber,
		"role":     role,
		"iat":      time.Now().Unix() - 30,
		"exp":      time.Now().Unix() + 3600,
		"tokenExp": time.Now().Unix() + 7200,
	})

	// Подписание токена.
	tokenString, err := token.SignedString([]byte(generator.zoomSDKSecret))
	return tokenString, err
}
