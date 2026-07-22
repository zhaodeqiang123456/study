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

// ==================== 2. Embedding 调用 ====================

// ==================== 3. 批量写入 Qdrant（HTTP） ====================
const qdrantBaseURL = "http://localhost:6333"

func upsertPoints(points []map[string]interface{}) error {
	url := fmt.Sprintf("%s/collections/knowledge_base/points?wait=true", qdrantBaseURL)
	body, err := json.Marshal(map[string]interface{}{
		"points": points,
	})
	if err != nil {
		return fmt.Errorf("序列化点失败: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
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
	dir := `D:\study\projects\study\project\pkg\rag\documents`
	log.Printf("[RAG] 文档目录: %s", dir)
	pattern := filepath.Join(dir, "*.txt")
	log.Printf("[RAG] 搜索模式: %s", pattern)
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
		// ========================================== 接入向量数据库
		// var points []map[string]interface{}
		// for _, chunk := range chunks {
		// 	// 调用 Embedding 获取向量
		// 	vec, err := llm.GetEmbedding(chunk)
		// 	if err != nil {
		// 		log.Printf("[RAG] Embedding 失败: %v", err)
		// 		continue
		// 	}

		// 	// 构造一个 Point
		// 	point := map[string]interface{}{
		// 		"id":     uuid.New().String(),
		// 		"vector": vec, // []float32
		// 		"payload": map[string]string{
		// 			"text":   chunk,
		// 			"source": file,
		// 		},
		// 	}
		// 	points = append(points, point)

		// 	// 避免超过 DeepSeek 频率限制，每次调用后短暂休息
		// 	time.Sleep(200 * time.Millisecond)
		// }

		// // 分批写入 Qdrant，每批 100 个
		// batchSize := 100
		// for i := 0; i < len(points); i += batchSize {
		// 	end := i + batchSize
		// 	if end > len(points) {
		// 		end = len(points)
		// 	}
		// 	if err := upsertPoints(points[i:end]); err != nil {
		// 		log.Printf("[RAG] 入库失败: %v", err)
		// 		continue
		// 	}
		// }
		// log.Printf("[RAG] 文件 %s 已成功入库 (%d 个片段)", file, len(points))
		for _, chunk := range chunks {
			documentChunks = append(documentChunks, chunk)
		}
		log.Printf("[RAG] 文件 %s 已成功加载入内存 (%d 个片段)", file, len(documentChunks))
	}
	return documentChunks, err
}

// 向量数据库检索
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
