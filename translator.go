package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	defaultLLMBaseURL = "https://api.deepseek.com/v1/"
	defaultLLMModel   = "deepseek-v4-flash"
)

type Translator struct {
	client openai.Client
	model  string
}

func NewTranslatorFromEnv() *Translator {
	apiKey := firstEnv("LLM_API_KEY", "OPENAI_API_KEY")
	if apiKey == "" {
		return nil
	}

	baseURL := firstEnv("LLM_BASE_URL", "OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = defaultLLMBaseURL
	}

	model := firstEnv("LLM_MODEL", "OPENAI_MODEL")
	if model == "" {
		model = defaultLLMModel
	}

	return NewTranslator(apiKey, baseURL, model)
}

func NewTranslator(apiKey, baseURL, model string) *Translator {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)

	return &Translator{
		client: client,
		model:  model,
	}
}

func (t *Translator) Model() string {
	if t == nil {
		return ""
	}

	return t.model
}

func (t *Translator) TranslateArticle(ctx context.Context, title, summary string) (string, error) {
	if t == nil {
		return "", errors.New("translator is not configured")
	}

	params := openai.ChatCompletionNewParams{
		Model: t.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是一名专业新闻翻译。请把英文新闻翻译成自然、准确的简体中文。只输出译文，不解释，不添加原文没有的信息。"),
			openai.UserMessage(buildTranslationPrompt(title, summary)),
		},
		MaxTokens: openai.Int(18000),
	}

	resp, err := t.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("translation response has no choices")
	}

	translation := strings.TrimSpace(resp.Choices[0].Message.Content)
	if translation == "" {
		return "", errors.New("translation response is empty")
	}

	return translation, nil
}

func buildTranslationPrompt(title, summary string) string {
	return strings.TrimSpace(`请翻译下面这条 BBC 新闻。

要求：
1. 标题和摘要分开翻译
2. 保留新闻语气
3. 不要输出英文原文
4. 不要输出解释

标题：
` + strings.TrimSpace(title) + `

摘要：
` + strings.TrimSpace(summary))
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}

	return ""
}
