package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"  gorm:"type:varchar(191)"`
	SenderID  string    `json:"sender_id"  gorm:"type:varchar(191)"`
	URL       string    `json:"url"`
	Message   string    `json:"message"`
	State     string    `json:"state"`
	ReadDate  time.Time `json:"read_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NotificationAPI struct {
	ID              string    `json:"id"`
	SenderID        string    `json:"sender_id"`
	SenderFirstName string    `json:"sender_first_name"`
	SenderLastName  string    `json:"sender_last_name"`
	URL             string    `json:"url"`
	Message         string    `json:"message"`
	State           string    `json:"state"`
	CreatedAt       time.Time `json:"created_at"`
}

func SaveNotification(db *sql.DB, userID string, senderID string, message string, url string) {

	//if the user that receives the notification is the same as the one who sends, do not save the notification
	if userID == senderID {
		return
	}

	//save the notification in the database
	db.Exec(`INSERT INTO notifications (id, user_id, sender_id, url, message, state) 
				VALUES (?, ?, ?, ?, ?, ?)`, uuid.New().String(), userID, senderID, url, message, "unread")
}

func GetNumberOfNotifications(db *sql.DB, id string) int {

	//get the number of the unread notifications
	var count int
	db.QueryRow(`SELECT COUNT(*) 
					FROM notifications
					WHERE notifications.user_id = ? AND notifications.state = 'unread'`, id).Scan(&count)
	return count
}

func GetNotifications(db *sql.DB, id string) ([]NotificationAPI, error) {

	//get the notifications from the database
	var notifications []NotificationAPI
	rows, err := db.Query(`SELECT notifications.id, notifications.sender_id, users.first_name, users.last_name, notifications.url, notifications.message, notifications.state,  notifications.created_at 
							FROM notifications 
							LEFT JOIN users ON users.id = notifications.sender_id 
							WHERE notifications.user_id = ? 
							ORDER BY notifications.created_at DESC
							LIMIT 20 `, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//loop through the rows of the result and fullfill the notifications slice
	for rows.Next() {
		var notification NotificationAPI
		if err := rows.
			Scan(&notification.ID,
				&notification.SenderID,
				&notification.SenderFirstName,
				&notification.SenderLastName,
				&notification.URL,
				&notification.Message,
				&notification.State,
				&notification.CreatedAt); err != nil {
			return notifications, err
		}
		notifications = append(notifications, notification)
	}
	if err = rows.Err(); err != nil {
		return notifications, err
	}
	return notifications, nil
}

func SetReadNotifications(db *sql.DB, id string) {
	//defer the execution of the function
	time.Sleep(1 * time.Second)

	//set the notifications as read
	sqlStatement := `UPDATE notifications 
						SET state = "read", read_date = ? 
						WHERE user_id= ?`
	db.Exec(sqlStatement, time.Now(), id)

}
