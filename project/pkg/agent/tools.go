package agent

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openai/openai-go"
)

// 函数注入：由 consumer 在启动时填入具体实现，避免循环依赖
var (
	EmbeddingFunc    func(string) ([]float32, error)
	VectorSearchFunc func([]float32, int) ([]string, error)
)

// ExecuteTool 根据工具名称执行对应函数
func ExecuteTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "calculator":
		expression, ok := args["expression"].(string)
		if !ok {
			return "", fmt.Errorf("缺少参数 expression")
		}
		return calculatorTool(expression)
	case "get_weather":
		city, ok := args["city"].(string)
		if !ok {
			return "", fmt.Errorf("缺少参数 city")
		}
		return getWeatherTool(city)

	case "search_knowledge":
		query, ok := args["query"].(string)
		if !ok {
			return "", fmt.Errorf("缺少参数 query")
		}
		return searchKnowledgeTool(query)
	default:
		return "", fmt.Errorf("未知工具: %s", name)
	}
}

// 简易计算器：只支持加减乘除
func calculatorTool(expression string) (string, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	parts := strings.FieldsFunc(expression, func(r rune) bool {
		return r == '+' || r == '-' || r == '*' || r == '/'
	})
	if len(parts) != 2 {
		return "", fmt.Errorf("无效表达式")
	}
	left, err1 := strconv.ParseFloat(parts[0], 64)
	right, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return "", fmt.Errorf("数字解析失败")
	}
	var result float64
	switch {
	case strings.Contains(expression, "+"):
		result = left + right
	case strings.Contains(expression, "-"):
		result = left - right
	case strings.Contains(expression, "*"):
		result = left * right
	case strings.Contains(expression, "/"):
		if right == 0 {
			return "", fmt.Errorf("除数不能为零")
		}
		result = left / right
	}
	return fmt.Sprintf("%.2f", result), nil
}

// 模拟天气查询
func getWeatherTool(city string) (string, error) {
	weatherDB := map[string]string{
		"北京": "晴天，37°C，湿度 40%",
		"上海": "多云，32°C，湿度 65%",
		"深圳": "阵雨，29°C，湿度 85%",
	}
	if weather, ok := weatherDB[city]; ok {
		return weather, nil
	}
	return fmt.Sprintf("未找到 %s 的天气信息", city), nil
}

// rag 向量增强检索
func searchKnowledgeTool(query string) (string, error) {
	// 1. 获取查询向量
	if EmbeddingFunc == nil {
		return "", fmt.Errorf("Embedding 服务未初始化")
	}
	queryVector, err := EmbeddingFunc(query)
	if err != nil {
		return "", fmt.Errorf("查询 Embedding 失败: %w", err)
	}

	// 2. 向量检索
	if VectorSearchFunc == nil {
		return "", fmt.Errorf("向量检索服务未初始化")
	}
	docs, err := VectorSearchFunc(queryVector, 3)
	if err != nil {
		return "", fmt.Errorf("向量检索失败: %w", err)
	}
	if len(docs) == 0 {
		return "未找到相关文档", nil
	}

	// 3. 拼接检索结果作为参考上下文
	return strings.Join(docs, "\n\n---\n\n"), nil
}

func MergeDeltaToolCalls(
	existing []openai.ChatCompletionChunkChoiceDeltaToolCall,
	incoming []openai.ChatCompletionChunkChoiceDeltaToolCall,
) []openai.ChatCompletionChunkChoiceDeltaToolCall {
	for _, in := range incoming {
		if int(in.Index) < len(existing) {
			if in.Function.Arguments != "" {
				existing[in.Index].Function.Arguments += in.Function.Arguments
			}
			// 如果 ID 是空的，保留之前的 ID
			if in.ID != "" {
				existing[in.Index].ID = in.ID
			}
		} else {
			existing = append(existing, in)
		}
	}
	return existing
}
