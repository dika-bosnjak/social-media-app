package controllers

import (
	"net/http"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreatePost(c *gin.Context) {

	//Get the data off req body
	var body struct {
		Photo string `json:"photo_url"`
		Text  string `json:"text"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	//Create the post
	postID := uuid.New().String()
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	post := models.Post{ID: postID, UserID: loggedInUserID, Photo: body.Photo, Text: body.Text}
	post, err := models.CreatePost(initializers.DB, post)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create post",
		})
		return
	}

	//Respond
	c.JSON(http.StatusOK, post)
}

func DisplayPost(c *gin.Context) {

	//get postID from request
	postID := c.Param("id")

	//get the post from the database
	post, err := models.GetPostByID(initializers.DB, postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The post is not found.",
		})
		return
	}

	//get the logged in user
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//check whether the logged in user can view post
	enablePostView := models.EnablePostView(loggedInUserID, post.UserID)
	if !enablePostView {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "This account is private.",
		})
		return
	}

	postInfo := models.PostInfo(post, loggedInUserID)
	c.JSON(http.StatusOK, postInfo)
}

func UpdatePost(c *gin.Context) {

	//get postID from request
	postID := c.Param("id")

	//get the post from the database
	post, err := models.GetPostByID(initializers.DB, postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The post is not found.",
		})
		return
	}

	//check whether the logged in user is the author of the post
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID
	if loggedInUserID != post.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	//Get the data off req body
	var body struct {
		Text  string `json:"text"`
		Photo string `json:"photo_url"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	//Update the post
	post, err = models.UpdatePost(initializers.DB, post, body.Text, body.Photo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to update post",
		})
		return
	}

	//Respond
	c.JSON(http.StatusOK, post)
}

func DeletePost(c *gin.Context) {
	//get post id from request
	postID := c.Param("id")

	//get the post from the database
	post, err := models.GetPostByID(initializers.DB, postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to find the post",
		})
		return
	}

	//check whether the logged in user is an author of the post
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID
	if loggedInUserID != post.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		c.Abort()
		return
	}

	//delete the post in the database
	err = models.DeletePost(initializers.DB, post.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to delete the post",
		})
		return
	}

	//Respond
	c.JSON(http.StatusOK, gin.H{
		"message": "Post is successfully deleted",
	})
}

func DisplayPostsByUserId(c *gin.Context) {
	//get userID from request
	userID := c.Param("id")

	//get the logged in user
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//check whether the logged in user can see the posts
	enablePostView := models.EnablePostView(loggedInUserID, userID)
	if !enablePostView {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "This account is private.",
		})
		return
	}

	//find posts in the database
	posts, err := models.GetPostsByAuthor(initializers.DB, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	//check if there is any post from that user yet
	if len(posts) > 0 {
		var postsInfo []models.PostAPI

		//loop through the posts
		for i := 0; i < len(posts); i++ {

			postInfo := models.PostInfo(posts[i], loggedInUserID)
			postsInfo = append(postsInfo, postInfo)
		}

		c.JSON(http.StatusOK, postsInfo)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "No posts yet.",
		})
	}
}

func DisplayPostsOnHomePage(c *gin.Context) {

	//get the logged in user info
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//get friends ids of the logged in user
	friends := models.GetFriendsIDs(initializers.DB, loggedInUserID)

	//find friends posts in the database
	posts, err := models.GetPostsByAuthors(initializers.DB, friends)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	//check if there is any post
	if len(posts) > 0 {

		var postsInfo []models.PostAPI
		//loop through the posts
		for i := 0; i < len(posts); i++ {

			postInfo := models.PostInfo(posts[i], loggedInUserID)
			postsInfo = append(postsInfo, postInfo)
		}
		c.JSON(http.StatusOK, postsInfo)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "No posts yet.",
		})
	}

}
