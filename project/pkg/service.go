package pkg

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// Service 改造为结构体，持有 http.Server
type Service struct {
	mu          sync.RWMutex // 保护所有字段的并发安全
	webService  *http.Server
	store       *TaskStore // 或其他依赖
	rdb         *redis.Client
	kafkaWriter *kafka.Writer
	kafkaReader *kafka.Reader
	dbService   *DbService
}

type TaskRequest struct {
	Prompt string `json:"prompt"`
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

// 初始化 Kafkawriter
func NewKafkaWriter() *kafka.Writer {
	kw := kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "task-queue",
		Balancer: &kafka.LeastBytes{}, // 最简单的分区策略
	}
	return &kw
}

// 初始化 KafkaReader
func NewKafkaReader() *kafka.Reader {
	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "task-queue",
		GroupID:  "task-consumer-group", // 消费者组，随便取
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	return kr
}

// 初始化web服务
func (svc *Service) NewWebService(addr string) *http.Server {
	mux := http.NewServeMux() // 不要用全局默认的 http.HandleFunc
	mux.HandleFunc("/task", svc.handlePostTask)
	mux.HandleFunc("/result", svc.handleGetResult)
	// mux.HandleFunc("/sqlTask", svc.handleSqlTask)
	webServer := http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return &webServer
}

// 泛型+类型断言实现懒加载
func GetInstance[T any](s *Service) *T {
	s.mu.Lock()
	defer s.mu.Unlock()

	var zero T
	switch any(zero).(type) {

	case http.Server:
		if s.webService == nil {
			s.webService = s.NewWebService(Port)
		}
		return any(s.webService).(*T)
	case TaskStore:
		if s.store == nil {
			s.store = NewTaskStore()
		}
		return any(s.store).(*T)

	case redis.Client:
		if s.rdb == nil {
			s.rdb = NewRedisClient()
		}
		return any(s.rdb).(*T)

	case kafka.Writer:
		if s.kafkaWriter == nil {
			s.kafkaWriter = NewKafkaWriter()
		}
		return any(s.kafkaWriter).(*T)
	case DbService:
		if s.dbService == nil {
			s.dbService = &DbService{}
		}
		return any(s.dbService).(*T)
	case kafka.Reader:
		if s.kafkaReader == nil {
			s.kafkaReader = NewKafkaReader()
		}
		return any(s.kafkaReader).(*T)
	default:
		return nil

	}
}

func NewService() *Service {
	// 创建服务
	svc := &Service{}
	return svc
}

func (s *Service) Start() error {
	log.Println("webService starting on")
	var webServer *http.Server = GetInstance[http.Server](s)
	return webServer.ListenAndServe()
}

func (s *Service) Shutdown(ctx context.Context) error {
	var webServer *http.Server = GetInstance[http.Server](s)
	return webServer.Shutdown(ctx)
}

func (s *Service) handlePostTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 解析请求体
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}
	var store *TaskStore = GetInstance[TaskStore](s)
	taskID := store.Create()
	task, _ := store.Get(taskID)

	go func() {
		// 将task插入数据库
		var dbService *DbService = GetInstance[DbService](s)
		err := dbService.InsertTask(task)
		if err != nil {
			log.Printf("insert task error: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// 立即返回响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(task)
	}()

	// 构造 Kafka 消息
	taskMsg := map[string]any{
		"id":     taskID,
		"prompt": req.Prompt,
	}

	msgBytes, _ := json.Marshal(taskMsg)
	var kafkaWriter *kafka.Writer = GetInstance[kafka.Writer](s)
	err := kafkaWriter.WriteMessages(r.Context(), kafka.Message{
		Key:   []byte(taskID),
		Value: msgBytes,
	})

	if err != nil {
		http.Error(w, "failed to enqueue task", http.StatusInternalServerError)
		return
	}
}

// handleGetResult 处理 GET /result?id=xxx 请求
func (s *Service) handleGetResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	task, err := s.getTaskWithCache(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (s *Service) getTaskWithCache(taskID string) (*Task, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 1. 尝试从 Redis 读取
	var rdb *redis.Client = GetInstance[redis.Client](s)
	cached, err := rdb.Get(ctx, "task:"+taskID).Result()
	var task Task
	if err == nil {
		log.Println("get cache successfully")
		// 缓存命中，直接反序列化返回
		json.Unmarshal([]byte(cached), &task)
		return &task, nil
	}
	log.Println("get cache fail, try to get task from sql")
	// 2. 缓存未命中，查 MySQL
	var dbService *DbService = GetInstance[DbService](s)
	task, err = dbService.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	// 3. 回写 Redis，设置 30s 过期
	taskJSON, _ := json.Marshal(task)
	rdb.Set(ctx, "task:"+taskID, taskJSON, 30*time.Second)

	return &task, nil
}
