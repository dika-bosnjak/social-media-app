package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/google/uuid"
)

type Friendship struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	UserSentReqID string    `json:"user_sent_req_id"  gorm:"type:varchar(191)"`
	UserGotReqID  string    `json:"user_got_req_id"  gorm:"type:varchar(191)"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type FriendshipAPI struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	UserSentReqID     string    `json:"user_sent_req_id"  gorm:"type:varchar(191)"`
	UserGotReqID      string    `json:"user_got_req_id"  gorm:"type:varchar(191)"`
	UserSentFirstName string    `json:"user_sent_req_firstname"`
	UserSentLastName  string    `json:"user_sent_req_lastname"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type FriendsAPI struct {
	FriendshipID       string `json:"friendship_id"`
	FriendID           string `json:"friend_id"`
	FriendFirstName    string `json:"friend_firstname"`
	FriendLastName     string `json:"friend_lastname"`
	FriendProfileImage string `json:"friend_profile_image"`
}

func CheckFriendship(db *sql.DB, user1ID string, user2ID string) (Friendship, error) {

	//get the friendship info from the database
	var friendship Friendship
	if err := db.QueryRow(`SELECT * 
							FROM friendships 
							WHERE (user_got_req_id = ? AND user_sent_req_id = ?) OR (user_sent_req_id = ? AND user_got_req_id = ?)`, user1ID, user2ID, user1ID, user2ID).
		Scan(
			&friendship.ID,
			&friendship.Status,
			&friendship.CreatedAt,
			&friendship.UpdatedAt,
			&friendship.UserSentReqID,
			&friendship.UserGotReqID); err != nil {
		if err == sql.ErrNoRows {
			return friendship, nil
		}
		return friendship, err
	}
	return friendship, nil
}

func GetFriendshipRequests(db *sql.DB, id string) ([]FriendshipAPI, error) {

	//get the friendship requests
	var friendshipRequests []FriendshipAPI
	rows, err := db.Query(`SELECT friendships.id, user_sent_req_id, user_got_req_id, users.first_name, users.last_name, status, friendships.created_at, friendships.updated_at 
								FROM friendships 
								LEFT JOIN users ON users.id = friendships.user_sent_req_id
								WHERE user_got_req_id = ? AND status = 'pending'`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//loop through the rows of the result and fullfill friendshipRequests slice
	for rows.Next() {
		var friendshipRequest FriendshipAPI
		if err := rows.
			Scan(&friendshipRequest.ID,
				&friendshipRequest.UserSentReqID,
				&friendshipRequest.UserGotReqID,
				&friendshipRequest.UserSentFirstName,
				&friendshipRequest.UserSentLastName,
				&friendshipRequest.Status,
				&friendshipRequest.CreatedAt,
				&friendshipRequest.UpdatedAt); err != nil {
			return friendshipRequests, err
		}
		friendshipRequests = append(friendshipRequests, friendshipRequest)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return friendshipRequests, nil
}

func GetFriends(db *sql.DB, searchUserID string) []FriendsAPI {

	//get all friends from the database
	var friends []FriendsAPI
	rows, err := db.Query(`SELECT friendships.id as FriendshipID, users.id as FriendID,  users.first_name, users.last_name, users.user_photo_url 
							FROM friendships 
							LEFT JOIN users ON users.id = friendships.user_sent_req_id 
							WHERE user_got_req_id = ? AND status = 'accepted'  
								UNION 
							SELECT friendships.id as FriendshipID, users.id as FriendID,  users.first_name as FriendFirstName, users.last_name as FriendLastName, users.user_photo_url as FriendProfileImage 
							FROM friendships 
							LEFT JOIN users ON users.id = friendships.user_got_req_id 
							WHERE user_sent_req_id = ? AND status = 'accepted'`, searchUserID, searchUserID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	//loop through the rows of the result and fullfill friends slice
	for rows.Next() {
		var friend FriendsAPI
		if err := rows.
			Scan(&friend.FriendshipID,
				&friend.FriendID,
				&friend.FriendFirstName,
				&friend.FriendLastName,
				&friend.FriendProfileImage); err != nil {
			return friends
		}
		friends = append(friends, friend)
	}
	if err = rows.Err(); err != nil {
		return nil
	}
	return friends
}

func NumberOfFriends(db *sql.DB, id string) int {

	//get the number of friends for a user
	return len(GetFriends(initializers.DB, id))
}

func GetFriendsIDs(db *sql.DB, searchUserID string) []string {

	//get all friends
	friends := GetFriends(initializers.DB, searchUserID)

	//create a slice with only friends ids
	var friendsIDs []string
	for _, friend := range friends {
		friendsIDs = append(friendsIDs, friend.FriendID)
	}
	return friendsIDs
}

func DeleteFriend(db *sql.DB, loggedInUserID string, userID string) error {

	//check the friendship
	friendship, _ := CheckFriendship(initializers.DB, loggedInUserID, userID)

	//if there is a friendship between the two users, delete it and delete the corresponding chat room
	if friendship.ID != "" {
		_, err := db.Exec(`DELETE 
							FROM friendships 
							WHERE id = ?`, friendship.ID)
		if err == nil {
			room, err := FindChatRoom(db, loggedInUserID, userID)
			if room.ID != "" || err == nil {
				err := DeleteRoom(initializers.DB, loggedInUserID, userID)
				return err
			}
			return err
		}
		return nil
	}
	return nil
}

func AddFriend(db *sql.DB, userID1 string, userID2 string) (Friendship, error) {

	//send the friendship request
	ID := uuid.New().String()
	status := "pending"

	//insert new friend request in the database
	friendship := Friendship{ID: ID, UserSentReqID: userID1, UserGotReqID: userID2, Status: status}
	_, err := db.Exec(`INSERT INTO friendships (id, status, user_sent_req_id, user_got_req_id) 
						VALUES (?, ?, ?, ?)`, ID, status, userID1, userID2)

	return friendship, err
}

func AcceptFriendshipRequest(db *sql.DB, id string) error {

	//update the friendship status to accepted
	sqlStatement := `UPDATE friendships
						SET status = "accepted"
						WHERE id = ?`
	_, err := db.Exec(sqlStatement, id)
	if err != nil {
		return err
	}

	return nil
}

func DeclineFriendshipRequest(db *sql.DB, id string) error {

	//delete the friendship
	_, err := db.Exec(`DELETE 
						FROM friendships 
						WHERE id = ?`, id)
	return err
}

func GetFriendshipUsers(db *sql.DB, id string) (string, string, error) {

	//get the friends from the friendship
	var user1, user2 string
	if err := db.QueryRow(`SELECT user_got_req_id, user_sent_req_id 
							FROM friendships 
							WHERE id = ?`, id).Scan(&user1, &user2); err != nil {
		if err == sql.ErrNoRows {
			return "", "", errors.New("User not found in the database")
		}
		return "", "", err
	}
	return user1, user2, nil
}
