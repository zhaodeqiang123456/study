package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"simple_service/pkg"
	"simple_service/pkg/llm"
	"simple_service/pkg/rag"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type TaskMessage struct {
	ID             string `json:"id"`
	Prompt         string `json:"prompt"`
	ConversationID string `json:"conversation_id"`
}

var documentChunks []string

func main() {

	srv := pkg.NewService() //构建服务
	log.Println("Kafka consumer started...")
	// 加载 .env 文件（如果不存在则跳过）
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}
	// 获取大语言模型的apikey
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY not set")
	}
	var reader *kafka.Reader = pkg.GetInstance[kafka.Reader](srv)
	var dbService *pkg.DbService = pkg.GetInstance[pkg.DbService](srv)
	var rdb *redis.Client = pkg.GetInstance[redis.Client](srv)
	// 建立qrant向量数据库连接
	srv.CreateCollection()
	// 预加载文本数据，插入向量数据库
	documentChunksTemp, err := rag.LoadKnowledgeBase()
	documentChunks = documentChunksTemp
	if err != nil {
		log.Print(err.Error())
	}
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		msg, err := reader.ReadMessage(ctx)
		defer cancel()
		if err != nil {
			log.Printf("read message error: %v", err)
			continue
		}

		// 反序列化消息
		var task TaskMessage
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			log.Printf("unmarshal error: %v", err)
			continue
		}
		log.Printf("processing task %s, prompt: %s", task.ID, task.Prompt)

		// // 调用 DeepSeek
		// reply, err := llm.CallDeepSeekWithSDK(apiKey, task.Prompt)
		// if err != nil {
		// 	log.Printf("LLM error: %v", err)
		// 	reply = "AI 服务暂时不可用，请稍后重试"
		// }

		// ==================== 向量检索 ====================
		// queryVector, _ := llm.GetEmbedding(task.Prompt)
		// // 检索相关文档
		// var docs []string
		// if queryVector != nil {
		// 	docs, err = rag.SearchByVector(queryVector, 3)
		// 	if err != nil {
		// 		log.Printf("search error: %v", err)
		// 		docs = nil
		// 	}
		// }

		// ========================= 关键词搜索 =============
		docs := searchByKeyword(task.Prompt)
		// 构建增强 prompt
		augmentedPrompt := task.Prompt
		if len(docs) > 0 {
			augmentedPrompt = fmt.Sprintf(
				"你是一个专业的AI助手。以下是可能相关的参考资料：\n\n%s\n\n"+
					"用户问题：%s\n\n"+
					"请根据参考资料提供准确答案，并在此基础上适当补充背景知识、实际案例或相关技术细节，使回答更加丰富和有用。",
				strings.Join(docs, "\n\n"),
				task.Prompt,
			)
		}
		// ==================== 流式调用 ====================
		history := dbService.QueryHistory(task.ConversationID)
		// 把增强的prompt追加到最后
		history = append(history, openai.UserMessage(augmentedPrompt))
		// SSE 流式调用
		reply, err := llm.ProcessTaskStreamly(apiKey, task.ID, history, srv)

		if err != nil {
			log.Printf("LLM error: %v", err)
			return
		}
		// TODO: 处理完成后更新 MySQL 和 Redis
		log.Printf("task %s done", reply)
		tk := pkg.Task{
			ID:     task.ID,
			Prompt: task.Prompt,
			Result: reply,
			Status: "done",
		}
		dbService.CompleteTaskWithLog(&tk)
		// 删除缓存（如果有）
		_, err = rdb.Get(ctx, "task:"+task.ID).Result()
		if err == nil {
			rdb.Del(ctx, "task:"+task.ID)
		}

		// 事务：插入两条消息并更新会话
		tx, _ := dbService.GetdbInstance().Begin()
		tx.Exec("INSERT INTO messages (conversation_id, role, content, created_at) VALUES (?, 'user', ?, NOW())", task.ConversationID, task.Prompt)
		tx.Exec("INSERT INTO messages (conversation_id, role, content, created_at) VALUES (?, 'assistant', ?, NOW())", task.ConversationID, reply)
		tx.Exec("UPDATE conversations SET updated_at = NOW(), title = ? WHERE id = ? AND title = '新对话'", truncateText(task.Prompt, 20), task.ConversationID)
		tx.Commit()

	}
}

func truncateText(s string, maxLen int) string {
	runes := []rune(s) // 正确处理中文等多字节字符
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func searchByKeyword(query string) []string {
	keywords := strings.Fields(query)
	var results []string
	for _, chunk := range documentChunks {
		for _, kw := range keywords {
			if strings.Contains(chunk, kw) {
				results = append(results, chunk)
				break
			}
		}
	}
	log.Printf("[RAG] 文件已成功匹配 (%d 个片段), 原文件含有 %d 个片段", len(results), len(documentChunks))
	return results
}
