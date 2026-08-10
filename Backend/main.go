package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Menginisialisasi router Gin dengan default middleware (logger dan recovery)
	r := gin.Default()

	// Endpoint dasar untuk testing health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "CampusConnect Backend is running!",
		})
	})

	// Menjalankan server di port 8080
	log.Println("Server is running on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}