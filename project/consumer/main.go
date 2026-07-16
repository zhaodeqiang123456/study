package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

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
		var taskMsg map[string]interface{}
		if err := json.Unmarshal(msg.Value, &taskMsg); err != nil {
			log.Printf("unmarshal error: %v", err)
			continue
		}

		taskID, _ := taskMsg["id"].(string)
		log.Printf("processing task: %s", taskID)

		// 模拟处理任务（比如 sleep 2秒）
		time.Sleep(2 * time.Second)

		// TODO: 处理完成后更新 MySQL 和 Redis（你可以在后续加固阶段加入）
		log.Printf("task %s done", taskID)
	}
}
