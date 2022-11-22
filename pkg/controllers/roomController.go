package controllers

import (
	"net/http"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/models"
	"github.com/gin-gonic/gin"
)

func OpenChatRoom(c *gin.Context) {

	//get the logged in user info
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//get the user from the params
	userID := c.Param("userID")

	//get the room
	room, err := models.FindChatRoom(initializers.DB, loggedInUserID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to find the chat room.",
		})
		return
	}

	//get the messages
	messages, err := models.GetMessagesByRoomID(initializers.DB, room.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to find the messages.",
		})
		return
	}

	//Respond
	c.JSON(http.StatusOK, gin.H{
		"room_id":  room.ID,
		"messages": messages,
	})
}
