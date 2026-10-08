# AI Agent 后端项目总结

## 一、项目概述

本项目从零开始，使用 **Go 语言** 构建了一个 AI Agent 后端系统，逐步集成了 **多轮对话、RAG（检索增强生成）、流式推送（SSE）、工具调用（Tool Calling）、配置化管理** 等核心能力。整个开发过程严格遵循“先跑通核心链路，再逐步加固”的原则，最终形成一个可演示、可扩展的智能体服务原型。

## 二、技术栈全景

### 编程语言

- **Go 1.18**（主要后端语言，高并发、高性能）

### 后端框架与库

- `net/http`：构建 HTTP API 与 SSE 流式接口
- `openai-go` v1.12.0：调用 DeepSeek API（对话与工具调用）
- `github.com/segmentio/kafka-go`：Kafka 生产者与消费者
- `github.com/redis/go-redis/v9`：缓存、限流、流式数据暂存
- `database/sql` + `go-sql-driver/mysql`：MySQL 业务数据持久化
- `github.com/google/uuid`：生成唯一标识
- `gopkg.in/yaml.v3`：解析 `agent.md` 配置文件

### 消息队列与存储

- **Kafka**：异步任务解耦（生产者-消费者模式）
- **MySQL**：业务数据（任务、会话、消息记录）
- **Redis**：缓存、限流计数器、SSE 流式数据列表
- **Qdrant**（已部署）：向量数据库，为语义检索做准备

### 大模型与 AI

- **DeepSeek 对话模型**（`deepseek-chat`）：实现推理、工具调用
- **OpenAI Embedding**（备用方案）：文本向量化
- **RAG**：基于文档分块 + 关键词检索（计划升级为向量检索）

### 前端

- 原生 HTML/CSS/JS：单页对话界面，支持多轮会话、流式展示、工具调用卡片

### 配置与规范

- **`agent.md`**：Agent 行为配置文件（YAML + Markdown 混合格式），定义系统 Prompt、工具列表、模型参数等

### 工具链

- Git/GitHub：版本控制
- Docker（学习阶段尝试，后改用本地部署）

## 三、项目架构演进

### 阶段一：基础后端 Demo

**目标**：跑通 Go HTTP 服务 + MySQL + Redis + Kafka 全链路**架构**：客户端 → API → Kafka → Consumer → MySQL/Redis**产出**：

- 异步任务提交与轮询查询结果
- 幂等性、优雅关闭、超时控制等基础加固

### 阶段二：AI 推理接入 + 多轮对话

**新增功能**：

- 消费者调用 DeepSeek API 进行推理
- 数据库新增 `conversations`、`messages` 表，实现多轮对话
- 前端支持会话切换、历史消息加载

### 阶段三：流式输出 + RAG（检索增强生成）

**新增功能**：

- SSE 流式推送，用户可实时查看模型逐字输出
- RAG 流程：文档分块 → 向量化（暂时使用关键词匹配代替 Embedding）→ 检索相关片段 → 拼入 Prompt
- Qdrant 向量数据库部署与 HTTP API 对接

### 阶段四：Agent 工具调用

**新增功能**：

- 实现 Tool Calling 循环，Agent 可主动调用 `calculator`、`get_weather` 等工具
- SSE 事件类型扩展为 `text_delta`、`tool_call`、`tool_result`、`done`，前端可视化展示工具执行过程
- 处理流式 API 返回的 `DeltaToolCall` 分片合并与类型转换

### 阶段五：配置化与加固

**新增功能**：

- 引入 `agent.md` 配置文件，动态加载系统 Prompt、工具列表、模型参数
- 代码与配置解耦，便于切换不同角色、不同工具集的 Agent
- 计划实施：全链路追踪、语义缓存、输出校验、安全护栏

## 四、关键技术难点与解决方案

### 1. Kafka 幂等性

- **问题**：消息重复消费导致任务重复处理
- **解决**：消费者处理前通过 `SELECT ... FOR UPDATE` 检查任务状态，若已为终态则跳过

### 2. Redis 缓存穿透/击穿

- **问题**：大量请求查询不存在或过期的数据，直接打到数据库
- **解决**：空值缓存防穿透；使用互斥锁实现 Cache-Aside 模式防击穿

### 3. SSE 超时与 context 使用

- **问题**：Redis 读取流式数据时遇到 `context deadline exceeded`
- **解决**：调整 context 超时策略，对 SSE 推送采用短轮询 + 小延迟替代长时间阻塞

### 4. DeepSeek Embedding 不可用

- **问题**：`/v1/embeddings` 返回 404
- **解决**：临时降级为**关键词匹配**，同时准备接入 OpenAI Embedding


### 5. Agent 工具调用重复

- **问题**：计算器被多次调用（不同格式的表达式）
- **原因**：工具返回格式不稳定，Prompt 未禁止重复调用
- **解决**：优化工具返回 JSON 结构化数据，并在系统 Prompt 中明确“工具结果准确，勿重复调用”

### 6. 前端无法显示 Agent 工具步骤

- **问题**：SSE 只推送最终文本，中间工具调用不可见
- **解决**：重新设计 SSE 事件类型（`text_delta`、`tool_call`、`tool_result`、`done`），前端分类型渲染

## 五、项目当前状态
多轮推理+流式输出：暂时移除rag检索增强生成模块，并将该部分移入tool calling中供LLM自行决定是否使用
