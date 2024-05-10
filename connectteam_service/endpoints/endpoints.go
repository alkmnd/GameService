package endpoints

const (
	BaseURL              = "http://localhost:8001"
	GetGameURL           = BaseURL + "/api/games/{id}"
	StartGameURL         = BaseURL + "/api/games/start/{id}"
	EndGameURL           = BaseURL + "/api/games/end/{id}"
	SaveResultsURL       = BaseURL + "/api/games/results/{id}"
	GetUserByIdURL       = BaseURL + "/api/users/{id}"
	GetUserActivePlanURL = BaseURL + "/api/users/{id}/plan"
	GetTopicWithIdURL    = BaseURL + "/api/topics/{id}"
	GetRandQuestionsURL  = BaseURL + "/api/questions/"
	GetRandTopicsURL     = BaseURL + "/api/topics/list/{limit}"
)
