#### Kafka 启动流程

##### 前置条件

1. 安装 jdk17+, 依赖Java编译环境
2. kafka 安装并解压

##### 生成集群uuid

1. cd kafka 目录
2. 运行 .\bin\windows\kafka-storage.bat random-uuid
3. 复制输出的uuid 例如 5l6dh7Y0Rri5\_LMyr64JAw

##### 格式化存储目录 （单机模式）

.\bin\windows\kafka-storage.bat format -t **UUID** -c config\server.properties --standalone

成功后输出Formatting

##### 启动Kafka服务

.\bin\windows\kafka-server-start.bat config\server.properties

等待出现 `[KafkaServer id=1] started`，表示启动成功。**此窗口保持打开**。

##### 创建topic （新开一个终端）

.\bin\windows\kafka-topics.bat --create --topic task-queue --bootstrap-server localhost:9092

##### kafka 操作

```go
// 生产者
import "github.com/segmentio/kafka-go"
w := &kafka.Writer{
    Addr:  kafka.TCP("localhost:9092"),
    Topic: "task-queue",
}
w.WriteMessages(context.Background(),
    kafka.Message{Key: []byte("task-1"), Value: []byte(`{"id":"task-1"}`)},
)
```

~~~go
// 消费者
r := kafka.NewReader(kafka.ReaderConfig{
    Brokers: []string{"localhost:9092"},
    Topic:   "task-queue",
    GroupID: "task-consumer-group",
})
msg, _ := r.ReadMessage(context.Background())
fmt.Println(string(msg.Value))  // {"id":"task-1"}
~~~

**生产者端负责将消息,数据推送到Kafka服务端， 消费者端负责将数据从Kafka端拉出**


---

### kafka 消费者幂等性

> 应用场景: 由于Kafka和消费者是两个服务端,  一条被推入Kafka的消息,   消费者端消费消息后，需要将信息同步到Kafka, 如果在信息同步之前， 消费者端异常退出，造成的结果---  数据的crud已经在数据库体现，但是消息仍然存在在Kafka队列中， 该消息可能会被消费端多次拉出，重复消费
>
> 目标：防止同一条消息被重复消费，导致重复更新数据库


实现方式:  状态 + 悲观锁，   在消费者处理任务之前，先检查数据库状态，如果任务已经是终态（如 `done`），则直接跳过，不做任何修改。

查询时， 使用 `SELECT ... FOR UPDATE` 防止并发问题。`SELECT ... FOR UPDATE` 在事务内会对这行加排他锁，防止其他消费者同时处理同一任务。


```go
// 在 consumer 的 processTask 函数中
func processTask(taskID string) error {
    // 开启事务
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            tx.Rollback()
        } else {
            tx.Commit()
        }
    }()

    // 查询任务状态，加行锁
    var status string
    err = tx.QueryRow("SELECT status FROM tasks WHERE id = ? FOR UPDATE", taskID).Scan(&status)
    if err != nil {
        return err
    }
    if status == "done" {
        log.Printf("任务 %s 已经处理过，跳过", taskID)
        return nil // 幂等性：已经 done 了，直接返回
    }

    // 模拟处理任务
    time.Sleep(3 * time.Second)

    // 更新为 done
    _, err = tx.Exec("UPDATE tasks SET status = 'done', result = '处理成功' WHERE id = ?", taskID)
    if err != nil {
        return err
    }

    // 删除 Redis 缓存，让下次查询强制回源
    rdb.Del(ctx, "task:"+taskID)

    return nil
}
```
