# 第五章：扩展思考（Extended Thinking）

本章介绍如何使用 Claude 的扩展思考能力，让 Claude 在回答前进行深度推理。

## 前置条件

1. 安装 Go 1.21+
2. 设置环境变量 `ANTHROPIC_API_KEY`
3. 安装依赖：
   ```bash
   go get github.com/anthropics/anthropic-sdk-go
   ```

## 课程内容

| 文件 | 主题 | 说明 |
|------|------|------|
| `01_basic_thinking.go` | 基础扩展思考 | 启用扩展思考，解决复杂数学问题，展示思考过程和最终答案 |
| `02_budget_tokens.go` | 思考预算控制 | 对比不同的 BudgetTokens 设置对推理深度的影响 |
| `03_streaming_thinking.go` | 流式思考输出 | 实时流式输出思考过程和文本内容 |
| `04_thinking_with_tools.go` | 思考 + 工具 | 在工具调用场景中启用扩展思考，展示交错的思考过程 |

## 运行方式

```bash
# 确保设置了 API 密钥
export ANTHROPIC_API_KEY="your-api-key"

# 运行任意示例
go run 01_basic_thinking.go
go run 02_budget_tokens.go
go run 03_streaming_thinking.go
go run 04_thinking_with_tools.go
```

## 核心概念

### 什么是扩展思考？
扩展思考（Extended Thinking）让 Claude 在给出最终回答之前，先进行一段内部推理过程。这类似于人类在解决复杂问题时的"思考过程"——先分析、推理、验证，最后再给出答案。

### BudgetTokens（思考预算）
- `BudgetTokens` 控制 Claude 可以用于思考的最大 token 数量
- 最小值为 **1024**，且必须小于 `MaxTokens`
- 更大的预算允许更深入的推理，但也会消耗更多的 token 和时间
- 思考 token 会计入使用量并产生费用

### 思考内容块
启用扩展思考后，Claude 的响应中会包含 `ThinkingBlock` 类型的内容块：
- `ThinkingBlock.Thinking`：思考过程的文本内容
- 思考块出现在文本回答之前

### 流式思考
流式模式下，思考内容通过 `ThinkingDelta` 事件逐步返回，让用户可以实时观察 Claude 的推理过程。

### 与工具调用结合
扩展思考可以与工具调用一起使用。Claude 会在决定调用工具之前进行推理，帮助它更准确地选择工具和构造参数。
