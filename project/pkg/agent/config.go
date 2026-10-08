package agent

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AgentConfig 对应 agent.md 的配置结构
type AgentConfig struct {
	Name          string    `yaml:"name"`
	Version       string    `yaml:"version"`
	Description   string    `yaml:"description"`
	SystemPrompt  string    `yaml:"system_prompt"`
	Tools         []ToolDef `yaml:"tools"`
	MaxIterations int       `yaml:"max_iterations"`
	Temperature   float64   `yaml:"temperature"`
	Model         string    `yaml:"model"`
}

// ToolDef 对应工具定义
type ToolDef struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Parameters  map[string]interface{} `yaml:"parameters"`
}

// LoadConfig 从 agent.md 加载配置
func LoadConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 提取 YAML 部分（以 "---" 开头到下一个 "---" 或文件结束）
	// 这里简单处理：直接解析整个文件为 YAML，忽略 Markdown 内容
	var config AgentConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 设置默认值
	if config.MaxIterations == 0 {
		config.MaxIterations = 10
	}
	if config.Temperature == 0.0 {
		config.Temperature = 0.7
	}
	if config.Model == "" {
		config.Model = "deepseek-chat"
	}

	return &config, nil
}
