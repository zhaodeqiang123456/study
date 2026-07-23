package agent

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openai/openai-go"
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

// rag 增强检索  --- 关键字匹配简易版
func searchKnowledgeTool(query string) (string, error) {

	// keywords := strings.Fields(query)
	var results string
	// for _, chunk := range documentChunks {
	// 	for _, kw := range keywords {
	// 		if strings.Contains(chunk, kw) {
	// 			results = append(results, chunk)
	// 			break
	// 		}
	// 	}
	// }
	// log.Printf("[RAG] 文件已成功匹配 (%d 个片段), 原文件含有 %d 个片段", len(results), len(documentChunks))

	// ==================== 向量检索 ====================
	// queryVector, _ := llm.GetEmbedding(task.Prompt)
	// // 检索相关文档
	// var docs []string
	// if queryVector != nil {
	// 	docs, err = rag.SearchByVector(queryVector, 3)
	// 	if err != nil {
	// 		log.Printf("search error: %v", err)
	// 		docs = nil
	// 	}
	// }

	// ========================= 关键词搜索 =============
	// docs := searchByKeyword(task.Prompt)
	// // 构建增强 prompt
	// augmentedPrompt := task.Prompt
	// if len(docs) > 0 {
	// 	augmentedPrompt = fmt.Sprintf(
	// 		"你是一个专业的AI助手。以下是可能相关的参考资料：\n\n%s\n\n"+
	// 			"用户问题：%s\n\n"+
	// 			"请根据参考资料提供准确答案，并在此基础上适当补充背景知识、实际案例或相关技术细节，使回答更加丰富和有用。",
	// 		strings.Join(docs, "\n\n"),
	// 		task.Prompt,
	// 	)
	// }

	// // ==================== 流式调用 ====================
	// history := dbService.QueryHistory(task.ConversationID)
	// // 把增强的prompt追加到最后
	// history = append(history, openai.UserMessage(augmentedPrompt))
	// // SSE 流式调用
	// reply, err = llm.ProcessTaskStreamly(apiKey, task.ID, history, srv)

	// if err != nil {
	// 	log.Printf("LLM error: %v", err)
	// 	return
	// }
	return results, nil
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
