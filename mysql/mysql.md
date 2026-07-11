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

---
### sql进阶
#### 多表关联  JOIN

```
# inner join 只保留能匹配的行

SELECT t.id, t.status, l.log_msg
FROM tasks t
INNER JOIN task_logs l ON t.id = l.task_id;


# left join 保留左表全部行，右表匹配不上的字段则填null

SELECT t.id, t.status, l.log_msg
FROM tasks t
LEFT JOIN task_logs l ON t.id = l.task_id;

```
#### 聚合与分组  group by
```
# 统计各status组COUNT(*)
SELECT status, COUNT(*) AS cnt
FROM tasks
GROUP BY status; 

# 约束, having
SELECT status, COUNT(*) AS cnt
FROM tasks
GROUP BY status
HAVING cnt > 1;


# order by排序，以及输出限制 limit
SELECT status, COUNT(*) AS cnt
FROM tasks
GROUP BY status
ORDER BY cnt DESC
LIMIT 3;


# 子查询，嵌套查询
SELECT * FROM tasks
WHERE created_at = (SELECT MAX(created_at) FROM task_logs);
```

#### 索引与执行计划

```
# 创建索引
-- 为 tasks 表的 status 字段创建普通索引
CREATE INDEX idx_status ON tasks(status);

-- 为 task_logs 表的 task_id 创建索引（常用于关联查询）
CREATE INDEX idx_task_id ON task_logs(task_id);

# explain 查看执行计划，分析查询效率
type 表示查询的方式， const 主键或者唯一索引查询， ref 普通索引查询， all 全表扫描
key 表示实际查询使用的索引 null表示此次查询未用索引

```

#### 重点总结
##### 查询进阶
1. order by 排序
2. limit 分页处理
3. as 设置别名，简化表和字段名称
4. join 连接查询
5. count(), sum(), avg(), max(), min(), 聚合函数 通常搭配group by做分组统计

##### 索引 优化查询效率
常用索引类型
1. primary key  主键索引，唯一且非空
2. unique 唯一索引，值不能重复
3. index 普通索引，只为加速查询
4. fulltext  全文索引，用于文本搜索
执行计划 explain 查看MySQL如何执行查询，用于sql优化

