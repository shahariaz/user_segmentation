package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shahariaz/user_segmentation/dgraph"
	"github.com/shahariaz/user_segmentation/internal/converter"
	models "github.com/shahariaz/user_segmentation/internal/model"
)

type QueryHandler struct {
	converter    *converter.Converter
	dgraphClient *dgraph.Client
}

func NewQueryHandler() *QueryHandler {

	dgraphClient, err := dgraph.NewClient(dgraph.DefaultConfig())
	if err != nil {

		fmt.Printf("⚠️ Warning: Could not connect to Dgraph: %v\n", err)
		fmt.Println("💡 To use /execute endpoint, start Dgraph with: docker-compose up -d")
	}

	return &QueryHandler{
		converter: converter.NewConverter(),

		dgraphClient: dgraphClient,
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

	dqlString := h.converter.GenerateDQLString(dqlQuery)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := h.dgraphClient.ExecuteDQL(ctx, dqlString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Query execution failed",
			"details":   err.Error(),
			"dql_query": dqlString,
		})
		return
	}

	stats := h.dgraphClient.GetExecutionStats(response)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response.Data,
		"query_info": gin.H{
			"dql":        dqlString,
			"query_time": response.QueryTime,
			"stats":      stats,
		},
	})

}
