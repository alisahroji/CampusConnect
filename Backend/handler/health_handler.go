package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck untuk memastikan aplikasi secara umum berjalan
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "CampusConnect Backend is healthy & running!",
	})
}

// ReadyCheck untuk memastikan aplikasi siap menerima lalu lintas
func ReadyCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "Ready",
		"message": "CampusConnect Backend is ready to accept traffic!",
	})
}