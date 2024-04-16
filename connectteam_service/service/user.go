package service

import (
	"GameService/connectteam_service/models"
	"errors"
	"github.com/dgrijalva/jwt-go"
)

type UserService struct{}

func (s *UserService) GetUserById(id int) (user models.User, err error) {
	//TODO implement me
	panic("implement me")
}

const (
	signingKey = "qrkjk#4#%35FSFJlja#4353KSFjH"
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId int    `json:"user_id"`
	Role   string `json:"access"`
}

func (s *UserService) ParseToken(accessToken string) (id int, access string, err error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return 0, "", err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, "", errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserId, claims.Role, nil

}

func NewUserService() User {
	return &UserService{}
}
