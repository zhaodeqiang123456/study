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

#### 表结构设计

数据库三范式(减少数据冗余，保证数据一致性)

1. 第一范式(1NF) 列不可再分,每个字段都是原子值
2. 第二范式(2NF) 在满足第一范式的前提下, 非主键列完全依赖于主键,**消除部分依赖**
3. 第三范式(3NF) 在满足第二范式的前提下, 非主键列不能依赖与其他非主键列, **消除传递依赖**
> 实际开发中，为了减少多表 JOIN 提高性能，常会故意冗余一些字段。比如在订单表里直接存 customer_name，哪怕客户表改了名，订单历史也不会变。这是一种用空间换时间的权衡

#### 索引设计策略
**适合建索引的列**: WHERE、JOIN、ORDER BY 中用到的列；区分度高的列（如用户ID）；外键列
**不适合建索引的列**: 频繁更新的列；区分度很低的列（如性别）；TEXT/BLOB 全字段索引（应用前缀索引）


#### 事务与隔离级别
> 事务是一组要么全部成功、要么全部失败回滚的 SQL 操作。用 START TRANSACTION 开始，COMMIT 提交，ROLLBACK 回滚。

##### ACID 四大特性
A原子性 (Atomicity)：事务是不可分割的最小单元，要么全做要么全不做。

C一致性 (Consistency)：事务执行前后，数据都满足业务定义的规则约束（如金额总和不变）。

I隔离性 (Isolation)：多个事务并发执行时，相互之间不应有干扰。由隔离级别保证。

D持久性 (Durability)：事务一旦提交，对数据的修改就是永久的。

##### 并发事务引发的问题
假设有两个事务 T1、T2 同时运行，可能出现：

**脏读 (Dirty Read)**：T1 读到了 T2 未提交的修改。如果 T2 回滚，T1 读到的就是无效数据。

**不可重复读 (Non-repeatable Read)**：T1 内两次读取同一行数据，读到的值却不一样（因为 T2 在这中间修改并提交了这行数据）。

**幻读 (Phantom Read)**：T1 内两次查询同一范围的数据，发现多了几行（因为 T2 在这中间插入了新行并提交）。注意与不可重复读的区别：不可重复读是针对同一行的修改，幻读是针对新增/删除的行。

> MySQL 默认的 REPEATABLE READ 已经解决了脏读和不可重复读，但在某些情况下仍可能出现幻读。面试常问如何防止幻读，答案是使用临键锁（Next-Key Lock），即行锁 + 间隙锁，锁定一个范围。


