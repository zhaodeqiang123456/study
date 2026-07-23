package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"simple_service/pkg"
	"simple_service/pkg/agent"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/redis/go-redis/v9"
)

const (
	deepseekBaseURL = "https://api.deepseek.com/v1"
	deepseekModel   = "deepseek-chat"
	maxIterations   = 10
)

type AgentEvent struct {
	Type string      `json:"type"` // text_delta, tool_call, tool_result, done
	Data interface{} `json:"data"`
}

var tools []openai.ChatCompletionToolParam = []openai.ChatCompletionToolParam{
	{
		Type: "function",
		Function: openai.FunctionDefinitionParam{
			Name:        "calculator",
			Description: openai.String("计算数学表达式，支持加减乘除"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]interface{}{
						"type":        "string",
						"description": "数学表达式，例如 '2+3'",
					},
				},
				"required": []string{"expression"},
			},
		},
	},
	{
		Type: "function",
		Function: openai.FunctionDefinitionParam{
			Name:        "get_weather",
			Description: openai.String("查询指定城市的天气"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"city": map[string]interface{}{
						"type":        "string",
						"description": "城市名称，例如北京",
					},
				},
				"required": []string{"city"},
			},
		},
	},
	{
		Type: "function",
		Function: openai.FunctionDefinitionParam{
			Name:        "search_knowledge",
			Description: openai.String("搜索本地知识库，获取与查询相关的文档片段。当需要了解特定领域的知识（如公司政策、产品信息、技术规范）时使用。"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "搜索查询，应使用完整的问题或关键词",
					},
				},
				"required": []string{"query"},
			},
		},
	},
}

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
func ProcessTaskStreamly(apiKey, taskID string, messages []openai.ChatCompletionMessageParamUnion, srv *pkg.Service) (fullResponse string, err error) {
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
		Messages: messages,
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

func CallDeepSeekWithToolsAndSSE(apiKey string, taskID string, history []openai.ChatCompletionMessageParamUnion, srv *pkg.Service) (string, error) {

	messages := history
	streamKey := "stream:" + taskID

	// 设置过期时间，避免内存泄漏
	var rdb *redis.Client = pkg.GetInstance[redis.Client](srv)
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	rdb.Expire(context.Background(), streamKey, 10*time.Minute)

	var fullText strings.Builder
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(deepseekBaseURL),
	)

	for range maxIterations {

		fullText.Reset()
		// 发送流式请求（stream: true）
		stream := client.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
			Model:    openai.ChatModel(deepseekModel), // 注意：v1.12.0 中 Model 可能是 string，如果是 openai.ChatModel 类型，可以这样写：openai.ChatModel("deepseek-chat")
			Messages: messages,
			Tools:    tools,
		})

		var curToolCalls []openai.ChatCompletionChunkChoiceDeltaToolCall

		// 读取流式响应
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) == 0 {
				continue
			}
			delta := chunk.Choices[0].Delta

			// 如果有工具调用
			if len(delta.ToolCalls) > 0 {
				curToolCalls = agent.MergeDeltaToolCalls(curToolCalls, delta.ToolCalls)
			}

			// 如果有文本内容
			if delta.Content != "" {
				fullText.WriteString(delta.Content)
				// 推送文本增量事件到 Redis
				// 推入 Redis 列表
				pushEvent(rdb, streamKey, AgentEvent{
					Type: "text_delta",
					Data: delta.Content,
				})
			}

		}

		if err := stream.Err(); err != nil {
			pushEvent(rdb, streamKey, AgentEvent{Type: "error", Data: err.Error()})
			return fullText.String(), err
		}

		// 循环结束后，如果有工具调用，转换为 MessageToolCall 并构造消息
		if len(curToolCalls) > 0 {

			// 推送工具调用事件
			for _, tc := range curToolCalls {
				pushEvent(rdb, streamKey, AgentEvent{
					Type: "tool_call",
					Data: map[string]string{
						"name": tc.Function.Name,
						"args": tc.Function.Arguments,
					},
				})
			}
			// 构造消息级别的工具调用列表
			var msgToolCalls []openai.ChatCompletionMessageToolCallParam
			for _, dtc := range curToolCalls {
				msgToolCalls = append(msgToolCalls, openai.ChatCompletionMessageToolCallParam{
					ID:   dtc.ID, // 流式过程中 ID 可能为空，但一般第一个 chunk 就有
					Type: "function",
					Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      dtc.Function.Name,
						Arguments: dtc.Function.Arguments,
					},
				})
			}
			// 加入助手消息（带 tool_calls）
			assistantMsg := openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{
					Content: openai.ChatCompletionAssistantMessageParamContentUnion{
						OfString: openai.String(fullText.String()),
					},
					ToolCalls: msgToolCalls,
				},
			}
			messages = append(messages, assistantMsg)
			// 将本轮请求返回的所有工具调用执行并将结果append到messages中
			for _, tc := range msgToolCalls {
				log.Printf("Executing tool: %s(%s)", tc.Function.Name, tc.Function.Arguments)

				var args map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					log.Printf("arguments parse error: %v", err)
					args = map[string]interface{}{}
				}

				result, err := agent.ExecuteTool(tc.Function.Name, args) // 调用你之前写的工具执行函数
				if err != nil {
					result = "错误: " + err.Error()
				}
				// 推送工具结果
				pushEvent(rdb, streamKey, AgentEvent{
					Type: "tool_result",
					Data: map[string]string{
						"name":   tc.Function.Name,
						"result": result,
					},
				})
				// 添加工具结果消息
				messages = append(messages, openai.ToolMessage(result, tc.ID))
			}

		} else {
			break
		}

		// 继续循环，让模型处理工具结果
	}

	// 没有工具调用，最终文本已通过 text_delta 推送完毕，只需告知 done
	pushEvent(rdb, streamKey, AgentEvent{Type: "done", Data: ""})

	// 如果 LLM 直接返回文本（没有工具调用），说明任务完成, 再次通过SSE接口，流式返回
	return fullText.String(), nil
}

func pushEvent(rdb *redis.Client, streamKey string, event AgentEvent) {
	data, _ := json.Marshal(event)
	rdb.RPush(context.Background(), streamKey, string(data))
}
