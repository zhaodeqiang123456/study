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

func main() {
	db := initDB()
	defer db.Close()

	task := Task{
		ID:     "task-go-002",
		Status: "processing",
	}
	f := sql_exec(Insert, db)
	f(&task)
}
