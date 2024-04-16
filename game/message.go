package game

import (
	"encoding/json"
	"log"
	"time"
)

const JoinGameAction = "join-game"
const LeaveGameAction = "leave-game"
const SendMessageAction = "send-message"
const SelectTopicAction = "select-topic"
const StartGameAction = "start-game"
const Error = "error"
const UserJoinedAction = "join-success"
const StartRoundAction = "start-round"
const RoundEndAction = "round-end"
const SendQuestionToUserAction = "send-question-to-user"
const UserStartAnswerAction = "start-answer"
const UserEndAnswerAction = "end-answer"
const RateAction = "rate-user"
const EndRateAction = "rate-end"
const EndGameAction = "game-end"
const StartStageAction = "start-stage"

const UserLeftAction = "user-left"

type Message struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Target  *Game           `json:"target"`
	Sender  *User           `json:"sender"`
	Time    time.Time       `json:"time,omitempty"`
}

type MessageReceive struct {
	Action  string          `json:"action"`
	Message json.RawMessage `json:"message,omitempty"`
	Target  *Game           `json:"target"`
	Sender  string          `json:"sender"`
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		println("123")
		log.Println(err)
	}

	return json
}
