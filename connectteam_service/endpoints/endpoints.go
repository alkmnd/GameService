package endpoints

const (
	BaseURL              = "localhost:8000"
	GetGameURL           = BaseURL + "/api/game/{id}"
	StartGameURL         = BaseURL + "/api/game/start/{id}"
	EndGameURL           = BaseURL + "/api/game/end/{id}"
	SaveResultsURL       = BaseURL + "/api/game/results/{id}"
	GetUserByIdURL       = BaseURL + "/api/users/{id}"
	GetUserActivePlanURL = BaseURL + "/api/users/{id}/plan"
	GetTopicWithIdURL    = BaseURL + "/api/topics/{id}"
	GetRandQuestionsURL  = BaseURL + "/api/questions/{id}/{limit}"
	GetRandTopicsURL     = BaseURL + "/api/topics/{limit}"
	GetQuestionTags      = BaseURL + "/api/questions/{id}/tags"
)
