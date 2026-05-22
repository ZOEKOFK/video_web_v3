package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GLMClient struct {
	config Config
	client *http.Client
}

func NewGLMClient(config Config) *GLMClient {
	return &GLMClient{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (g *GLMClient) Chat(ctx context.Context, message string) (string, error) {

	messages := []Message{
		{Role: "user", Content: message},
	}

	return g.ChatWithHistory(ctx, messages)
}

// ChatWithHistory 携带历史对话请求llm的api
func (g *GLMClient) ChatWithHistory(ctx context.Context, messages []Message) (string, error) {
	//根据llm的api要求定义
	requestBody := map[string]interface{}{
		"model":    g.config.Model,
		"messages": messages,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	url := g.config.BaseURL + "/chat/completions"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.config.APIKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GLM API错误 (%d): %s", resp.StatusCode, string(body))
	}

	//返回格式根据llm的api要求定义
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API没有返回内容")
	}

	return result.Choices[0].Message.Content, nil
}
