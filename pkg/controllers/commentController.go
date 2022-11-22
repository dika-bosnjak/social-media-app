package controllers

import (
	"net/http"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/models"
	"github.com/gin-gonic/gin"
)

func AddComment(c *gin.Context) {

	//get post id from params
	postID := c.Param("id")

	//check whether this post exists
	post, err := models.GetPostByID(initializers.DB, postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Post does not exist",
		})
		return
	}

	//get logged in user
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//Get the data off req body
	var body struct {
		CommentText string `json:"comment"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	//save the comment in the database
	comment, err := models.CreateComment(initializers.DB, body.CommentText, loggedInUserID, post.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to comment the post",
		})
		return
	}

	//send the notification
	models.SaveNotification(initializers.DB, post.UserID, loggedInUserID, "commented your post", "/post/"+postID)

	//Respond
	c.JSON(http.StatusOK, comment)
}

func DeleteComment(c *gin.Context) {

	//get comment id
	commentID := c.Param("id")

	//get logged in user
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//check if the logged in user can delete a comment
	enableDelete := models.EnableDeleteComment(initializers.DB, loggedInUserID, commentID)

	//if logged in user can delete a comment, delete it
	if enableDelete {
		err := models.DeleteComment(initializers.DB, commentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Comment could not be deleted.",
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Comment deleted",
			})
			return
		}
	} else {
		//if a user does not have a permission to delete, display the error message
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized.",
		})
		return
	}
}
