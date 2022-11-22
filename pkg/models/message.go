package models

import (
	"database/sql"
	"time"
)

type Message struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	RoomID    string    `json:"room_id"  gorm:"type:varchar(191)"`
	Sender    string    `json:"sender_id" gorm:"type:varchar(191)"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func SendMessage(db *sql.DB, id string, roomID string, senderID string, message string, createdAt time.Time) {

	//save the message in the database
	db.Exec(`INSERT INTO messages (id, room_id, sender, message, created_at) 
	VALUES (?, ?, ?, ?, ?)`, id, roomID, senderID, message, createdAt)
}

func GetMessagesByRoomID(db *sql.DB, id string) ([]Message, error) {

	//get the messages by the room id
	var messages []Message
	rows, err := db.Query(`SELECT id, room_id, sender, message, created_at
							FROM messages 
							WHERE room_id = ? 
							ORDER BY created_at ASC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//loop through the rows of the result and fullfill messages slice
	for rows.Next() {
		var message Message
		if err := rows.
			Scan(&message.ID,
				&message.RoomID,
				&message.Sender,
				&message.Message,
				&message.CreatedAt); err != nil {
			return messages, err
		}
		messages = append(messages, message)
	}
	if err = rows.Err(); err != nil {
		return messages, err
	}
	return messages, nil
}
