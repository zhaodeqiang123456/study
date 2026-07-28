package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"simple_service/pkg/llm"
	"time"

	"github.com/google/uuid"
	"github.com/pkoukk/tiktoken-go"
)

// ==================== 1. 文本切片 ====================
func splitByTokens(text string, maxTokens int, overlap int) ([]string, error) {
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("获取编码器失败: %w", err)
	}
	tokens := tke.Encode(text, nil, nil)
	step := maxTokens - overlap
	if step <= 0 {
		step = maxTokens
	}
	var chunks []string
	for start := 0; start < len(tokens); start += step {
		end := start + maxTokens
		if end > len(tokens) {
			end = len(tokens)
		}
		chunkTokens := tokens[start:end]
		chunks = append(chunks, tke.Decode(chunkTokens))
	}
	return chunks, nil
}

// ==================== 2. Embedding 调用（已移至 pkg/llm） ====================

// ==================== 3. 写入 Qdrant（HTTP，结构体保证字段顺序） ====================
const qdrantBaseURL = "http://localhost:6333"

// Qdrant 1.18 对 JSON 字段顺序敏感，必须用 struct 而非 map
type qdrantPoint struct {
	ID      string            `json:"id"`
	Vector  []float32         `json:"vector"`
	Payload map[string]string `json:"payload"`
}

type upsertRequest struct {
	Points []qdrantPoint `json:"points"`
}

// upsertPoint 单点写入 Qdrant
func upsertPoint(id string, vector []float32, payload map[string]string) error {
	url := fmt.Sprintf("%s/collections/knowledge_base/points?wait=true", qdrantBaseURL)
	body, err := json.Marshal(upsertRequest{
		Points: []qdrantPoint{{ID: id, Vector: vector, Payload: payload}},
	})
	if err != nil {
		return fmt.Errorf("序列化点失败: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Qdrant 返回错误 %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// ==================== 4. 加载知识库主函数 ====================
func LoadKnowledgeBase() ([]string, error) {
	dir := "pkg/rag/documents" // 基于消费者工作目录的相对路径
	log.Printf("[RAG] 文档目录: %s", dir)
	files, err := filepath.Glob(filepath.Join(dir, "*.txt"))
	if err != nil {
		return nil, fmt.Errorf("查找文档失败: %w", err)
	}
	if len(files) == 0 {
		log.Println("[RAG] 没有找到任何文档，跳过入库")
		return nil, nil
	}
	var documentChunks []string
	for _, file := range files {
		log.Printf("[RAG] 正在处理文件: %s", file)
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("[RAG] 读取文件 %s 失败: %v", file, err)
			continue
		}

		// 切片：每个片段 500 token，重叠 50
		chunks, err := splitByTokens(string(content), 500, 50)
		if err != nil {
			log.Printf("[RAG] 切片失败 %s: %v", file, err)
			continue
		}

		// ===== 接入向量数据库：Embedding → Qdrant 逐点入库 =====
		pointCount := 0
		for _, chunk := range chunks {
			// 调用 Embedding 获取向量
			vec, err := llm.GetEmbedding(chunk)
			if err != nil {
				log.Printf("[RAG] Embedding 失败: %v", err)
				continue
			}

			// 逐点写入 Qdrant
			if err := upsertPoint(uuid.New().String(), vec, map[string]string{
				"text":   chunk,
				"source": file,
			}); err != nil {
				log.Printf("[RAG] 入库失败: %v", err)
				continue
			}
			pointCount++

			// 避免超过 API 频率限制
			time.Sleep(200 * time.Millisecond)
		}
		log.Printf("[RAG] 文件 %s 已成功入库 (%d 个片段)", file, pointCount)

		// 同时保留到内存供关键词检索兜底
		documentChunks = append(documentChunks, chunks...)
	}
	return documentChunks, nil
}

// ==================== 5. 向量检索 ====================
func SearchByVector(queryVector []float32, limit int) ([]string, error) {
	url := "http://localhost:6333/collections/knowledge_base/points/search"
	body, _ := json.Marshal(map[string]interface{}{
		"vector":       queryVector,
		"limit":        limit,
		"with_payload": true,
	})
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s", string(body))
	}
	var result struct {
		Result []struct {
			Payload map[string]string `json:"payload"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	var docs []string
	for _, p := range result.Result {
		if text, ok := p.Payload["text"]; ok {
			docs = append(docs, text)
		}
	}
	return docs, nil
}
