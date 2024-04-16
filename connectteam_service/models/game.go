package models

type Game struct {
	Id        int    `json:"id"`
	Status    string `json:"status"`
	Name      string `json:"name"`
	CreatorId int    `json:"creator_id"`
}
