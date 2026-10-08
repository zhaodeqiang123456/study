# AI Agent 后端系统 · 工程实战指南

> 记录从零构建 AI Agent 后端系统的完整过程，涵盖 WSL 环境搭建、服务启动、RAG 接入、向量检索、网络诊断等所有踩坑与解决方案。

---

## 目录

1. [项目架构概览](#1-项目架构概览)
2. [WSL 环境搭建](#2-wsl-环境搭建)
3. [国内镜像与下载优化](#3-国内镜像与下载优化)
4. [基础设施服务管理](#4-基础设施服务管理)
5. [RAG 系统构建](#5-rag-系统构建)
6. [Qdrant 向量数据库](#6-qdrant-向量数据库)
7. [Embedding 向量化](#7-embedding-向量化)
8. [Go 工程化实践](#8-go-工程化实践)
9. [网络诊断与修复](#9-网络诊断与修复)
10. [会话持久化与恢复](#10-会话持久化与恢复)

---

## 1. 项目架构概览

### 消息流

```
浏览器 (SSE) → Go HTTP Server (API) → Kafka → Consumer → DeepSeek LLM
                 ↑                        ↑           ↓
              Redis Cache              MySQL DB    RAG (Qdrant + Embedding)
```

### 目录结构

```
project/
├── main.go                  # HTTP 服务入口（端口 8080）
├── consumer/main.go         # Kafka 消费者（消费 task-queue）
├── config/
│   └── agent.md             # Agent 配置（YAML 格式：system_prompt、tools、model）
├── pkg/
│   ├── service.go           # Web 服务（路由、SSE、缓存）
│   ├── db.go                # MySQL 操作 + 对话历史查询
│   ├── task.go              # 任务状态管理
│   ├── cst.go               # 全局常量
│   ├── llm/
│   │   └── client.go        # LLM 调用（DeepSeek）、Embedding、Agent 循环
│   ├── agent/
│   │   ├── config.go        # agent.md YAML 解析
│   │   └── tools.go         # 工具执行 + MergeDeltaToolCalls
│   └── rag/
│       ├── rag.go           # 文本切片、Qdrant 入库/检索
│       └── documents/       # 知识库源文件（.txt）
├── static/
│   └── index.html           # 前端（多轮对话、SSE 流式、工具调用卡片）
└── scripts/
    ├── init_db.sql          # 数据库初始化 DDL
    ├── start.bat            # Windows 一键启动脚本
    └── test_qdrant.py       # Qdrant 诊断脚本
```

---

## 2. WSL 环境搭建

### 2.1 WSL 修复：从「灾难性故障」到正常运行

**故障现象**：运行 `wsl -d Ubuntu-26.04` 报错：
```
灾难性故障
错误代码: Wsl/Service/E_UNEXPECTED
```

**诊断过程**：

```bash
# 查看发行版状态
wsl -l -v
# 输出: Ubuntu-26.04  Stopped  2

# WSL 状态
wsl --status
# 确认发行版存在但无法启动
```

**尝试的修复方案**：

| 方案 | 结果 |
|------|------|
| `wsl --shutdown` 后重试 | ❌ 仍报 E_UNEXPECTED |
| `net stop wslservice` 重启服务 | ❌ 无权限 |
| `wsl --update` | ✅ E_UNEXPECTED 消失，但出现新错误 |

**新错误**：`Exec format error` — 内核启动了但无法执行 bash。

**根本原因**：注册表中 `BasePath = D:\tools` 指向的 `ext4.vhdx` 磁盘文件已丢失。

```bash
# 验证：查看注册表
reg query HKCU\Software\Microsoft\Windows\CurrentVersion\Lxss\{UUID} /v BasePath
# BasePath = D:\tools

# 检查 VHDX 文件
dir D:\tools\ext4.vhdx
# 文件不存在
```

**最终修复**：

```bash
# 1. 注销损坏的发行版
MSYS_NO_PATHCONV=1 PATH="" /c/Windows/System32/wsl.exe --unregister Ubuntu-26.04

# 2. wsl --install 走不通（GitHub raw 被墙，WSL 需要从 Microsoft Store 下载）
#    改为手动导入 rootfs

# 3. 从国内镜像下载 Ubuntu 26.04 base rootfs
curl -L -o /d/tools/ubuntu-26.04-rootfs.tar.gz \
  "https://mirrors.tuna.tsinghua.edu.cn/ubuntu-cdimage/ubuntu-base/releases/26.04/release/ubuntu-base-26.04-base-amd64.tar.gz"
# 33MB，2 秒下完

# 4. 导入 WSL
wsl --import Ubuntu-26.04 D:\wsl\ubuntu-26.04 D:\tools\ubuntu-26.04-rootfs.tar.gz

# 5. 启动测试
MSYS_NO_PATHCONV=1 PATH="" /c/Windows/System32/wsl.exe -d Ubuntu-26.04 -- uname -a
# Linux zdq 6.18.33.2-microsoft-standard-WSL2 x86_64 GNU/Linux ✅
```

### 2.2 MSYS 路径转换问题

**问题**：Git Bash 会自动将类 Unix 路径转换为 Windows 路径再传给 wsl.exe，导致大量 `Failed to translate` 错误。

**解决方案**：三件套缺一不可：

```bash
MSYS_NO_PATHCONV=1  # 阻止 MSYS 路径转换
PATH=""              # 清空 PATH 避免 Windows 路径混入 WSL
/c/Windows/System32/wsl.exe  # 使用 Windows 绝对路径
```

**完整命令模板**：

```bash
MSYS_NO_PATHCONV=1 PATH="" /c/Windows/System32/wsl.exe \
  -d Ubuntu-26.04 \                        # 发行版名
  -u zdq \                                 # 用户（root 需要权限时用 -u root）
  -- bash -c "你的命令"
```

### 2.3 Ubuntu base 极简版配置

`ubuntu-base` 是最小安装，需手动补齐组件：

```bash
# 1. 设置 apt 源（清华镜像）
sed -i 's|ports.ubuntu.com|mirrors.tuna.tsinghua.edu.cn|g' /etc/apt/sources.list.d/ubuntu.sources
apt update

# 2. 安装工具链
apt install -y golang-go git build-essential curl wget sudo
# golang-go 安装的版本可能较旧（1.26.0）

# 3. 安装特定版本 Go
# 从 golang.google.cn 下载官方二进制（go.dev 在国内被限速）
wget https://golang.google.cn/dl/go1.26.4.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.26.4.tar.gz
ln -sf /usr/local/go/bin/go /usr/local/bin/go

# 4. 创建用户
useradd -m -s /bin/bash zdq
echo 'zdq:你的密码' | chpasswd
echo 'zdq ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/zdq

# 5. 永久环境变量
echo 'export GOPROXY=https://goproxy.cn,direct' > /etc/profile.d/go.sh
echo 'export GOTOOLCHAIN=local' >> /etc/profile.d/go.sh
```

### 2.4 WSL 网络模式

**默认 NAT 模式**：WSL2 有独立 IP，通过 Windows 做 NAT 转发。问题在于 WSL2 的 IP 会变化，且某些情况下 localhost 转发不生效。

**推荐 mirrored 模式**（WSL 2.x+）：

创建 `%USERPROFILE%\.wslconfig`：

```ini
[wsl2]
networkingMode=mirrored   # WSL 与 Windows 共享网络栈，localhost 互通
memory=4GB
processors=4
```

```bash
# 修改后必须重启 WSL
wsl --shutdown
```

---

## 3. 国内镜像与下载优化

### 3.1 各下载源对照

| 用途 | 官方源（被墙/慢） | 国内镜像（快） |
|------|-------------------|----------------|
| Go 包 | `proxy.golang.org` | `goproxy.cn` |
| Go 二进制 | `go.dev/dl/` | `golang.google.cn/dl/` |
| apt 源 | `archive.ubuntu.com` | `mirrors.tuna.tsinghua.edu.cn` |
| Ubuntu rootfs | `cloud-images.ubuntu.com` | `mirrors.tuna.tsinghua.edu.cn/ubuntu-cdimage/` |
| pip 包 | `pypi.org` | `pypi.tuna.tsinghua.edu.cn/simple` |
| WSL 安装 | `raw.githubusercontent.com` (GitHub) | ❌ 不可达，只能手动导入 |

### 3.2 Go 安装的正确姿势

```
错误路径：apt install golang-go → Go 1.26.0（版本不对）
            go 自动下载 toolchain 1.26.4 → proxy.golang.org 超时

正确路径：golang.google.cn 下载 1.26.4 二进制 → 手动安装
           设置 GOTOOLCHAIN=local 禁止自动下载
```

```bash
# 永久配置（~/.profile 或 /etc/profile.d/go.sh）
export GOPROXY=https://goproxy.cn,direct
export GOTOOLCHAIN=local
export PATH=/usr/local/go/bin:$PATH
```

### 3.3 curl 下载超时处理

```bash
# 核心参数
curl -L \                    # 跟随重定向
     --connect-timeout 10 \  # 连接超时 10 秒
     --max-time 300 \        # 总超时 5 分钟
     -o output.tar.gz \      # 输出文件
     "URL"

# 断点续传（大文件）
curl -C - -L --max-time 600 -o output.tar.gz "URL"
```

---

## 4. 基础设施服务管理

### 4.1 MySQL 8.0

```bash
# 查看服务状态
sc query MySQL80

# 启动/停止
net start MySQL80
net stop MySQL80

# 连接（注意字符集）
mysql --default-character-set=utf8mb4 -uroot -p密码 -h127.0.0.1 -P3306

# 执行 SQL 文件
mysql --default-character-set=utf8mb4 -uroot -p密码 < init_db.sql
```

### 4.2 Redis

```bash
# 前台启动（调试用）
redis-server.exe redis.windows.conf

# 后台启动
redis-server.exe redis.windows.conf > /dev/null 2>&1 &

# 验证
redis-cli.exe ping   # 期望: PONG

# 检查端口
netstat -ano | grep ":6379"
```

**重要配置**：Windows 上 Redis 默认 `bind 127.0.0.1`，WSL2 无法访问。改为：

```conf
# redis.windows.conf
bind 0.0.0.0
```

### 4.3 Kafka (KRaft 模式)

Kafka 3.x+ 支持 KRaft，不再依赖 ZooKeeper。

```bash
# 1. 首次启动：格式化存储
kafka-storage.bat format \
  -t 7a3b2c1d-4e5f-6a7b-8c9d-0e1f2a3b4c5d \  # Cluster ID（UUID）
  -c config/server.properties

# 2. 启动 Kafka
kafka-server-start.bat config/server.properties > logs/server.log 2>&1 &

# 3. 验证
kafka-metadata-quorum.bat --bootstrap-server localhost:9092 describe --status

# 4. 创建 Topic
kafka-topics.bat --bootstrap-server localhost:9092 \
  --create --topic task-queue --partitions 3 --replication-factor 1

# 5. 查看 Topic
kafka-topics.bat --bootstrap-server localhost:9092 --list
```

**Key 配置** (`config/server.properties`)：

```properties
process.roles=broker,controller
node.id=1
controller.quorum.bootstrap.servers=localhost:9093
listeners=PLAINTEXT://:9092,CONTROLLER://:9093
advertised.listeners=PLAINTEXT://localhost:9092,CONTROLLER://localhost:9093
log.dirs=/tmp/kraft-combined-logs
```

### 4.4 Qdrant 向量数据库

```bash
# 启动
qdrant.exe > /dev/null 2>&1 &

# 健康检查（v1.18）
curl http://localhost:6333/telemetry    # REST API 健康状态
# /health 路径已移除，改用 /telemetry

# Dashboard
# 浏览器打开 http://localhost:6333/dashboard
```

### 4.5 查看各服务日志

```bash
# MySQL 日志
# C:\ProgramData\MySQL\MySQL Server 8.0\Data\*.err

# Kafka 日志
tail -f /d/tools/kafka/logs/server.log
# 或检查 topic 创建日志
kafka-topics.bat --bootstrap-server localhost:9092 --describe --topic task-queue

# Redis 日志
# redis-server.exe 默认输出到 stdout，需要重定向到文件查看
redis-server.exe redis.windows.conf > redis.log 2>&1 &

# Qdrant 日志
# 输出到 stdout，重定向到文件
qdrant.exe > qdrant.log 2>&1 &

# Go 服务日志
./server.exe > server.log 2>&1 &
tail -f server.log
```

---

## 5. RAG 系统构建

### 5.1 整体流程

```
文档(.txt) → 文本切片(tiktoken) → Embedding(词汇哈希) → Qdrant 入库 → 向量检索
                                                                          ↓
                                                              search_knowledge 工具 ← LLM
```

### 5.2 文本切片（tiktoken）

```go
func splitByTokens(text string, maxTokens int, overlap int) ([]string, error) {
    tke, _ := tiktoken.GetEncoding("cl100k_base")  // OpenAI 编码器
    tokens := tke.Encode(text, nil, nil)
    step := maxTokens - overlap   // 步长 = 500 - 50 = 450
    // 滑动窗口切片
    for start := 0; start < len(tokens); start += step {
        end := start + maxTokens
        chunks = append(chunks, tke.Decode(tokens[start:end]))
    }
    return chunks, nil
}
```

**参数设置**：
- `maxTokens=500`：每段最多 500 token
- `overlap=50`：相邻片段重叠 50 token（防止关键信息被切断）
- 步长 = 500 - 50 = 450

### 5.3 循环依赖解决：函数注入模式

**问题**：`agent` 包需要调用 `llm.GetEmbedding` 和 `rag.SearchByVector`，但 `llm` 包又引用 `agent` 包，形成循环引用。

```
agent ──→ llm ──→ agent   ❌ 循环依赖
```

**解决方案**：在 `agent` 包定义函数类型变量，由 `consumer/main.go` 在启动时注入：

```go
// pkg/agent/tools.go
var (
    EmbeddingFunc    func(string) ([]float32, error)
    VectorSearchFunc func([]float32, int) ([]string, error)
)

func searchKnowledgeTool(query string) (string, error) {
    queryVector, _ := EmbeddingFunc(query)
    docs, _ := VectorSearchFunc(queryVector, 3)
    return strings.Join(docs, "\n\n"), nil
}

// consumer/main.go
agent.EmbeddingFunc = llm.GetEmbedding
agent.VectorSearchFunc = rag.SearchByVector
```

### 5.4 工具结果如何注入 LLM 上下文

`searchKnowledgeTool` 返回的检索结果，最终由 agent 循环框架自动追加到 messages 中：

```go
// llm/client.go 中的 Agent 循环
result, _ := agent.ExecuteTool(tc.Function.Name, args)
messages = append(messages, openai.ToolMessage(result, tc.ID))
// LLM 在下一轮推理时看到工具结果，基于它生成回答
```

`agent.md` 中的 system_prompt 规则引导模型主动调用：

```yaml
system_prompt: |
  5. 当用户询问专业知识、技术问题或不确定的信息时，
     先调用 search_knowledge 工具检索知识库。
```

---

## 6. Qdrant 向量数据库

### 6.1 创建 Collection

```go
// pkg/service.go
func (s *Service) CreateCollection() error {
    client.CreateCollection(ctx, &qdrant.CreateCollection{
        CollectionName: "knowledge_base",
        VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
            Size:     1536,              // 向量维度（必须与 Embedding 输出一致）
            Distance: qdrant.Distance_Cosine,  // 余弦相似度
        }),
    })
}
```

### 6.2 写入 Point（HTTP REST API）

**踩坑 1：JSON 字段顺序**

Go 的 `map[string]interface{}` → `json.Marshal` 会按 key 字母序排列字段，导致：

```json
{"id":"xxx", "payload":{...}, "vector":[...]}  ← vector 在 payload 之后
```

Qdrant 1.18 会报错：`"Format error in JSON body: missing field ids at line 1 column XXXX"`

**解决方案**：必须使用 **struct** 而非 map，struct 的 JSON 字段顺序 = 定义顺序：

```go
type qdrantPoint struct {
    ID      string            `json:"id"`      // 第 1 个字段
    Vector  []float32         `json:"vector"`  // 第 2 个字段（必须在 payload 前）
    Payload map[string]string `json:"payload"` // 第 3 个字段
}
```

**踩坑 2：POST vs PUT**

Qdrant 的 upsert endpoint 必须用 **PUT** 方法，`http.Post()` 发的是 POST。

```go
// ❌ 错误
http.Post(url, "application/json", body)  // 发 POST 请求

// ✅ 正确
req, _ := http.NewRequest(http.MethodPut, url, body)
req.Header.Set("Content-Type", "application/json")
http.DefaultClient.Do(req)
```

### 6.3 向量检索

```go
// SearchByVector: HTTP POST /collections/{name}/points/search
func SearchByVector(queryVector []float32, limit int) ([]string, error) {
    body, _ := json.Marshal(map[string]interface{}{
        "vector":       queryVector,
        "limit":        limit,
        "with_payload": true,
    })
    resp, _ := http.Post(url, "application/json", bytes.NewReader(body))
    // 解析 result[].payload.text
}
```

### 6.4 Qdrant 诊断技巧

```python
# 查看 collection 信息
GET /collections/knowledge_base
# → points_count, indexed_vectors_count, vector size

# 滚动查看所有点
POST /collections/knowledge_base/points/scroll
{"limit": 10, "with_payload": true, "with_vector": false}

# 搜索
POST /collections/knowledge_base/points/search
{"vector": [...], "limit": 3, "with_payload": true}

# 删除点
POST /collections/knowledge_base/points/delete
{"points": ["id1", "id2"]}
```

---

## 7. Embedding 向量化

### 7.1 问题：两个远程 API 都不通

| API | 地址 | 结果 |
|-----|------|------|
| DeepSeek | `api.deepseek.com/v1/embeddings` | HTTP 404 |
| xiaoyiapi.xyz | `xiaoyiapi.xyz/v1/embeddings` | model_not_found (gpt-plus 组无 embedding 模型) |

### 7.2 务实的降级方案：本地词汇哈希 Embedding

当远程 Embedding API 不可用时，实现一个**可替换的本地替代**，让整个 RAG 管线先跑通：

```go
func GetEmbedding(text string) ([]float32, error) {
    const dim = 1536
    vector := make([]float32, dim)

    // 1. 分词：保留中英文和数字
    text = strings.ToLower(text)
    words := strings.FieldsFunc(text, func(r rune) bool {
        return !('a' <= r && r <= 'z' ||
            '0' <= r && r <= '9' ||
            r >= 0x4e00 && r <= 0x9fff)  // CJK 统一汉字区
    })

    // 2. 对每个词，用 FNV-1a 哈希确定向量位置并累加
    for _, word := range words {
        h := fnvHash(word)
        for j := uint32(0); j < 4; j++ {
            pos := int((h + j*0x9E3779B9) % dim)  // 黄金比例散列
            vector[pos] += 1.0
        }
    }

    // 3. L2 归一化（使向量长度 = 1，适配 Cosine 距离）
    var norm float64
    for _, v := range vector { norm += float64(v * v) }
    if norm > 1e-8 {
        inv := float32(1.0 / math.Sqrt(norm))
        for i := range vector { vector[i] *= inv }
    }

    return vector, nil
}
```

**设计要点**：

| 要点 | 说明 |
|------|------|
| 维度 1536 | 与 `text-embedding-3-small` 一致，Qdrant collection 创建时指定 |
| FNV-1a 哈希 | `offset=2166136261, prime=16777619` — 简单快速、分布均匀 |
| 黄金比例散列 | `0x9E3779B9` 确保每个词激活的 4 个位置分散 |
| L2 归一化 | 使 Cosine 距离计算有意义，向量长度 = 1 |
| 函数签名不变 | 返回 `([]float32, error)`，与真实 API 一致，替换时只改内部实现 |

**限制**：本地词汇哈希不考虑语义相似度，只做词汇重叠检测。后续接入真实 Embedding API 时替换即可。

---

## 8. Go 工程化实践

### 8.1 Kafka 消息幂等处理

```go
// pkg/db.go — SELECT ... FOR UPDATE 加锁 + 状态检查
func (dbS *DbService) CompleteTaskWithLog(task *Task) error {
    tx, _ := dbS.GetdbInstance().Begin()
    defer func() {
        if err != nil { tx.Rollback() } else { tx.Commit() }
    }()

    // FOR UPDATE 加排他锁
    var status string
    tx.QueryRow("SELECT status FROM tasks WHERE id = ? FOR UPDATE", task.ID).Scan(&status)
    if status == "done" {
        return nil  // 幂等：已处理则跳过
    }

    tx.Exec("UPDATE tasks SET status = 'done', result = ? WHERE id = ?", task.Result, task.ID)
    return nil
}
```

### 8.2 缓存穿透/击穿处理

```go
// pkg/service.go — Cache-Aside 模式 + 空值缓存
func (s *Service) getTaskWithCache(taskID string) (*Task, error) {
    // 1. 查 Redis
    cached, err := rdb.Get(ctx, "task:"+taskID).Result()
    if err == nil {
        json.Unmarshal([]byte(cached), &task)
        return &task, nil
    }

    // 2. 缓存未命中，查 MySQL
    task, err = dbService.GetTask(taskID)
    if err != nil {
        return nil, err
    }

    // 3. 回写 Redis（30s 过期）
    taskJSON, _ := json.Marshal(task)
    rdb.Set(ctx, "task:"+taskID, taskJSON, 30*time.Second)
    return &task, nil
}
```

### 8.3 context 泄漏的经典 bug

```go
// ❌ 错误：defer 在 for 循环内，永远不会执行
for {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    msg, err := reader.ReadMessage(ctx)
    defer cancel()  // 只有函数返回时才执行 → 内存泄漏
}

// ✅ 正确：立即释放
for {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    msg, err := reader.ReadMessage(ctx)
    cancel()  // 用完马上释放
}
```

### 8.4 硬编码路径 vs 相对路径

```go
// ❌ 硬编码（换机器就挂）
config.LoadConfig(`D:\study\projects\study\project\config\agent.md`)

// ✅ 相对路径 + 环境变量覆盖
configPath := os.Getenv("AGENT_CONFIG_PATH")
if configPath == "" {
    configPath = "config/agent.md"
}
config.LoadConfig(configPath)
```

### 8.5 优雅关闭

```go
// main.go
func main() {
    srv := pkg.NewService()

    go srv.Start()  // 后台启动

    // 等待信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 给 5 秒完成现有请求
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}
```

---

## 9. 网络诊断与修复

### 9.1 通用诊断命令

```bash
# 端口监听检查
netstat -ano | grep "LISTENING" | grep -E ":3306|:6379|:9092|:6333|:8080"

# 进程检查
tasklist | grep -i "redis\|qdrant\|java\|server\|consumer"

# 连通性测试
curl -v http://localhost:6333/telemetry
redis-cli.exe ping

# DNS 检查
cat /etc/resolv.conf

# WSL 网络
wsl hostname -I       # WSL IP
```

### 9.2 WSL2 ↔ Windows 网络修复过程

**现象**：WSL2 中 `curl localhost:6333` 超时。

**诊断**：

```bash
# WSL 中没有 eth0 接口 → 网络栈未正常初始化
ip addr show eth0  # 无输出

# DNS 服务器异常
cat /etc/resolv.conf  # nameserver 10.255.255.254
```

**修复步骤**：

1. 创建/修改 `.wslconfig`：`networkingMode=mirrored`
2. `wsl --shutdown` 重启
3. 验证：`curl http://localhost:6333/telemetry`

**端口绑定检查**：确认服务监听的 IP 地址

```bash
netstat -ano | grep ":6379"
# 127.0.0.1:6379  → 只有本机进程能访问（WSL 访问不到）
# 0.0.0.0:6379    → 所有网络接口可访问（WSL 能访问）
```

### 9.3 当 Windows 上服务正常但 WSL 访问不到时

```bash
# 1. 确认 Windows 上服务运行正常
curl http://localhost:6333/telemetry

# 2. 检查监听地址
netstat -ano | grep ":6333" | grep LISTENING
# 必须是 0.0.0.0 或 *:*，不能是 127.0.0.1

# 3. 如果是 127.0.0.1，修改服务配置后重启
# 4. 检查防火墙是否阻止
# 5. 确认 .wslconfig 中有 networkingMode=mirrored 并已重启 WSL
```

---

## 10. 会话持久化与恢复

### 10.1 约定

每次会话结束后保存两样东西：

1. **会话记录** `docs/sessions/YYYY-MM-DD-NN.md`
2. **恢复指针** `docs/sessions/.resume`（一行，最新会话路径）

### 10.2 会话记录模板

```markdown
# 会话记录 · 2026-07-28

## 会话概要
## 代码变更
| 文件 | 说明 |
|------|------|

## 当前项目状态
| 服务 | 端口 | 状态 |
|------|------|------|

## 关键技术决策
## 待完成
```

### 10.3 持久记忆

项目级记忆存储在 `C:\Users\32657\.claude\projects\D--study-study\memory\`：

```markdown
<!-- MEMORY.md — 记忆索引 -->
- [Student Profile](student-profile.md) — 学生背景与技术栈
- [Run in WSL](run-in-wsl.md) — WSL 执行规范
- ...每个文件一个事实
```

---

## 附录：常用命令速查

### 服务启动（Windows Git Bash）

```bash
# Redis
/d/tools/Redis/redis-server.exe /d/tools/Redis/redis.windows.conf > /dev/null 2>&1 &

# Qdrant
/d/tools/qdrant/qdrant.exe > /dev/null 2>&1 &

# Kafka
/d/tools/kafka/bin/windows/kafka-server-start.bat /d/tools/kafka/config/server.properties > /d/tools/kafka/logs/server.log 2>&1 &

# MySQL (系统服务)
net start MySQL80
```

### WSL 编译与运行

```bash
# 编译
MSYS_NO_PATHCONV=1 PATH="" /c/Windows/System32/wsl.exe \
  -d Ubuntu-26.04 -u zdq -- \
  bash -c "cd /mnt/d/study/study/project && go build ./..."

# 运行 consumer
MSYS_NO_PATHCONV=1 PATH="" /c/Windows/System32/wsl.exe \
  -d Ubuntu-26.04 -u zdq -- \
  bash -c "
    export GOPROXY=https://goproxy.cn,direct GOTOOLCHAIN=local
    export PATH=/usr/local/go/bin:\$PATH
    cd /mnt/d/study/study/project && go run ./consumer/
  "
```

### Qdrant 操作

```bash
# 查看 collection
curl http://localhost:6333/collections/knowledge_base | python3 -m json.tool

# 搜索
curl -X POST http://localhost:6333/collections/knowledge_base/points/search \
  -H 'Content-Type: application/json' \
  -d '{"vector":[...], "limit":3, "with_payload":true}'

# 删除所有点
curl -X POST http://localhost:6333/collections/knowledge_base/points/delete \
  -H 'Content-Type: application/json' \
  -d '{"filter":{}}'
```

### Go 开发

```bash
# 设置代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOTOOLCHAIN=local

# 查看环境
go env GOPROXY GOTOOLCHAIN GOROOT GOPATH

# 编译
go build -o output.exe .              # 主程序
go build -o output.exe ./consumer/     # consumer

# 运行
go run .               # 主程序
go run ./consumer/     # consumer

# 清理缓存
go clean -modcache
```
