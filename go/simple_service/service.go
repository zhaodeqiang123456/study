package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type TaskStatus string // 重新声明字符串类型, 使其语义更明确

const (
	StatusPending TaskStatus = "pending" // 任务状态为待解决
	StatusDone    TaskStatus = "done"    // 任务状态为已解决
)

type Task struct {
	ID     string     `json:"id"` // tag标签规范, 在做类型转换时发挥作用
	Status TaskStatus `json:"status"`
	Result string     `json:"result,omitempty"`
}

type TaskStore struct { // 一个任务仓库, 存储了所有任务的向量（任务的存储实际地址）
	mu    sync.RWMutex // 一个读写互斥信号量, 以保证线程互斥访问的安全性
	tasks map[string]*Task
}

func NewTaskStore() *TaskStore { // 新建一个任务存储仓库
	return &TaskStore{
		tasks: make(map[string]*Task), // mu 如果未初始化则默认为0，即当前处于解锁状态
	}
}

func (s *TaskStore) Create() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("%d-%d", time.Now().UnixNano(), len(s.tasks))

	s.tasks[id] = &Task{
		ID:     id,
		Status: StatusPending,
	}

	return id
}

func (s *TaskStore) Get(id string) (t *Task, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok = s.tasks[id]
	return
}

func (s *TaskStore) Complete(id, result string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.tasks[id]; ok {
		t.Status = StatusDone
		t.Result = result
	}
}

// Server 包含路由和任务存储
type Server struct {
	store *TaskStore
}

func (s *Server) handlePostTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := s.store.Create()

	go func(taskID string) {

		// 模拟耗时处理（比如调用模型推理）
		time.Sleep(3 * time.Second) // 用 sleep 模拟实际工作

		result := fmt.Sprintf("task %s completed successfully", taskID)
		s.store.Complete(taskID, result)
		log.Printf("task %s completed", taskID)
	}(id)

	// 立即返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"id":     id,
		"status": string(StatusPending),
		"msg":    "task submitted",
	})

}

// handleGetResult 处理 GET /result?id=xxx 请求
func (s *Server) handleGetResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	// task, ok := s.store.Get(id)
	// if !ok {
	// 	http.Error(w, "task not found", http.StatusNotFound)
	// 	return
	// }
	db := initDB()
	detail, err := getTaskDetail(db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func Service() {
	store := NewTaskStore()
	service := &Server{store: store}

	http.HandleFunc("/task", service.handlePostTask)
	http.HandleFunc("/result", service.handleGetResult)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
