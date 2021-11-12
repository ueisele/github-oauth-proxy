package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func hello(c *gin.Context) {
	c.String(http.StatusOK, "Hello World!")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.GET("/", hello)

	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}