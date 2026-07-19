package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	deepseekBaseURL = "https://api.deepseek.com/v1"
	deepseekModel   = "deepseek-chat"
)

func CallDeepSeekWithSDK(apiKey, userPrompt string) (string, error) {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(deepseekBaseURL),
	)

	completion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		// 模型字段需要 openai.ChatModel 类型，可用 openai.ChatModel(deepseekModel) 或直接字符串
		Model: openai.ChatModel(deepseekModel),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(userPrompt),
		},
	})
	if err != nil {
		return "", err
	}
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no choices")
	}
	return completion.Choices[0].Message.Content, nil
}
