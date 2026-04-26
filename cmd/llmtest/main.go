package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	defaultBaseURL = "https://api.deepseek.com/v1/"
	defaultModel   = "deepseek-v4-flash"
)

func main() {
	apiKey := firstEnv("LLM_API_KEY", "OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("missing LLM_API_KEY or OPENAI_API_KEY")
		os.Exit(1)
	}

	baseURL := firstEnv("LLM_BASE_URL", "OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	model := firstEnv("LLM_MODEL", "OPENAI_MODEL")
	if model == "" {
		model = defaultModel
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	params := openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一名专业新闻翻译。只输出译文，不解释。"),
			openai.UserMessage("Translate this into Simplified Chinese: Global leaders meet to discuss climate change targets."),
		},
		MaxTokens:   openai.Int(300),
		Temperature: openai.Float(0.2),
	}

	resp, err := client.Chat.Completions.New(context.Background(), params)
	if err != nil {
		panic(err)
	}
	if len(resp.Choices) == 0 {
		panic("empty response choices")
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}

	return ""
}
