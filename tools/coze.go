package tools

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CozeChatRequest struct {
	BotID              string                 `json:"bot_id"`
	UserID             string                 `json:"user_id"`
	Stream             bool                   `json:"stream"`
	AdditionalMessages []AdditionalMessage    `json:"additional_messages"`
	Parameters         map[string]interface{} `json:"parameters"`
}

type AdditionalMessage struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	Role        string `json:"role"`
	Type        string `json:"type"`
}

type CozeEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type MessageData struct {
	Content string `json:"content"`
	Type    string `json:"type"`
}

func CozeChatStream(ctx context.Context, c *gin.Context, token, botID, userID, message string) error {
	requestBody := CozeChatRequest{
		BotID:  botID,
		UserID: userID,
		Stream: true,
		AdditionalMessages: []AdditionalMessage{
			{
				Content:     message,
				ContentType: "text",
				Role:        "user",
				Type:        "question",
			},
		},
		Parameters: map[string]interface{}{},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.coze.cn/v3/chat?", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 120 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败，状态码: %d，响应内容: %s", resp.StatusCode, string(errBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	msgID := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			dataStr := strings.TrimPrefix(line, "data:")
			dataStr = strings.TrimSpace(dataStr)

			if dataStr == "[DONE]" {
				break
			}

			var event CozeEvent
			if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
				continue
			}

			if event.Event == "conversation.message.delta" || event.Event == "message" {
				var msgData MessageData
				if err := json.Unmarshal(event.Data, &msgData); err == nil {
					if msgData.Content != "" {
						msgID++
						c.SSEvent("message", map[string]interface{}{
							"id":      msgID,
							"content": msgData.Content,
							"time":    time.Now().Format(time.RFC3339),
						})
						c.Writer.Flush()
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取流式响应失败: %w", err)
	}

	c.SSEvent("done", map[string]interface{}{
		"message": "Stream completed",
	})
	c.Writer.Flush()

	return nil
}
