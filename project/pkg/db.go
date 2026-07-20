package pkg

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql" // 记得导入驱动
	"github.com/openai/openai-go"
)

type Exec int

type DbService struct {
	mu sync.RWMutex
	db *sql.DB
}

type TaskDetail struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Logs   []string `json:"logs"`
}

func (dbS *DbService) GetdbInstance() *sql.DB {
	dbS.mu.Lock()
	defer dbS.mu.Unlock()
	if dbS.db == nil {
		dbS.initDB()
	}
	return dbS.db
}

func (dbS *DbService) initDB() {
	var err error
	dbS.db, err = sql.Open("mysql", "root:000922@tcp(127.0.0.1:3306)/mysql_learn")
	if err != nil {
		log.Fatal(err)
	}
	if err = dbS.db.Ping(); err != nil {
		log.Fatal("无法连接到数据库:", err)
	}
	fmt.Println("数据库连接成功！")
}

func (dbS *DbService) getTaskDetail(taskID string) (*TaskDetail, error) {
	var detail TaskDetail
	err := dbS.GetdbInstance().QueryRow("SELECT id, status FROM tasks WHERE id = ?", taskID).Scan(&detail.ID, &detail.Status)
	if err != nil {
		return nil, err
	}

	// 2. 查关联日志（利用索引）
	rows, err := dbS.GetdbInstance().Query("SELECT log_msg FROM task_logs WHERE task_id = ? ORDER BY created_at", taskID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg string
		if err := rows.Scan(&msg); err != nil {
			return nil, err
		}
		detail.Logs = append(detail.Logs, msg)
	}

	return &detail, rows.Err()
}

// 完成任务并记录日志（事务演示）
func (dbS *DbService) CompleteTaskWithLog(task *Task) error {
	tx, err := dbS.GetdbInstance().Begin() // 开启事务
	if err != nil {
		return err
	}
	// 确保事务最终提交或回滚
	defer func() {
		if err != nil {
			tx.Rollback() // 出错回滚
		} else {
			tx.Commit() // 提交
		}
	}()

	var status string
	// FOR UPDATE 加排他锁
	err = tx.QueryRow("SELECT status FROM tasks WHERE id = ? FOR UPDATE", task.ID).Scan(&status)
	if err != nil {
		return err
	}
	if status == "done" {
		return nil // 幂等跳过
	}
	// 操作1：更新任务状态
	_, err = tx.Exec("UPDATE tasks SET status = 'done', result = ? WHERE id = ?", task.Result, task.ID)
	if err != nil {
		return err
	}

	// // 操作2：插入日志
	// _, err = tx.Exec("INSERT INTO task_logs (task_id, log_msg, created_at) VALUES (?, ?, NOW())", task.ID, "任务已自动完成")
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (dbS *DbService) InsertTask(task *Task) error {
	_, err := dbS.GetdbInstance().Exec("insert into tasks (id, status, result, created_at, conversation_id) values (?, ?, ?, now(), ?)", task.ID, task.Status, task.Result, task.ConversationID)
	return err
}

func (dbS *DbService) GetTask(taskID string) (Task, error) {
	var task Task
	err := dbS.GetdbInstance().QueryRow("select id, status, result from tasks where id = ?", taskID).Scan(&task.ID, &task.Status, &task.Result)
	return task, err
}

func (dbS *DbService) InsertConversation() (string, error) {

	id := fmt.Sprintf("%d", time.Now().UnixNano()) // 简单 ID，生产环境建议 uuid
	now := time.Now()
	_, err := dbS.GetdbInstance().Exec("INSERT INTO conversations (id, title, created_at, updated_at) VALUES (?, ?, ?, ?)", id, "新对话", now, now)
	return id, err
}

func (dbS *DbService) QueryConversation() (any, error) {

	rows, err := dbS.GetdbInstance().Query("SELECT id, title, updated_at FROM conversations ORDER BY updated_at DESC")

	if err != nil { /* 错误处理 */
	}
	defer rows.Close()
	var convs []map[string]interface{}
	for rows.Next() {
		var id, title string
		var updated time.Time
		rows.Scan(&id, &title, &updated)
		convs = append(convs, map[string]interface{}{
			"id": id, "title": title, "updated_at": updated,
		})
	}
	return convs, rows.Err()
}

func (dbS *DbService) QueryHistory(convID string) []openai.ChatCompletionMessageParamUnion {
	rows, _ := dbS.GetdbInstance().Query("SELECT role, content FROM messages WHERE conversation_id = ? ORDER BY created_at ASC LIMIT 20", convID)
	var history []openai.ChatCompletionMessageParamUnion
	for rows.Next() {
		var role, content string
		rows.Scan(&role, &content)
		switch role {
		case "user":
			history = append(history, openai.UserMessage(content))
		case "assistant":
			history = append(history, openai.AssistantMessage(content))
		case "system":
			history = append(history, openai.SystemMessage(content))
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("遍历历史消息出错: %v", err)
		return nil // 或其他错误处理
	}
	return history
}
