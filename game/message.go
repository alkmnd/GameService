package game

import (
	"encoding/json"
	"github.com/google/uuid"
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
const UserStartAnswerAction = "start-answer"
const UserEndAnswerAction = "end-answer"
const RateAction = "rate-user"
const RateEndAction = "rate-end"
const EndGameAction = "game-end"
const StartStageAction = "start-stage"

const UserLeftAction = "user-left"

type Message struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload,omitempty"`
	Target  uuid.UUID   `json:"target"`
	Sender  *User       `json:"sender"`
	Time    time.Time   `json:"time,omitempty"`
}

func (message *Message) encode() []byte {
	messageJson, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return messageJson
}

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}


