# 2026-08-01 知识归档：Handler 生命周期、Context 与滑动窗口

来源：[`learning/lessons/2026-08-01.md`](../lessons/2026-08-01.md)；训练答案：[`learning/training/2026-08-01.md`](../training/2026-08-01.md)。

## 可复用结论

- `ResponseWriter` 的唯一所有者应是 handler；handler 返回后任何后台 goroutine 都不得继续写响应。
- `go test -race` 只能发现已执行路径上的内存竞争，不能证明 HTTP 响应协议不变量或外部消息是否已提交。
- `r.Context()` 传播取消和 deadline，但不会自动回收 goroutine，也不会回滚已发生的 Kafka/DB 副作用。
- 请求内工作使用请求 context；脱离请求的异步任务由服务级 worker、队列和服务生命周期 context 管理。
- 并发测试用 channel 屏障建立确定事件顺序，不用 `time.Sleep`；测试替身自身也必须避免 race。
- 滑动窗口不变量：频次表对应闭区间 `[left,right]`，`distinct` 等于正频次键数量；左右指针总移动次数各不超过 `n`，因此为 `O(n)`。
- 可控 fake queue = 内存中的“同步虚拟沙盘”，它允许测试代码随心所欲地操控错误、延迟和积压，用来验证业务代码在极端异常（超时、断网、重复消息）下的表现，同时跑通 -race 检测，无需启动笨重的真实 MQ 容器。
- 
## 已知失败模式

- Kafka 失败路径写 `500`，后台 DB 成功路径随后写 `202`。
- fake queue 取消测试通过，被误解为真实 Kafka 没有产生消息。
- goroutine “最终退出一次”被误解为不存在泄漏。
- 只凭样例测试宣称全输入正确或只凭嵌套循环外观宣称 `O(n²)`。

## 后续验证入口

把 Queue/Repository 抽成消费者接口，先完成“阻塞时取消”的同步屏障测试；再补真实 `_test.go`、`go test` 与可用时的 `go test -race` 输出。
