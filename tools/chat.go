package tools

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type MessageRequest struct {
	Content string `json:"content" binding:"required"`
}

type ChatSession struct {
	mu       sync.Mutex
	messages chan string
}

var chatSession = &ChatSession{
	messages: make(chan string, 10),
}

func Chat(c *gin.Context) {
	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	chatSession.messages <- req.Content
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func ChatSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	clientGone := c.Writer.CloseNotify()
	token := "sat_nYBqdtrceTm4e9xpfdFT9tLIGSRTohdQzcXk8jcsaiHBbHGgteATUCYeh576uHXz"
	botID := "7592415425774911507"
	userID := "123456789"
	question := c.Query("question")

	go func() {
		CozeChatStream(c.Request.Context(), c, token, botID, userID, question)
	}()
	for {
		select {
		case <-clientGone:
			fmt.Println("Client disconnected")
			return
		case content := <-chatSession.messages:
			fmt.Println(`收到消息：`, content)
		}
	}

}
