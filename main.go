package main

import (
	"io"
	"os"

	"github.com/dika-bosnjak/social-media-app/pkg/initializers"
	"github.com/dika-bosnjak/social-media-app/pkg/middleware"
	"github.com/dika-bosnjak/social-media-app/pkg/routes"
	"github.com/gin-gonic/gin"
)

// runs before the main function
func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	//models.SyncDatabase()
}

func main() {
	// Logging to a file.
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	//gin default router
	r := gin.Default()

	//use cors middleware
	middleware.EnableCORS(r)

	//import routes
	routes.Router(r)

	//run the server on port 8080
	r.Run(":8080")
}
