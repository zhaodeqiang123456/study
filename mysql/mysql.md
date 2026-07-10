# mysql

#### 启动mysql 服务

1. 基于window的服务管理services.msc 找到mysql的服务进程（本地服务器）
2. 通过navicat 或者 其他客户端连接到服务器
3. 主机地址IP, 端口号（用于唯一标识一台网络中的终端）
4. 用户名密码
   
#### 创建数据库
在mysql中创建数据库也叫做创建Schema

```
CREATE SCHEMA `mysql_learn` 
```

#### 创建表
```
CREATE TABLE `mysql_learn`.`new_table` (
`id` varchar(64) not null,
`status` varchar(20) not null,
`result` text(64) null,
`created_at` datetime(6) not null,
primary key (`id`)
)
engine = InnoDB;

// create table 创建表
// varchar text datetime 存储的数据的基本类型
// not null, null, primary key 表示对字段的约束
```

#### CRUD 增删查改

```
# 新增数据
INSERT INTO table (字段1， 字段2，，，) VALUES ('values'， 'values', , , );

INSERT INTO tasks (id, status, result, created_at) VALUES ('task-001', 'pending', NULL, NOW());

# 查询数据
SELECT * FROM tasks;
SELECT * FROM tasks WHERE status = 'pending';

# 修改数据
UPDATE tasks
SET status = 'done', result = '任务完成'
WHERE id = 'task-001';

# 删除数据
DELETE FROM tasks WHERE id = 'task-002';

```

#### 常用数据类型
1. INT, BIGINT: 整数
2. VARCHAR(n): 变长字符串
3. TEXT: 长文本
4. DATETIME, TIMESTAMP: 日期时间
5. DECIMAL: 精确小数（用于金额）


#### go项目连接mysql以及操作

1. 在go 项目的根目录下安装驱动 go get github.com/go-sql-driver/mysql
2. 连接数据库 db, err := sql.Open("mysql", "用户名:密码@tcp(IP:端口号)/数据库")
3. CRUD操作   _, err = db.Exec("INSERT INTO tasks (id, status, result, created_at) VALUES (?, ?, ?, NOW())", x, y, z)

