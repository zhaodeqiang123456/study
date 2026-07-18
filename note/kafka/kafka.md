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
