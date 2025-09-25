package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shahariaz/user_segmentation/internal/converter"
	models "github.com/shahariaz/user_segmentation/internal/model"
)

type QueryHandler struct {
	converter *converter.Converter
}

func NewQueryHandler() *QueryHandler {
	return &QueryHandler{
		converter: converter.NewConverter(),
	}
}

func (h *QueryHandler) HandleQuery(c *gin.Context) {
	var jsonQuery models.JSONQuery
	if err := c.ShouldBindJSON(&jsonQuery); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	dqlQuery, err := h.converter.ConvertToDQL(&jsonQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to convert query",
			"details": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"status": "query received", "query": dqlQuery})

}
func (h *QueryHandler) ExecuteQuery(c *gin.Context) {
	var jsonQuery models.JSONQuery
	if err := c.ShouldBindJSON(&jsonQuery); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	dqlQuery, err := h.converter.ConvertToDQL(&jsonQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to convert query",
			"details": err.Error(),
		})
		return
	}
	stringQuery := h.converter.GenerateDQLString(dqlQuery)
	c.JSON(200, gin.H{"status": "query executed", "query": stringQuery})
}
