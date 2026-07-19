package main

import (
	"context"
	"encoding/json"
	"log"
	"simple_service/pkg/llm"
	"time"

	"github.com/segmentio/kafka-go"
)

type TaskMessage struct {
	ID     string `json:"id"`
	Prompt string `json:"prompt"`
}

func main() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "task-queue",
		GroupID:  "task-consumer-group", // 消费者组，随便取
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Println("Kafka consumer started...")

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		msg, err := reader.ReadMessage(ctx)
		cancel()
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

		apiKey := ""
		// 调用 DeepSeek（或你封装好的 llm.CallDeepSeek）
		reply, err := llm.CallDeepSeekWithSDK(apiKey, task.Prompt)
		if err != nil {
			log.Printf("LLM error: %v", err)
			reply = "AI 服务暂时不可用，请稍后重试"
		}

		// TODO: 处理完成后更新 MySQL 和 Redis
		log.Printf("task %s done", reply)

	}
}
