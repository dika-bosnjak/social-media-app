package controllers

import (
	"net/http"
	"sort"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
)

func AddFriend(c *gin.Context) {
	//get the searched user id from params
	addedUser := c.Param("id")

	//get the logged in user id
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//record the friendship request
	_, err := models.AddFriend(initializers.DB, loggedInUserID, addedUser)

	//respond
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User can not be added as a friend.",
		})
		return
	} else {
		models.SaveNotification(initializers.DB, addedUser, loggedInUserID, "sent you a friend request", "/user/"+loggedInUserID)
		c.JSON(http.StatusOK, gin.H{
			"message": "Friend request is successfully sent.",
		})
		return
	}
}

func DeleteFriend(c *gin.Context) {
	//get the searched user id from params
	userID := c.Param("id")

	//get the logged in user id
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//delete the friend
	err := models.DeleteFriend(initializers.DB, loggedInUserID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Friend is deleted.",
	})
}

func ShowFriendshipRequests(c *gin.Context) {

	//get the logged in user id
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//get all friendship request that user got with status pending
	friendshipRequests, err := models.GetFriendshipRequests(initializers.DB, loggedInUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}

	//respond
	if len(friendshipRequests) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No friendship requests.",
		})
		return
	} else {
		c.JSON(http.StatusOK, friendshipRequests)
		return
	}
}

func AcceptFriendshipRequest(c *gin.Context) {

	//get the friendship req id
	requestID := c.Param("id")

	//accept the friendship request
	err := models.AcceptFriendshipRequest(initializers.DB, requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not find the friendship request",
		})
		return
	}

	//get the ids of new friends
	user1, user2, err := models.GetFriendshipUsers(initializers.DB, requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not find the friendship request",
		})
		return
	}

	//create a new chat room
	_, err = models.CreateRoom(initializers.DB, user1, user2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not create a new room",
		})
		return
	}

	//send the notification
	models.SaveNotification(initializers.DB, user1, user2, "accepted a friend request", "/user/"+user2)

	//respond
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully accepted friendship.",
	})
}

func DeclineFriendshipRequest(c *gin.Context) {

	//get the friendship req id
	requestID := c.Param("id")

	//delete the request from the database
	err := models.DeclineFriendshipRequest(initializers.DB, requestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not delete a friendship request",
		})
		return
	}

	//respond
	c.JSON(http.StatusOK, gin.H{
		"message": "Request is deleted.",
	})
}

func FriendshipStatus(c *gin.Context) {

	//get the searched user id from params
	userID := c.Param("id")

	//get the logged in user id
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//check whether these two users are friends
	friendship, err := models.CheckFriendship(initializers.DB, userID, loggedInUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	//if there is no record, they are not friends
	if friendship.ID == "" {
		c.JSON(http.StatusOK, gin.H{
			"status": "not_friends",
		})
		return
	} else if friendship.ID != "" && friendship.Status == "pending" {
		//when the status is pending, show pending
		if friendship.UserSentReqID == loggedInUserID {
			c.JSON(http.StatusOK, gin.H{
				"status": "pending",
			})
			return
		} else {
			//when the status is pending, but user can accept the req, show pending and accept button
			c.JSON(http.StatusOK, gin.H{
				"status":             "pending",
				"show_accept_button": true,
			})
			return
		}

	} else {
		//if the status is not pending, they are friends
		c.JSON(http.StatusOK, gin.H{
			"status": "accepted",
		})
		return
	}
}

func ShowFriends(c *gin.Context) {

	//get the user id from params
	userID := c.Param("id")

	//get the logged in user id
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	var searchUserID string
	if userID != "" {
		searchUserID = userID
	} else {
		searchUserID = loggedInUserID
	}

	//get the friends of the searched person and blocked users for a logged in user
	friends := models.GetFriends(initializers.DB, searchUserID)
	blockedUsers, err := models.GetBlockedUsersID(initializers.DB, loggedInUserID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not find blocked users",
		})
		return
	}

	//remove the friends (persons) that are blocked/blocked by
	var displayFriends []models.FriendsAPI
	for i := 0; i < len(friends); i++ {
		if !slices.Contains(blockedUsers, friends[i].FriendID) {
			displayFriends = append(displayFriends, friends[i])
		}
	}

	//send json with all frineds (sorted) or display a message
	if len(displayFriends) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No friends yet.",
		})
		return
	} else {
		sort.Slice(displayFriends, func(p, q int) bool {
			return displayFriends[p].FriendFirstName < displayFriends[q].FriendFirstName
		})
		c.JSON(http.StatusOK, displayFriends)
		return
	}
}
