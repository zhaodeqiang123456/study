package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"simple_service/pkg"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// Service 改造为结构体，持有 http.Server
type Service struct {
	server      *http.Server
	store       *pkg.TaskStore // 或其他依赖
	rdb         *redis.Client
	kafkaWriter *kafka.Writer
}

// 初始化 Redis 客户端
func NewRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 如果有密码就填
		DB:       0,
	})
	// 测试连接
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("连接 Redis 失败: %v", err)
	}
	log.Println("Redis 连接成功")
	return rdb
}

func NewService(addr string) *Service {
	store := pkg.NewTaskStore()
	rdb := NewRedisClient()
	kw := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "task-queue",
		Balancer: &kafka.LeastBytes{}, // 最简单的分区策略
	}
	svc := &Service{
		store:       store,
		rdb:         rdb,
		kafkaWriter: kw,
	}

	mux := http.NewServeMux() // 不要用全局默认的 http.HandleFunc
	mux.HandleFunc("/task", svc.handlePostTask)
	mux.HandleFunc("/result", svc.handleGetResult)
	mux.HandleFunc("/sqlTask", svc.handleSqlTask)

	svc.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return svc
}

func (s *Service) Start() error {
	log.Println("Server starting on", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Service) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Service) handlePostTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := "task-001"
	task := pkg.Task{
		ID:     taskID,
		Status: "pending",
	}

	msgBytes, _ := json.Marshal(task)

	err := s.kafkaWriter.WriteMessages(r.Context(), kafka.Message{
		Key:   []byte(taskID),
		Value: msgBytes,
	})

	if err != nil {
		http.Error(w, "failed to enqueue task", http.StatusInternalServerError)
		return
	}

	// id := s.store.Create()

	// go func(taskID string) {

	// 	// 模拟耗时处理（比如调用模型推理）
	// 	time.Sleep(3 * time.Second) // 用 sleep 模拟实际工作

	// 	result := fmt.Sprintf("task %s completed successfully", taskID)
	// 	s.store.Complete(taskID, result)
	// 	log.Printf("task %s completed", taskID)
	// }(id)

	// 立即返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"id":     taskID,
		"status": string(pkg.StatusPending),
		"msg":    "task submitted",
	})

}

// handleGetResult 处理 GET /result?id=xxx 请求
func (s *Service) handleGetResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	log.Println("get id = ", id)
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	// task, ok := s.store.Get(id)
	// if !ok {
	// 	http.Error(w, "task not found", http.StatusNotFound)
	// 	return
	// }

	// detail, err := inst.dbS.getTaskDetail(id)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusNotFound)
	// 	return
	// }
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(detail)
	if s.rdb == nil {
		log.Println("FATAL: s.rdb is nil in getTaskWithCache")
	}
	task, err := s.getTaskWithCache(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)

}

func (s *Service) handleSqlTask(w http.ResponseWriter, r *http.Request) {

	taskID := "task-002"
	err := inst.GetDbS().CompleteTaskWithLog(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("任务已自动完成")
	}
}

func (s *Service) getTaskWithCache(taskID string) (*pkg.Task, error) {

	var ctx = context.Background()
	// 1. 尝试从 Redis 读取
	cached, err := s.rdb.Get(ctx, "task:"+taskID).Result()
	var task pkg.Task
	if err == nil {
		log.Println("get cache successfully")
		// 缓存命中，直接反序列化返回
		json.Unmarshal([]byte(cached), &task)
		return &task, nil
	}
	log.Println("get cache fail, try to get task from sql")
	// 2. 缓存未命中，查 MySQL
	task, err = inst.GetDbS().GetTask(taskID)
	if err != nil {
		return nil, err
	}

	// 3. 回写 Redis，设置 30s 过期
	taskJSON, _ := json.Marshal(task)
	s.rdb.Set(ctx, "task:"+taskID, taskJSON, 30*time.Second)

	return &task, nil
}
