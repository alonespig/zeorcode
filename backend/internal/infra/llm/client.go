package llm

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"
)

var ErrDisabled = errors.New("ai model is not configured")

type Message struct {
	Role    string
	Content string
}

// Client keeps all provider-specific Eino types behind one local boundary.
// Business services receive plain text deltas and never receive the API key.
type Client struct {
	model   einomodel.BaseChatModel
	enabled bool
	name    string
	initErr error
}

func NewClient() *Client {
	enabled := viper.GetBool("ai.enabled")
	apiKey := firstNonEmpty(os.Getenv("ZEORCODE_AI_API_KEY"), viper.GetString("ai.api_key"))
	modelName := firstNonEmpty(os.Getenv("ZEORCODE_AI_MODEL"), viper.GetString("ai.model"))
	baseURL := firstNonEmpty(os.Getenv("ZEORCODE_AI_BASE_URL"), viper.GetString("ai.base_url"))
	if !enabled || apiKey == "" || modelName == "" {
		return &Client{enabled: false, name: modelName, initErr: ErrDisabled}
	}
	timeout := time.Duration(viper.GetInt("ai.request_timeout_seconds")) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	maxTokens := viper.GetInt("ai.max_output_tokens")
	if maxTokens <= 0 {
		maxTokens = 2048
	}
	temperature := float32(0.2)
	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey: apiKey, Model: modelName, BaseURL: baseURL, Timeout: timeout,
		MaxCompletionTokens: &maxTokens, Temperature: &temperature,
	})
	if err != nil {
		return &Client{enabled: false, name: modelName, initErr: err}
	}
	return &Client{model: chatModel, enabled: true, name: modelName}
}

func (c *Client) Enabled() bool { return c != nil && c.enabled && c.model != nil }
func (c *Client) ModelName() string {
	if c == nil || c.name == "" {
		return "deterministic-workflow"
	}
	return c.name
}
func (c *Client) InitError() error {
	if c == nil {
		return ErrDisabled
	}
	return c.initErr
}

func (c *Client) Generate(ctx context.Context, system string, history []Message) (string, error) {
	if !c.Enabled() {
		return "", ErrDisabled
	}
	messages := buildMessages(system, history)
	result, err := c.model.Generate(ctx, messages)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Content), nil
}

// Stream sends model-generated text chunks to onDelta and returns the complete
// answer for persistence. The reader is always closed, including cancellation
// and downstream write failures.
func (c *Client) Stream(ctx context.Context, system string, history []Message, onDelta func(string) error) (string, error) {
	if !c.Enabled() {
		return "", ErrDisabled
	}
	reader, err := c.model.Stream(ctx, buildMessages(system, history))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var answer strings.Builder
	for {
		chunk, recvErr := reader.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return "", recvErr
		}
		if chunk == nil || chunk.Content == "" {
			continue
		}
		answer.WriteString(chunk.Content)
		if onDelta != nil {
			if err := onDelta(chunk.Content); err != nil {
				return "", err
			}
		}
	}
	return strings.TrimSpace(answer.String()), nil
}

func buildMessages(system string, history []Message) []*schema.Message {
	messages := make([]*schema.Message, 0, len(history)+1)
	messages = append(messages, schema.SystemMessage(system))
	for _, item := range history {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		if item.Role == "assistant" {
			messages = append(messages, schema.AssistantMessage(content, nil))
		} else {
			messages = append(messages, schema.UserMessage(content))
		}
	}
	return messages
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
