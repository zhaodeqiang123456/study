# 第一周：后端基础与可靠任务闭环

## 周目标

本周不升级 Embedding、不做语义缓存。先掌握并应用 HTTP 请求生命周期、context、SQL 事务、Kafka 投递语义、Redis/SSE、Go 接口和测试，把当前项目从“能跑”推进到“失败时行为可解释”。

建议每天 2～3 小时：1 小时学习，1～1.5 小时编码，最后 15 分钟记录问题和结论。

## Day 1：HTTP 与 context

学习：Handler 生命周期、`ResponseWriter` 使用边界、context cancellation/deadline、goroutine 退出条件。

编码：移除 `/task` Handler 中写响应的 goroutine；响应必须在 Handler 返回前完成。为数据库写入和 Kafka 投递分别传递请求 context，并定义超时。

验收：客户端能稳定收到 `202 + task_id`；数据库失败和 Kafka 失败返回明确状态；Handler 返回后没有后台写响应。

## Day 2：MySQL 事务与状态机

学习：事务边界、提交失败、`SELECT ... FOR UPDATE`、唯一约束、状态机和幂等键。

编码：把任务状态定义为 `pending/processing/done/failed`；实现 `ClaimTask`，只允许一个消费者从 `pending` 抢到 `processing`；补齐 `Rollback` 和 `Commit` 错误处理。

验收：两个并发消费者处理同一 task ID 时只有一个能进入 LLM；已有 `done` 任务不会再次执行。

## Day 3：Kafka 消费语义

学习：at-most-once、at-least-once、`ReadMessage` 与 `FetchMessage + CommitMessages`、重试和 DLQ。

编码：改用 `FetchMessage`；只有任务结果和消息记录成功后才提交 offset；失败进入有限重试或 DLQ。先接受 at-least-once，再用幂等保证业务效果。

验收：在“LLM 前、LLM 后、数据库提交后”分别终止 consumer，重启后任务不丢；重复消息不会重复调用 LLM。

## Day 4：Redis、SSE 与取消

学习：Cache-Aside、穿透/击穿/雪崩区别、Redis List/Stream、SSE 断开检测和背压。

编码：SSE 循环使用 `r.Context().Done()` 退出；Redis 操作不使用已取消的 consumer context；为任务结果增加明确 TTL 和错误分类。暂不实现语义缓存。

验收：浏览器断开后下游请求可以取消；Redis 不可用时接口有降级行为；无无限循环和无限增长的 stream key。

## Day 5：Go 接口与单元测试

学习：依赖倒置、接口隔离、表驱动测试、fake/mock、竞态检测。

编码：抽象 `TaskRepository`、`LLMClient`、`EventSink`、`MessageConsumer`；为状态机、重复任务、工具参数错误和配置校验写测试。

验收：核心测试不依赖真实 MySQL、Kafka、Redis、Qdrant；`go test -race ./...` 通过。

## Day 6：集成改造

编码：把 Day 1～5 的实现接回现有 consumer；增加优雅关闭，停止接收新消息，等待处理中任务，关闭 Kafka/Redis/DB 客户端。

验收：启动两个 consumer 实例并发消费；模拟 LLM 超时、Redis 断开、Kafka 重启；每种故障都有日志、状态和恢复结果。

## Day 7：复盘与输出

完成：画一张 API → Kafka → Claim → LLM/Tool → DB → Commit → SSE 的时序图；补 README 的运行方式、故障模型和测试命令；写一页 STAR 项目描述。

闭卷回答：

1. 为什么 offset 不能和 MySQL 事务天然原子提交？
2. 为什么 `FOR UPDATE` 放在 LLM 调用之后没有幂等价值？
3. Handler 返回后继续写 `ResponseWriter` 会发生什么？
4. SSE 客户端断开后，服务端如何取消下游调用？
5. 如何证明系统是 at-least-once，但业务效果接近 exactly-once？

## 本周完成标准

- 代码：任务状态机、手动提交 offset、前置幂等、context 传播、优雅关闭。
- 测试：至少 8 个核心测试，覆盖重复消息、超时、失败重试和进程恢复。
- 工程：`gofmt`、`go vet ./...`、`go test -race ./...` 全部通过。
- 文档：时序图、故障矩阵、运行命令、一次故障复盘。

未达到以上标准前，不进入真实 Embedding、语义缓存和复杂 Agent 优化。这样可以先建立后端基础，再让每一项 AI 能力都有可靠的运行底座。
