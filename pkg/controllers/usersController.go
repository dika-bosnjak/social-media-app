package controllers

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Register(c *gin.Context) {

	//Get the data off req body
	var body models.User
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	//Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash the password",
		})
		return
	}

	//add user in the database
	userID := uuid.New().String()
	user, err := models.CreateUser(initializers.DB, userID, body, hash)

	//check for the errors and display them to the user
	if err != nil {
		if strings.Contains(err.Error(), "email") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "There is already a user with that email address",
			})
		} else if strings.Contains(err.Error(), "username") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "There is already a user with that username",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to create a user",
			})
		}
		return
	}
	//Respond
	c.JSON(http.StatusOK, user)
}

func Login(c *gin.Context) {

	//Get the email and pass off req body
	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	//Look up requested user
	user, err := models.GetUserByEmail(initializers.DB, body.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid credientials.",
			})
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	//Compare sent in pass with saved user pass hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid credientials.",
		})
		return
	}

	//Generate a jwt token
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 24 * 30).Unix()

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create a token",
		})
		return
	}

	//Store the value
	user.Password = ""
	c.Set("user", user)

	//Respond
	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": tokenString,
	})
}

func DisplayLoggedInUser(c *gin.Context) {

	loggedInUser, _ := c.Get("user")
	userInfo := models.UserInfo(loggedInUser.(models.User))

	//Respond
	c.JSON(http.StatusOK, userInfo)
}

func UpdateUser(c *gin.Context) {

	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//Get the data off req body
	var body struct {
		Email              string `json:"email"`
		Username           string `json:"username"`
		FirstName          string `json:"first_name"`
		LastName           string `json:"last_name"`
		ProfileType        string `json:"profile_type"`
		ProfileDescription string `json:"profile_description"`
		UserPhotoURL       string `json:"user_photo_url"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	//get user from db
	user, err := models.GetUserByID(initializers.DB, loggedInUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User not found.",
		})
		return
	}

	//update the specific values from the body
	user.Email = body.Email
	user.Username = body.Username
	user.FirstName = body.FirstName
	user.LastName = body.LastName
	user.ProfileType = body.ProfileType
	user.ProfileDescription = body.ProfileDescription
	user.UserPhotoURL = body.UserPhotoURL

	//save the updates
	user, err = models.UpdateUser(initializers.DB, user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to update the user.",
		})
		return
	}

	//Respond
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {

	//get logged in user data
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//delete user in the database
	err := models.DeleteUser(initializers.DB, loggedInUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User could not be deleted." + err.Error(),
		})
		return
	}

	//Respond
	c.JSON(http.StatusOK, gin.H{
		"message": "User is successfully deleted",
	})
}

func SearchUser(c *gin.Context) {

	//get logged in user data
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	//get search param
	searchKeyword := c.Query("search")
	searchKeyword = "%" + searchKeyword + "%"

	var users []models.User
	var err error

	//check the search param
	if searchKeyword == "" {
		//if search param is empty, return all users
		users, err = models.GetUsers(initializers.DB)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	} else {
		//else, return users where username contains the search param
		users, err = models.GetUsersByName(initializers.DB, searchKeyword)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	var usersInfo []models.UserAPI
	//append additional info about each user
	for i := 0; i < len(users); i++ {
		user := users[i]
		user.Password = ""

		blocked, err := models.CheckBlockStatus(initializers.DB, loggedInUserID, user.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if blocked {
			continue
		}

		userInfo := models.UserInfo(users[i])
		usersInfo = append(usersInfo, userInfo)
	}

	//check if there is any user that satisfies the search param
	if len(users) > 0 {
		c.JSON(http.StatusOK, usersInfo)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "There is no user with that username.",
		})
	}
}

func DisplayProfile(c *gin.Context) {

	//get user id from params
	userID := c.Param("id")

	//get logged in user data
	loggedInUser, _ := c.Get("user")
	loggedInUserID := loggedInUser.(models.User).ID

	blocked, err := models.CheckBlockStatus(initializers.DB, loggedInUserID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if blocked {
		c.JSON(http.StatusOK, gin.H{
			"message": "Profile could not be displayed.",
		})
		return
	}

	//find the user in the database
	user, err := models.GetUserByID(initializers.DB, userID)
	user.Password = ""
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User not found.",
		})
		return
	}

	userInfo := models.UserInfo(user)

	//if there is a user with that id, show the user, otherwise show the message
	if user.ID != "" {
		c.JSON(http.StatusOK, userInfo)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "User not found",
		})
	}
}
