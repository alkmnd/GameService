package requests

import (
	"GameService/connectteam_service/endpoints"
	"GameService/connectteam_service/models"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type UserService struct{ apiKey string }

func (s *UserService) GetCreatorPlan(id uuid.UUID) (plan models.UserPlan, err error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("X-API-Key", s.apiKey).SetPathParam("id", id.String()).Get(endpoints.GetUserActivePlanURL)
	err = json.Unmarshal(resp.Body(), &plan)
	if err != nil {
		return plan, err
	}

	return plan, err
}

func (s *UserService) GetUserById(id uuid.UUID) (user models.User, err error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("X-API-Key", s.apiKey).SetPathParam("id", id.String()).Get(endpoints.GetUserByIdURL)
	if err != nil {
		return user, err
	}

	err = json.Unmarshal(resp.Body(), &user)
	if err != nil {
		return user, err
	}

	return user, err

}

const (
	signingKey = "qrkjk#4#%35FSFJlja#4353KSFjH"
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId uuid.UUID `json:"user_id"`
	Role   string    `json:"access"`
}

func (s *UserService) ParseToken(accessToken string) (id uuid.UUID, access string, err error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return uuid.Nil, "", err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return uuid.Nil, "", errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserId, claims.Role, nil

}

func NewUserService(apiKey string) *UserService {
	return &UserService{apiKey: apiKey}
}
