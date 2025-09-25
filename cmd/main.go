package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/shahariaz/user_segmentation/internal/handler"
)

func main() {
	router := gin.Default()
	queryHandler := handler.NewQueryHandler()

	api := router.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		api.POST("/query", queryHandler.HandleQuery)
		api.POST("/execute", queryHandler.ExecuteQuery)
	}

	log.Fatal(router.Run(":8010"))
}
