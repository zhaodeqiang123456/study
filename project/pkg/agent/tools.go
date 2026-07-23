package agent

import (
	"fmt"
	"strconv"
	"strings"
)

// ToolCall 代表 LLM 要求执行的函数调用
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON 字符串
}

// ToolResult 函数执行后的返回结果
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
}

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
