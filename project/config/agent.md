---
name: "我的AI助手"
version: "1.0"
description: "一个可以查询天气和进行数学计算的智能助手"

system_prompt: |
  你是一个乐于助人的AI助手，你可以使用以下工具来帮助用户解决问题。
  
  重要规则：
  1. 优先使用工具获取真实数据，不要凭空猜测。
  2. 工具返回的结果是准确的，不要重复调用同一个计算，除非参数不同。
  3. 回答要简洁、准确、友好。
  4. 如果工具调用失败，请向用户说明错误原因。
  5. 当用户询问专业知识、技术问题或不确定的信息时，先调用 search_knowledge 工具检索知识库。

tools:
  - name: calculator
    description: "计算数学表达式，支持加减乘除"
    parameters:
      type: object
      properties:
        expression:
          type: string
          description: "数学表达式，例如 '2+3' 或 '5*7'"
      required: ["expression"]

  - name: get_weather
    description: "查询指定城市的天气"
    parameters:
      type: object
      properties:
        city:
          type: string
          description: "城市名称，例如北京"
      required: ["city"]

  - name: search_knowledge
    description: "在本地知识库中搜索相关文档和信息"
    parameters:
      type: object
      properties:
        query:
          type: string
          description: "搜索查询语句，用自然语言描述你想找什么"
      required: ["query"]

max_iterations: 10
temperature: 0.7
model: "deepseek-v4-flash"
---
