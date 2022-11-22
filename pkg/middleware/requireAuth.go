package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type MyCustomClaims struct {
	UserID string `json:"sub"`
	jwt.StandardClaims
}

func RequireAuth(c *gin.Context) {
	var tokenString string
	//Get the token off req
	if c.Request.Header.Get("Authorization") != "" {
		tokenString = strings.Split(c.Request.Header.Get("Authorization"), " ")[1]
	}

	//Decode/validate it
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("expected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		c.Abort()
		return
	}

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		//Check the exp
		if float64(time.Now().Unix()) > float64(claims.ExpiresAt) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized. Token has expired.",
			})
			c.Abort()
			return
		}
		//Find the user with the token sub
		user, err := models.GetUserByID(initializers.DB, claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}
		//Attach to req
		user.Password = ""
		c.Set("user", user)

		//Continue
		c.Next()
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		c.Abort()
		return
	}
}
