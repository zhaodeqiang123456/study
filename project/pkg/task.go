package pkg

import (
	"fmt"
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
