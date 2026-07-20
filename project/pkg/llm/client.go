package llm

import (
	"context"
	"fmt"
	"simple_service/pkg"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/redis/go-redis/v9"
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

// 流式处理，每收到一个 chunk 就 RPUSH 到 Redis
func ProcessTaskStreamly(apiKey, taskID, prompt string, srv *pkg.Service) (fullResponse string, err error) {
	// ... 幂等性检查、事务等

	streamKey := "stream:" + taskID
	// 设置过期时间，避免内存泄漏
	var rdb *redis.Client = pkg.GetInstance[redis.Client](srv)
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	rdb.Expire(context.Background(), streamKey, 10*time.Minute)

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(deepseekBaseURL),
	)
	// 构造流式请求
	stream := client.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
		Model: openai.ChatModel("deepseek-chat"),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	})
	defer stream.Close()

	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				// 推入 Redis 列表
				rdb.RPush(context.Background(), streamKey, delta)
				fullResponse += delta
			}
		}
	}
	if err := stream.Err(); err != nil {
		rdb.RPush(context.Background(), streamKey, "[ERROR] "+err.Error())
		return fullResponse, err
	}

	// 流结束后，推送一个特殊结束标记
	rdb.RPush(context.Background(), streamKey, "[DONE]")

	return fullResponse, err
}
