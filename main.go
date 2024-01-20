package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		if err := CreatePass(); err != nil {
			log.Println("Failed to create pass:", err)
		} else {
			log.Println("Pass created successfully.")
		}

		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/create", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "create endpoint hit",
		})
	})

	r.POST("/design_list", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "design_list endpoint hit",
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
