package controllers

import (
	"net/http"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/models"
	"github.com/gin-gonic/gin"
)

func ReadNotifications(c *gin.Context) {

	//get logged in user data
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//get the notifications for the logged in user
	notifications, err := models.GetNotifications(initializers.DB, loggedInUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	//Respond
	c.JSON(http.StatusOK, notifications)

	//Set the notifications as read
	models.SetReadNotifications(initializers.DB, loggedInUserID)
}
