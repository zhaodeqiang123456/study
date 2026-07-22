package llm

import (
	"context"
	"fmt"
	"os"
	"simple_service/pkg"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
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
func ProcessTaskStreamly(apiKey, taskID string, history []openai.ChatCompletionMessageParamUnion, srv *pkg.Service) (fullResponse string, err error) {
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
		Model:    openai.ChatModel("deepseek-chat"),
		Messages: history,
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

func GetEmbedding(text string) ([]float32, error) {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("DEEPSEEK_API_KEY")),
		option.WithBaseURL("https://api.deepseek.com/v1"),
	)

	// Model 字段直接使用字符串常量
	// Input 字段使用 SDK 提供的构造函数 EmbeddingNewParamsInputArrayOfStrings
	resp, err := client.Embeddings.New(context.Background(), openai.EmbeddingNewParams{
		Model: openai.EmbeddingModelTextEmbedding3Small,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: param.NewOpt(text), // 将 string 包装为 param.Opt
		},
	})
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	embedding64 := resp.Data[0].Embedding
	// 转换为 []float32
	embedding32 := make([]float32, len(embedding64))
	for i, v := range embedding64 {
		embedding32[i] = float32(v)
	}
	return embedding32, nil
}
