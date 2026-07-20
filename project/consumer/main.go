package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"simple_service/pkg"
	"simple_service/pkg/llm"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type TaskMessage struct {
	ID     string `json:"id"`
	Prompt string `json:"prompt"`
}

func main() {

	srv := pkg.NewService() //构建服务
	log.Println("Kafka consumer started...")
	// 加载 .env 文件（如果不存在则跳过）
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY not set")
	}
	var reader *kafka.Reader = pkg.GetInstance[kafka.Reader](srv)
	var dbService *pkg.DbService = pkg.GetInstance[pkg.DbService](srv)
	var rdb *redis.Client = pkg.GetInstance[redis.Client](srv)
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

		// SSE 流式调用
		reply, err := llm.ProcessTaskStreamly(apiKey, task.ID, task.Prompt, srv)

		if err != nil {
			log.Printf("LLM error: %v", err)
			return
		}
		// TODO: 处理完成后更新 MySQL 和 Redis
		log.Printf("task %s done", reply)
		var tk pkg.Task
		tk.ID = task.ID
		tk.Prompt = task.Prompt
		tk.Result = reply
		tk.Status = "done"
		dbService.CompleteTaskWithLog(&tk)
		// 删除缓存（如果有）
		_, err = rdb.Get(ctx, "task:"+task.ID).Result()
		if err == nil {
			rdb.Del(ctx, "task:"+task.ID)
		}

	}
}
