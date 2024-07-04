package endpoints

const (
	BaseURL              = "http://80.90.185.79:8001"
	GetGameURL           = BaseURL + "/api/games/{id}"
	StartGameURL         = BaseURL + "/api/games/start/{id}"
	EndGameURL           = BaseURL + "/api/games/end/{id}"
	SaveResultsURL       = BaseURL + "/api/games/{id}/results"
	GetUserByIdURL       = BaseURL + "/api/users/{id}"
	GetUserActivePlanURL = BaseURL + "/api/users/{id}/plan"
	GetTopicWithIdURL    = BaseURL + "/api/topics/{id}"
	GetRandQuestionsURL  = BaseURL + "/api/questions/"
	GetRandTopicsURL     = BaseURL + "/api/topics/list/{limit}"
)
