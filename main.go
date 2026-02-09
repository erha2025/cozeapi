package main

import (
	"time"

	"cozeapi/tools"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.StaticFile("/", "./index.html")

	r.POST("/api/chat/send", tools.Chat)
	r.GET("/api/chat/sse", tools.ChatSSE)

	r.Run(":8080")
}
