package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/jmoiron/sqlx"
)

type Post struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Photo     string    `json:"photo_url"`
	Text      string    `json:"text"`
	UserID    string    `json:"user_id" gorm:"type:varchar(191)"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostAPI struct {
	Post         Post         `json:"post"`
	User         Author       `json:"user"`
	Likes        int          `json:"likes"`
	CommentCount int          `json:"comment_count"`
	Comments     []CommentAPI `json:"comments"`
	IsAuthor     bool         `json:"is_author"`
	AlreadyLiked bool         `json:"already_liked"`
}

type Author struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	UserPhotoURL string `json:"user_photo_url"`
}

func CreatePost(db *sql.DB, post Post) (Post, error) {

	//crate a new post in the database
	_, err := db.Exec(`INSERT INTO posts (id, photo, text, user_id) 
						VALUES (?, ?, ?, ?)`, post.ID, post.Photo, post.Text, post.UserID)
	return post, err

}

func GetPostByID(db *sql.DB, id string) (Post, error) {

	//get the post by id from the database
	var post Post
	if err := db.QueryRow(`SELECT * 
							FROM posts 
							WHERE id = ?`, id).
		Scan(
			&post.ID,
			&post.Photo,
			&post.Text,
			&post.UserID,
			&post.CreatedAt,
			&post.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return post, errors.New("Post not found in the database")
		}
		return post, err
	}
	return post, nil
}

func GetPostAuthor(db *sql.DB, id string) (Author, error) {

	//get the author of the post from the database
	var author Author
	if err := db.QueryRow(`SELECT first_name, last_name, user_photo_url 
							FROM users 
							WHERE id = ?`, id).
		Scan(
			&author.FirstName,
			&author.LastName,
			&author.UserPhotoURL); err != nil {
		if err == sql.ErrNoRows {
			return author, errors.New("Author not found in the database")
		}
		return author, err
	}
	return author, nil
}

func UpdatePost(db *sql.DB, post Post, text string, photo string) (Post, error) {

	//update the post in the database
	sqlStatement := `UPDATE posts 
						SET text = ?, photo = ? 
						WHERE id = ?`
	_, err := db.Exec(sqlStatement, text, photo, post.ID)
	if err != nil {
		return post, err
	}
	post, _ = GetPostByID(initializers.DB, post.ID)
	return post, nil
}

func DeletePost(db *sql.DB, id string) error {

	//delete the post in the database
	_, err := db.Exec(`DELETE 
						FROM posts 
						WHERE id = ?`, id)
	return err
}

func GetPostsByAuthor(db *sql.DB, id string) ([]Post, error) {

	//get all posts from the author(one user)
	var posts []Post
	rows, err := db.Query(`SELECT id, photo, text, user_id, created_at, updated_at 
							FROM posts 
							WHERE user_id = ? 
							ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//loop through the rows of the result and fullfill posts
	for rows.Next() {
		var post Post
		if err := rows.
			Scan(&post.ID,
				&post.Photo,
				&post.Text,
				&post.UserID,
				&post.CreatedAt,
				&post.UpdatedAt); err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return posts, err
	}
	return posts, nil

}

func GetPostsByAuthors(db *sql.DB, ids []string) ([]Post, error) {

	//get posts of all authors
	var posts []Post
	query, args, _ := sqlx.In(`SELECT id, photo, text, user_id, created_at, updated_at 
								FROM posts 
								WHERE user_id IN (?) 
								ORDER BY created_at DESC`, ids)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//loop through the rows of the result and fullfill the posts
	for rows.Next() {
		var post Post
		if err := rows.
			Scan(&post.ID,
				&post.Photo,
				&post.Text,
				&post.UserID,
				&post.CreatedAt,
				&post.UpdatedAt); err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return posts, err
	}

	return posts, nil
}

func GetNumberOfPostsByUserID(db *sql.DB, userID string) int {

	//get the number of the posts written by user id
	var count int
	db.QueryRow(`SELECT COUNT(*) 
					FROM posts 
					WHERE user_id = ?`, userID).Scan(&count)
	return count
}

func EnablePostView(userID string, ownerID string) bool {

	//if the logged in user is the owner, enable view
	if userID == ownerID {
		return true
	}

	//if the profile type of the owner is public, enable view
	profileType, _ := GetUserProfileType(initializers.DB, ownerID)
	if profileType == "public" {
		return true
	}

	//check whether the logged in user and the owner are friends
	friendship, _ := CheckFriendship(initializers.DB, userID, ownerID)
	if friendship.ID != "" && friendship.Status == "accepted" {
		return true
	}

	return false
}

func PostInfo(post Post, loggedInUserID string) PostAPI {
	//get the post author
	author, _ := GetPostAuthor(initializers.DB, post.UserID)

	//get the number of likes
	likeCount := LikeCount(initializers.DB, post.ID)

	//check whether the logged in user already liked the post
	alreadyLiked := CheckIfAlreadyLiked(initializers.DB, post.ID, loggedInUserID)

	//get the number of the comments
	commentCount := GetNumberOfComments(initializers.DB, post.ID)

	//get the comments
	comments := CommentsByPostID(initializers.DB, post.ID, loggedInUserID)

	//append the post info to the list of posts
	postInfo := PostAPI{
		Post:         post,
		User:         author,
		Likes:        likeCount,
		AlreadyLiked: alreadyLiked,
		CommentCount: commentCount,
		Comments:     comments,
	}

	return postInfo
}
