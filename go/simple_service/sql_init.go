package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // 记得导入驱动
)

type Exec int

type dbService struct {
	db *sql.DB
}

const (
	Insert Exec = iota + 1
	Delete
	Query
	Alter
)

type TaskDetail struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Logs   []string `json:"logs"`
}

func (dbS *dbService) initDB() {
	var err error
	dbS.db, err = sql.Open("mysql", "root:000922@tcp(127.0.0.1:3306)/mysql_learn")
	if err != nil {
		log.Fatal(err)
	}
	if err = dbS.db.Ping(); err != nil {
		log.Fatal("无法连接到数据库:", err)
	}
	fmt.Println("数据库连接成功！")
	return
}

func (dbS *dbService) sql_exec(exec Exec) func(*Task) {
	switch exec {
	case Insert:
		return func(task *Task) {
			_, err := dbS.db.Exec("INSERT INTO tasks (id, status, result, created_at) VALUES (?, ?, ?, NOW())", task.ID, task.Status, task.Result)
			if err != nil {
				log.Fatal("插入失败:", err)
			}
			fmt.Println("任务插入成功！")
		}
	// 其他 case 同理
	default:
		return nil
	}
}

func (dbS *dbService) getTaskDetail(taskID string) (*TaskDetail, error) {
	var detail TaskDetail
	err := dbS.db.QueryRow("SELECT id, status FROM tasks WHERE id = ?", taskID).Scan(&detail.ID, &detail.Status)
	if err != nil {
		return nil, err
	}

	// 2. 查关联日志（利用索引）
	rows, err := dbS.db.Query("SELECT log_msg FROM task_logs WHERE task_id = ? ORDER BY created_at", taskID)

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
func (dbS *dbService) completeTaskWithLog(taskID string) error {
	tx, err := dbS.db.Begin() // 开启事务
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

	// 操作1：更新任务状态
	_, err = tx.Exec("UPDATE tasks SET status = 'done', result = 'success' WHERE id = ?", taskID)
	if err != nil {
		return err
	}

	// 操作2：插入日志
	_, err = tx.Exec("INSERT INTO task_logs (task_id, log_msg, created_at) VALUES (?, ?, NOW())", taskID, "任务已自动完成")
	if err != nil {
		return err
	}

	return nil
}
