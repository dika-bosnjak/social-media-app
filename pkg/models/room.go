package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type Room struct {
	ID      string `json:"id" gorm:"primaryKey"`
	User1ID string `json:"user1_id" gorm:"type:varchar(191)"`
	User2ID string `json:"user2_id" gorm:"type:varchar(191)"`
}

func FindChatRoom(db *sql.DB, userID1 string, userID2 string) (Room, error) {

	//find the chat room for the users
	var room Room
	if err := db.QueryRow(`SELECT * 
							FROM rooms 
							WHERE (user1_id = ? AND user2_id = ?) OR (user2_id = ? AND user1_id = ?)`, userID1, userID2, userID1, userID2).
		Scan(
			&room.ID,
			&room.User1ID,
			&room.User2ID); err != nil {
		if err == sql.ErrNoRows {
			return room, err
		}
		return room, err
	}
	return room, nil
}

func CreateRoom(db *sql.DB, userID1 string, userID2 string) (Room, error) {

	//crate new chat room for two users
	roomID := uuid.New().String()
	room := Room{ID: roomID, User1ID: userID1, User2ID: userID2}
	_, err := db.Exec(`INSERT INTO rooms (id, user1_id, user2_id) 
						VALUES (?, ?, ?)`, roomID, userID1, userID2)
	return room, err
}

func DeleteRoom(db *sql.DB, userID1 string, userID2 string) error {

	//delete the chat room
	room, _ := FindChatRoom(db, userID1, userID2)
	_, err := db.Exec(`DELETE rooms 
						WHERE id = ?`, room.ID)
	return err
}

func FindChatRoomUsers(db *sql.DB, roomID string) (string, string) {

	//find the users from the room
	var user1, user2 string
	if err := db.QueryRow(`SELECT user1_id, user2_id 
							FROM rooms 
							WHERE id = ?)`, roomID).Scan(&user1, &user2); err != nil {
		return "", ""
	}
	return user1, user2
}
