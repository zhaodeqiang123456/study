package pkg

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql" // 记得导入驱动
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

func (dbS *DbService) getdbInstance() *sql.DB {
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
	err := dbS.getdbInstance().QueryRow("SELECT id, status FROM tasks WHERE id = ?", taskID).Scan(&detail.ID, &detail.Status)
	if err != nil {
		return nil, err
	}

	// 2. 查关联日志（利用索引）
	rows, err := dbS.getdbInstance().Query("SELECT log_msg FROM task_logs WHERE task_id = ? ORDER BY created_at", taskID)

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
	tx, err := dbS.getdbInstance().Begin() // 开启事务
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
	_, err := dbS.getdbInstance().Exec("insert into tasks (id, status, result, created_at) values (?, ?, ?, now())", task.ID, task.Status, task.Result)
	return err
}

func (dbS *DbService) GetTask(taskID string) (Task, error) {
	var task Task
	err := dbS.getdbInstance().QueryRow("select id, status, result from tasks where id = ?", taskID).Scan(&task.ID, &task.Status, &task.Result)
	return task, err
}
