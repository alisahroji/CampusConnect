package handler

import (
	"campusconnect/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	service service.SearchService
}

func NewSearchHandler(service service.SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

// Search mencari user dan project sekaligus (publik)
// GET /api/search?q=<kata kunci>&limit=10
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.service.Search(query, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Berhasil",
		"query":    query,
		"users":    result.Users,
		"projects": result.Projects,
	})
}
