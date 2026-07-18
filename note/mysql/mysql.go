package mysql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Task struct {
	id     string
	status string
}

func note() {

	// 连接数据库
	db, _ := sql.Open("mysql", "root:password@tcp(ip:port)/my_task_db")
	db.Ping()

	tk := &Task{
		id:     "task01",
		status: "pending",
	}

	// 插入数据(参数化查询)
	db.Exec("Insert into tasks (id, status, created_at) values (?, ?, now())", &tk.id, &tk.status)

	// 查询单行
	db.QueryRow("select status from tasks where id = ?", tk.id).Scan(&tk.status)

	// 查询多行
	rows, _ := db.Query("select id, status from tasks where status = ?", &tk.status)
	defer rows.Close()
	for rows.Next() {
		var task = &Task{}
		rows.Scan(&task.id, &task.status)
	}

	// 更新与删除
	db.Exec("update tasks set status = 'done' where id = ?", tk.id)
	db.Exec("delete from tasks where id = ?", tk.id)

	// 事务
	tx, _ := db.Begin()
	tx.Exec("UPDATE tasks SET status = 'done' WHERE id = ?", "task-1")
	tx.Exec("INSERT INTO task_logs (task_id, log_msg, created_at) VALUES (?, ?, NOW())", "task-1", "完成")
	tx.Commit()
	// 出错时调用 tx.Rollback()

	// 使用 EXPLAIN 查看执行计划
	// EXPLAIN SELECT * FROM tasks WHERE status = 'pending';
}
