package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" // 记得导入驱动
)

type Exec int

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

func initDB() (db *sql.DB) {
	var err error
	db, err = sql.Open("mysql", "root:000922@tcp(127.0.0.1:3306)/mysql_learn")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("无法连接到数据库:", err)
	}
	fmt.Println("数据库连接成功！")
	return
}

func sql_exec(exec Exec, db *sql.DB) func(*Task) {
	switch exec {
	case Insert:
		return func(task *Task) {
			_, err := db.Exec("INSERT INTO tasks (id, status, result, created_at) VALUES (?, ?, ?, NOW())", task.ID, task.Status, task.Result)
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

func getTaskDetail(db *sql.DB, taskID string) (*TaskDetail, error) {
	var detail TaskDetail
	err := db.QueryRow("SELECT id, status FROM tasks WHERE id = ?", taskID).Scan(&detail.ID, &detail.Status)
	if err != nil {
		return nil, err
	}

	// 2. 查关联日志（利用索引）
	rows, err := db.Query("SELECT log_msg FROM task_logs WHERE task_id = ? ORDER BY created_at", taskID)

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
