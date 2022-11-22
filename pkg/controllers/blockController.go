package controllers

import (
	"net/http"
	"sort"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func BlockUser(c *gin.Context) {

	//get the searched user id from params
	blockUser := c.Param("id")

	//get the logged in user id
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//check whether these two users are friends
	var friendshipRequest models.Friendship
	friendshipRequest, err := models.CheckFriendship(initializers.DB, loggedInUserID, blockUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Friendship could not be found",
		})
		return
	}

	//if they are friends, delete the friendship
	if friendshipRequest.ID != "" {
		models.DeleteFriend(initializers.DB, loggedInUserID, blockUser)
	}

	//block the user
	ID := uuid.New().String()
	models.BlockUser(initializers.DB, ID, loggedInUserID, blockUser)

	c.JSON(http.StatusOK, gin.H{
		"message": "User is successfully blocked.",
	})
}

func UnblockUser(c *gin.Context) {

	//get the searched user id from params
	userID := c.Param("id")

	//get the logged in user id
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//get the block, and then unblock the person
	block, err := models.GetBlock(initializers.DB, loggedInUserID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}
	if err := models.Unblock(initializers.DB, block.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User is unblocked.",
	})

}

func ShowBlockedUsers(c *gin.Context) {
	//get the logged in user id
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	blockedUsers, err := models.GetBlockedUsers(initializers.DB, loggedInUserID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	//send json with all blocked users (sorted) or display a message
	if len(blockedUsers) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No blocked users yet.",
		})
		return
	} else {
		sort.Slice(blockedUsers, func(p, q int) bool {
			return blockedUsers[p].UserBlockedFirstName < blockedUsers[q].UserBlockedFirstName
		})

		c.JSON(http.StatusOK, blockedUsers)
		return
	}
}
