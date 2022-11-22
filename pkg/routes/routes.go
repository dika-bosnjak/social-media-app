package routes

import (
	"fmt"

	"github.com/dika-bosnjak/social-media-app/pkg/controllers"
	"github.com/dika-bosnjak/social-media-app/pkg/middleware"
	"github.com/dika-bosnjak/social-media-app/pkg/websocketrooms"
	"github.com/gin-gonic/gin"
)

var Router = func(r *gin.Engine) {

	r.GET("/ping", func(c *gin.Context) {
		fmt.Printf("ClientIP: %s\n", c.ClientIP())
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	r.GET("/notifications", middleware.RequireAuth, controllers.ReadNotifications)
	r.GET("/users", middleware.RequireAuth, controllers.SearchUser)

	r.GET("/friendshipRequests", middleware.RequireAuth, controllers.ShowFriendshipRequests)
	r.GET("/friendshipStatus/:id", middleware.RequireAuth, controllers.FriendshipStatus)

	r.GET("/posts", middleware.RequireAuth, controllers.DisplayPostsOnHomePage)

	user := r.Group("/user")
	{
		user.GET("/", middleware.RequireAuth, controllers.DisplayLoggedInUser)
		user.PUT("/", middleware.RequireAuth, controllers.UpdateUser)
		user.DELETE("/", middleware.RequireAuth, controllers.DeleteUser)

		user.GET("/:id", middleware.RequireAuth, controllers.DisplayProfile)

		user.GET("/:id/posts", middleware.RequireAuth, controllers.DisplayPostsByUserId)

		//user.GET("/friends", middleware.RequireAuth, controllers.ShowFriends)
		user.GET("/:id/friends", middleware.RequireAuth, controllers.ShowFriends)
		user.POST("/:id/add", middleware.RequireAuth, controllers.AddFriend)
		user.DELETE("/:id/remove", middleware.RequireAuth, controllers.DeleteFriend)

		user.GET("/blocked-users", middleware.RequireAuth, controllers.ShowBlockedUsers)
		user.POST("/:id/block", middleware.RequireAuth, controllers.BlockUser)
		user.DELETE("/:id/unblock", middleware.RequireAuth, controllers.UnblockUser)
	}

	friendshipRequest := r.Group("/friendshipRequest")
	{
		friendshipRequest.PUT("/:id/accept", middleware.RequireAuth, controllers.AcceptFriendshipRequest)
		friendshipRequest.DELETE("/:id/decline", middleware.RequireAuth, controllers.DeclineFriendshipRequest)
	}

	post := r.Group("/post")
	{
		post.POST("/", middleware.RequireAuth, controllers.CreatePost)
		post.GET("/:id", middleware.RequireAuth, controllers.DisplayPost)
		post.PUT("/:id", middleware.RequireAuth, controllers.UpdatePost)
		post.DELETE("/:id", middleware.RequireAuth, controllers.DeletePost)

		post.POST("/:id/like", middleware.RequireAuth, controllers.LikePost)
		post.POST("/:id/comment", middleware.RequireAuth, controllers.AddComment)
	}

	r.DELETE("/comment/:id", middleware.RequireAuth, controllers.DeleteComment)

	r.GET("/chatroom/:userID", middleware.RequireAuth, controllers.OpenChatRoom)

	wsServer := websocketrooms.NewWebsocketServer()
	go wsServer.Run()

	r.GET("/websocket/:userID", func(c *gin.Context) {
		userID := c.Param("userID")
		websocketrooms.ServeWs(wsServer, c.Writer, c.Request, userID)
	})
	r.GET("/websocket/", func(c *gin.Context) {
		websocketrooms.ServeWs(wsServer, c.Writer, c.Request, "")
	})

	r.GET("/websocket/notifications/:userID", func(c *gin.Context) {
		userID := c.Param("userID")
		websocketrooms.ServeWsForNotifications(wsServer, c.Writer, c.Request, userID)
	})

}
