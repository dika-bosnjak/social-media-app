package websocketrooms

import (
	"encoding/json"
	"log"
	"time"
)

const SendMessageAction = "send-message"
const JoinRoomPrivateAction = "join-room"
const LeaveRoomAction = "leave-room"

type Message struct {
	Action  string    `json:"action"`
	Message string    `json:"message"`
	Target  string    `json:"target"`
	Sender  *Client   `json:"sender"`
	Time    time.Time `json:"created_at"`
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}
