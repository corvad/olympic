package olympic

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func createAccount(c *gin.Context) {

	c.Request.ParseForm()
	email := c.Request.Form.Get("email")
	password := c.Request.Form.Get("password")

	c.JSON(http.StatusOK, gin.H{
		"message": "Account created successfully",
	})
}

func createLink(c *gin.Context) {

	c.Request.ParseForm()
	url := c.Request.Form.Get("url")

	c.JSON(http.StatusOK, gin.H{
		"message": "Link created successfully",
	})
}

func getLink(c *gin.Context) {

	c.Request.ParseForm()
	url := c.Request.Form.Get("url")

	c.JSON(http.StatusOK, gin.H{
		"message": "Link retrieved successfully",
	})
}

func registerRoutes(r *gin.Engine) {
	// Define a simple GET endpoint
	r.POST("/createAccount", createAccount)
	r.POST("/createLink", createLink)
	r.GET("/getLink", getLink)
}

func Init() {
	r := gin.Default()

	// Define a simple GET endpoint
	r.GET("/ping", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	r.Run()
}
