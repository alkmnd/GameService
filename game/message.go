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
const GameEndedAction = "game-end"
const EndGameAction = "end-game"
const StartStageAction = "start-stage"
const DeleteUserAction = "delete-user"
const UserDeletedAction = "user-deleted"
const UserLeftAction = "user-left"
const GameAbortedAction = "game-abort"

type Message struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload,omitempty"`
	Target  uuid.UUID   `json:"target"`
	Sender  *User       `json:"sender,omitempty"`
	Time    time.Time   `json:"time,omitempty"`
}

func NewMessage(action string,
	payload interface{},
	target uuid.UUID,
	sender *User,
	time time.Time) *Message {
	return &Message{
		Action:  action,
		Payload: payload,
		Target:  target,
		Sender:  sender,
		Time:    time,
	}

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
