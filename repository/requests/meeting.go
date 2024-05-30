package requests

import (
	"GameService/repository/endpoints"
	"GameService/repository/models"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"net/http"
	"strconv"
)

type MeetingRepo struct {
	accessToken  string
	refreshToken string
	clientId     string
	clientSecret string
}

func (r *MeetingRepo) CreateMeeting() (meetingNumber string, passcode string, err error) {
	client := resty.New()
	resp, err := client.R().
		SetAuthToken(r.accessToken).
		SetBody(models.CreateMeetingRequest{
			Topic:    "Game meeting",
			Type:     2,
			Settings: models.Settings{},
		}).
		SetPathParam(
			"user_id", "me").Post(endpoints.CreateMeetingURL)
	if err != nil {
		return "", "", err
	}
	if resp.RawResponse.StatusCode == http.StatusUnauthorized {
		err := r.refreshAccessToken()
		if err != nil {
			return "", "", err
		}
		resp, err = client.R().
			SetAuthToken(r.accessToken).
			SetBody(models.CreateMeetingRequest{
				Topic:    "Game meeting",
				Type:     2,
				Settings: models.Settings{},
			}).
			SetPathParam(
				"user_id", "me").Post(endpoints.CreateMeetingURL)
		if err != nil {
			return "", "", err
		}
	}
	var meeting models.CreateMeetingResponse
	err = json.Unmarshal(resp.Body(), &meeting)
	if err != nil {
		return "", "", err
	}

	return strconv.FormatInt(meeting.Id, 10), meeting.Password, err

}

func (r *MeetingRepo) refreshAccessToken() error {
	client := resty.New()
	resp, err := client.R().
		SetBasicAuth(r.clientId, r.clientSecret).
		SetQueryParams(map[string]string{"grant_type": "refresh_token", "refresh_token": r.refreshToken}).
		Post(endpoints.RefreshTokenURL)
	if err != nil {
		return err
	}
	var authData models.AuthData
	err = json.Unmarshal(resp.Body(), &authData)
	if err != nil {
		return err
	}
	r.accessToken = authData.AccessToken
	return nil
}

func NewMeetingRepo(accessToken string,
	refreshToken string,
	clientId string,
	clientSecret string) Meeting {
	return &MeetingRepo{accessToken: accessToken, refreshToken: refreshToken, clientId: clientId, clientSecret: clientSecret}
}
