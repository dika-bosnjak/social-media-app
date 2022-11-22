package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
)

type User struct {
	ID                 string    `json:"id" gorm:"primaryKey"`
	Email              string    `json:"email" gorm:"unique"`
	Password           string    `json:"password,omitempty"`
	Username           string    `json:"username" gorm:"unique"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	ProfileDescription string    `json:"profile_description"`
	UserPhotoURL       string    `json:"user_photo_url"`
	ProfileType        string    `json:"profile_type"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type UserAPI struct {
	User                 User `json:"user"`
	NumberOfFriends      int  `json:"number_of_friends"`
	NumberOfPosts        int  `json:"number_of_posts"`
	NumberOfBlockedUsers int  `json:"number_of_blocked_users"`
}

func CreateUser(db *sql.DB, userID string, body User, hash []byte) (User, error) {

	//Create the user
	user := User{
		ID:          userID,
		Email:       body.Email,
		Password:    string(hash),
		Username:    body.Username,
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		ProfileType: "public"}

	//insert the user in the database
	_, err := db.Exec(`INSERT INTO users (id, email, password, username, first_name, last_name, profile_type, profile_description, user_photo_url) 
						VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		user.Email,
		user.Password,
		user.Username,
		user.FirstName,
		user.LastName,
		user.ProfileType,
		"",
		"")

	return user, err
}

func GetUserByEmail(db *sql.DB, email string) (*User, error) {

	//get the user by email from the database
	var user User
	if err := db.
		QueryRow(`SELECT id, email, password, username, first_name, last_name, profile_description, user_photo_url, profile_type 
					FROM users 
					WHERE email = ?`, email).
		Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.ProfileDescription,
			&user.UserPhotoURL,
			&user.ProfileType,
		); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("User not found in the database")
		}
		return nil, err
	}
	return &user, nil
}

func GetUserByID(db *sql.DB, id string) (User, error) {

	//get the user by id from the database
	var user User
	if err := db.QueryRow(`SELECT id, email, password, username, first_name, last_name, profile_description, user_photo_url, profile_type
							FROM users 
							WHERE id = ?`, id).
		Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.ProfileDescription,
			&user.UserPhotoURL,
			&user.ProfileType); err != nil {
		if err == sql.ErrNoRows {
			return user, errors.New("User not found in the database")
		}
		return user, err
	}

	return user, nil
}

func GetUserProfileType(db *sql.DB, id string) (string, error) {

	//get the user profile type from the database
	var profileType string
	if err := db.QueryRow(`SELECT profile_type 
							FROM users 
							WHERE id = ?`, id).
		Scan(&profileType); err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("User not found in the database")
		}
		return "", err
	}
	return profileType, nil
}

func UpdateUser(db *sql.DB, user User) (User, error) {

	//update the user in the database
	sqlStatement := `UPDATE users 
						SET email = ?, username = ?, first_name = ?, last_name = ?, profile_description = ?, user_photo_url = ?, profile_type = ? 
						WHERE id = ?`
	_, err := db.Exec(sqlStatement, user.Email, user.Username, user.FirstName, user.LastName, user.ProfileDescription, user.UserPhotoURL, user.ProfileType, user.ID)
	if err != nil {
		return user, err
	}

	user.Password = ""
	return user, nil
}

func DeleteUser(db *sql.DB, id string) error {

	//delete the user in the database
	_, err := db.Exec(`DELETE 
						FROM users 
						WHERE id = ?`, id)
	return err
}

func GetUsers(db *sql.DB) ([]User, error) {

	//get the users from the database
	var users []User
	rows, err := db.Query(`SELECT id, email, username, first_name, last_name, profile_description, user_photo_url, profile_type 
							FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//loop through the rows in the result and fullfill the users slice
	for rows.Next() {
		var user User
		if err := rows.
			Scan(&user.ID,
				&user.Email,
				&user.Username,
				&user.FirstName,
				&user.LastName,
				&user.ProfileDescription,
				&user.UserPhotoURL,
				&user.ProfileType); err != nil {
			return users, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return users, err
	}

	return users, nil
}

func GetUsersByName(db *sql.DB, name string) ([]User, error) {

	//serach for the users by name
	var users []User
	rows, err := db.Query(`SELECT id, email, username, first_name, last_name, profile_description, user_photo_url, profile_type 
							FROM users 
							WHERE first_name LIKE ? OR last_name LIKE ?`, name, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//loop through the rows in the result and fullfill the users slice
	for rows.Next() {
		var user User
		if err := rows.
			Scan(&user.ID,
				&user.Email,
				&user.Username,
				&user.FirstName,
				&user.LastName,
				&user.ProfileDescription,
				&user.UserPhotoURL,
				&user.ProfileType); err != nil {
			return users, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return users, err
	}

	return users, nil
}

func UserInfo(user User) UserAPI {
	//find the number of posts
	numberOfPosts := GetNumberOfPostsByUserID(initializers.DB, user.ID)

	//find the number of friends
	numberOfFriends := NumberOfFriends(initializers.DB, user.ID)

	//find the number of blocked users
	numberOfBlockedUsers := NumberOfBlockedUsers(initializers.DB, user.ID)

	userInfo := UserAPI{
		User:                 user,
		NumberOfPosts:        numberOfPosts,
		NumberOfFriends:      numberOfFriends,
		NumberOfBlockedUsers: numberOfBlockedUsers,
	}
	return userInfo
}
