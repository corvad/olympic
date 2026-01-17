package olympic

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var database *DB
var accountManager *AccountManager

func createAccount(c *gin.Context) {

	c.Request.ParseForm()
	email := c.Request.Form.Get("email")
	password := c.Request.Form.Get("password")
	_, err := accountManager.CreateAccount(email, password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
		})
		log.Println("CreateAccount error:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func login(c *gin.Context) {
	c.Request.ParseForm()
	email := c.Request.Form.Get("email")
	password := c.Request.Form.Get("password")
	refreshToken, jwtToken, err := accountManager.Login(email, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
		log.Println("Authentication error:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"refreshToken": refreshToken, // Return refresh token
		"jwtToken":     jwtToken,     // Return JWT token
	})
}

func createLink(c *gin.Context) {

	c.Request.ParseForm()
	url := c.Request.Form.Get("url")

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"url":    url,
	})
}

func getLink(c *gin.Context) {

	c.Request.ParseForm()
	url := c.Request.Form.Get("url")

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"url":    url,
	})
}

func registerRoutes(r *gin.Engine) {
	// Define a simple GET endpoint
	r.GET("/api/createAccount", createAccount)
	r.GET("/api/createLink", createLink)
	r.GET("/api/getLink", getLink)
	r.GET("/api/login", login)
}

func Init(dbName string, jwtSigningSecret string) {
	var err error
	database, err = OpenDB(dbName)
	if err != nil {
		log.Panicf("failed to connect database: %v", err)
	}
	log.Println("Connected to Database:", dbName)
	accountManager = &AccountManager{db: database.DB, jwtSigningSecret: jwtSigningSecret}
}

func Run(port int) {
	// Set Gin to release mode to disable debug output
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Define a simple GET endpoint
	r.GET("/ping", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	registerRoutes(r)

	log.Printf("Starting server on port %d...", port)
	r.Run(":" + fmt.Sprintf("%d", port))
}

func Shutdown() {
	if database != nil {
		err := database.Close()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed.")
		}
	}
}
