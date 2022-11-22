package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"  gorm:"type:varchar(191)"`
	PostID    string    `json:"post_id"  gorm:"type:varchar(191)"`
	Text      string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CommentAPI struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	UserPhotoURL string    `json:"user_photo_url"`
	PostID       string    `json:"post_id"`
	PostOwner    string    `json:"post_owner"`
	Text         string    `json:"comment"`
	EnableDelete bool      `json:"enable_delete"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func CreateComment(db *sql.DB, text string, userID string, postID string) (Comment, error) {

	//create a new comment in the database
	commentID := uuid.New().String()
	comment := Comment{ID: commentID, UserID: userID, PostID: postID, Text: text}
	_, err := db.Exec(`INSERT INTO comments (id, user_id, post_id, text) 
						VALUES (?, ?, ?, ?)`, commentID, userID, postID, text)
	return comment, err
}

func GetNumberOfComments(db *sql.DB, postID string) int {

	//get the number of the comments for a specific post
	var count int
	db.QueryRow(`SELECT COUNT(*) 
					FROM comments 
					WHERE post_id = ?`, postID).Scan(&count)
	return count

}

func CommentsByPostID(db *sql.DB, postID string, userID string) []CommentAPI {

	//get comments for the post
	var comments []CommentAPI
	rows, err := db.Query(`SELECT comments.id, comments.user_id, users.first_name, users.last_name, users.user_photo_url, comments.post_id, posts.user_id, comments.text, comments.created_at, comments.updated_at  
							FROM comments 
							LEFT JOIN users ON users.id = comments.user_id 
							LEFT JOIN posts ON comments.post_id = posts.id 
							WHERE comments.post_id = ? 
							ORDER BY comments.created_at asc`, postID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	//loop through the rows of the result and fullfill comments slice
	for rows.Next() {
		var comment CommentAPI
		if err := rows.
			Scan(&comment.ID,
				&comment.UserID,
				&comment.FirstName,
				&comment.LastName,
				&comment.UserPhotoURL,
				&comment.PostID,
				&comment.PostOwner,
				&comment.Text,
				&comment.CreatedAt,
				&comment.UpdatedAt); err != nil {
			return comments
		}
		comments = append(comments, comment)
	}
	if err = rows.Err(); err != nil {
		return comments
	}

	//loop through all comments and add enable option for the users with the permission to delete comment
	for i := 0; i < len(comments); i++ {
		if comments[i].UserID == userID || comments[i].PostOwner == userID {
			comments[i].EnableDelete = true
		}
	}

	return comments
}

func EnableDeleteComment(db *sql.DB, userID string, commentID string) bool {

	//check if the logged in user is comment creator
	var commentByUser Comment
	db.QueryRow(`SELECT * 
					FROM comments 
					WHERE id = ? AND user_id = ?`, commentID, userID).
		Scan(
			&commentByUser.ID,
			&commentByUser.UserID,
			&commentByUser.PostID,
			&commentByUser.Text,
			&commentByUser.CreatedAt,
			&commentByUser.UpdatedAt)

	//check if the logged in user is a post creator
	var commentByOwner Comment
	db.QueryRow(`SELECT c.id, c.user_id, c.post_id, c.text, c.created_at, c.updated_at
					FROM comments c
					LEFT JOIN posts ON posts.id = c.post_id 
					WHERE c.id = ? AND posts.user_id = ?`, commentID, userID).
		Scan(
			&commentByOwner.ID,
			&commentByOwner.UserID,
			&commentByOwner.PostID,
			&commentByOwner.Text,
			&commentByOwner.CreatedAt,
			&commentByOwner.UpdatedAt)

	//if the logged in user did not write the post nor the comment, disable delete option
	if commentByUser.ID != "" || commentByOwner.ID != "" {
		return true
	} else {
		//otherwise, enable delete option
		return false
	}
}

func DeleteComment(db *sql.DB, id string) error {

	//delete the comment in the database
	_, err := db.Exec(`DELETE 
						FROM comments
						WHERE id = ?`, id)
	return err
}
