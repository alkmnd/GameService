package game

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type JWTGenerator struct {
	zoomSDKKey    string
	zoomSDKSecret string
}

func NewJWTGenerator(zoomSDKKey string, zoomSDKSecret string) *JWTGenerator {
	return &JWTGenerator{
		zoomSDKKey:    zoomSDKKey,
		zoomSDKSecret: zoomSDKSecret,
	}
}

func (generator *JWTGenerator) GenerateJWTForMeeting(meetingNumber string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sdkKey":   generator.zoomSDKSecret,
		"appKey":   generator.zoomSDKSecret,
		"role":     0,
		"mn":       meetingNumber,
		"iat":      time.Now().Unix() - 30,
		"exp":      time.Now().Unix() + 3600,
		"tokenExp": time.Now().Unix() + 7200,
	})

	tokenString, err := token.SignedString([]byte(generator.zoomSDKSecret))
	return tokenString, err
}
