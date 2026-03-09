# 第一章：API 基础

本章介绍如何使用 Go 语言调用 Claude API，涵盖最核心的基础操作。

## 前置条件

1. 安装 Go 1.21+
2. 设置环境变量 `ANTHROPIC_API_KEY`（从 [Anthropic Console](https://console.anthropic.com/) 获取）
3. 安装依赖：
   ```bash
   go get github.com/anthropics/anthropic-sdk-go
   ```

## 课程内容

| 文件 | 主题 | 说明 |
|------|------|------|
| `01_hello_claude.go` | 最简 API 调用 | 发送一条消息，获取回复，理解基本请求/响应结构 |
| `02_conversation.go` | 多轮对话 | 使用 `ToParam()` 构建对话历史，实现连续对话 |
| `03_streaming.go` | 流式响应 | 使用 `NewStreaming` 实时获取生成内容，逐字输出 |
| `04_error_handling.go` | 错误处理 | 处理 API 错误，获取状态码和错误信息 |
| `05_system_prompt.go` | 系统提示词 | 通过系统提示词控制 Claude 的行为和角色设定 |

## 运行方式

每个文件都可以独立运行：

```bash
# 确保设置了 API 密钥
export ANTHROPIC_API_KEY="your-api-key"

# 运行任意示例
go run 01_hello_claude.go
go run 02_conversation.go
go run 03_streaming.go
go run 04_error_handling.go
go run 05_system_prompt.go
```

## 核心概念

### 客户端初始化
SDK 会自动从环境变量 `ANTHROPIC_API_KEY` 读取密钥，无需手动传入。

### 消息结构
Claude API 使用"消息"模型：每次请求包含一个消息列表（用户消息和助手消息交替出现），API 返回助手的下一条消息。

### 流式 vs 非流式
- **非流式**：等待完整响应后一次性返回，适合后台处理
- **流式**：逐步返回生成内容，适合实时展示给用户

### 系统提示词
系统提示词在消息列表之外单独设置，用于定义 Claude 的角色、行为规范和输出格式。
