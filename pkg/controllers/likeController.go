package controllers

import (
	"net/http"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/models"
	"github.com/gin-gonic/gin"
)

// LikePost works as a toggle function (toogle like/dislike)
func LikePost(c *gin.Context) {

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

	//get user id from logged in user
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//check whether this user already liked post
	like, err := models.CheckLike(initializers.DB, postID, loggedInUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	if like.ID != "" {
		if like.LikeBool {
			//if there is already a record in the database, switch to dislike
			err := models.UpdateLike(initializers.DB, like, false)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Failed to unlike the post",
				})
				return
			}
			//Respond
			c.JSON(http.StatusOK, gin.H{
				"message": "Successfully disliked post",
			})
			return
		} else {
			//if there is already a record in the database, switch to like
			err := models.UpdateLike(initializers.DB, like, true)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Failed to like the post",
				})
				return
			}
			models.SaveNotification(initializers.DB, post.UserID, loggedInUserID, "liked your post", "/post/"+postID)
			//Respond
			c.JSON(http.StatusOK, gin.H{
				"message": "Successfully liked post",
			})
			return
		}
	} else {
		//if there is no previous record, crate a new one
		_, err := models.LikePost(initializers.DB, loggedInUserID, postID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to like the post",
			})
			return
		}
		models.SaveNotification(initializers.DB, post.UserID, loggedInUserID, "liked your post", "/post/"+postID)

		//Respond
		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully liked post",
		})
		return
	}
}
