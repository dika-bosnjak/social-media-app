package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Like struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"  gorm:"type:varchar(191)"`
	PostID    string    `json:"post_id"  gorm:"type:varchar(191)"`
	LikeBool  bool      `json:"like_bool"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CheckLike(db *sql.DB, postID string, userID string) (Like, error) {

	//check whether there is already a record about the like
	var like Like
	if err := db.QueryRow(`SELECT * 
							FROM likes 
							WHERE post_id = ? AND user_id = ?`, postID, userID).
		Scan(
			&like.ID,
			&like.UserID,
			&like.PostID,
			&like.CreatedAt,
			&like.UpdatedAt,
			&like.LikeBool); err != nil {
		if err == sql.ErrNoRows {
			return like, nil
		}
		return like, err
	}
	return like, nil
}

func CheckIfAlreadyLiked(db *sql.DB, postID string, userID string) bool {

	//check whether the user already liked the post
	var like Like
	if err := db.QueryRow(`SELECT * 
							FROM likes 
							WHERE post_id = ? AND user_id = ? AND like_bool = true`, postID, userID).
		Scan(
			&like.ID,
			&like.UserID,
			&like.PostID,
			&like.CreatedAt,
			&like.UpdatedAt,
			&like.LikeBool); err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		return false
	}
	return true
}

func LikePost(db *sql.DB, loggedInUserID string, postID string) (Like, error) {

	//save the like in the database
	likeID := uuid.New().String()
	like := Like{ID: likeID, UserID: loggedInUserID, PostID: postID, LikeBool: true}
	_, err := db.Exec(`INSERT INTO likes (id, user_id, post_id, like_bool) 
						VALUES (?, ?, ?, ?)`, likeID, loggedInUserID, postID, true)
	return like, err
}

func UpdateLike(db *sql.DB, like Like, likeBool bool) error {

	//like/unlike already created like
	sqlStatement := `UPDATE likes
						SET like_bool = ?
						WHERE id = ?`
	_, err := db.Exec(sqlStatement, likeBool, like.ID)
	if err != nil {
		return err
	}
	return nil
}

func LikeCount(db *sql.DB, id string) int {

	//get the number of the likes on the post
	var count int
	db.QueryRow(`SELECT COUNT(*) 
					FROM likes 
					WHERE post_id = ? AND like_bool = true`, id).Scan(&count)
	return count
}
