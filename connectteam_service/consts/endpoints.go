package consts

const (
	BaseURL        = "localhost:5432"
	GetGameURL     = BaseURL + "/api/game/{id}"
	StartGameURL   = BaseURL + "/api/game/start/{id}"
	EndGameURL     = BaseURL + "/api/game/end/{id}"
	SaveResultsURL = BaseURL + "/api/game/results/{id}"
)
