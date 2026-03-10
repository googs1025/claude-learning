# Claude 全面学习指南 (Golang)

使用 Go 语言系统学习 Claude API 及相关生态，从入门到实战。

## 两条学习路径

### 路径 A：无 API Key（Claude Pro/Max 订阅用户）

只需安装 `claude` CLI，无需 API Key：

```bash
# 1. 安装 Go 1.22+
# 2. 安装 Claude Code CLI
npm install -g @anthropic-ai/claude-code
```

推荐学习顺序：**00** → **09** → **11** → **12** → **07**（MCP Server 不需要 API Key）

### 路径 B：有 API Key（完整学习路径）

```bash
# 1. 安装 Go 1.22+
# 2. 获取 API Key: https://console.anthropic.com/
# 3. 设置环境变量
export ANTHROPIC_API_KEY="your-api-key"

# 4. 安装依赖
go mod tidy
```

推荐学习顺序：**01** → **10** 全部章节

## 学习路线

| 章节 | 主题 | 难度 | 需要 API Key | 说明 |
|------|------|------|:---:|------|
| [00-cli-quick-start](./00-cli-quick-start/) | CLI 快速入门 | ★☆☆ | | 通过 `claude` CLI 调用，零门槛上手 |
| [01-api-basics](./01-api-basics/) | API 基础 | ★☆☆ | ✓ | SDK 初始化、Messages API、流式响应 |
| [02-prompt-engineering](./02-prompt-engineering/) | Prompt 工程 | ★☆☆ | ✓ | Few-shot、思维链、结构化输出 |
| [03-tool-use](./03-tool-use/) | Tool Use | ★★☆ | ✓ | 函数调用、工具定义、Agent 循环 |
| [04-vision](./04-vision/) | 图像理解 | ★★☆ | ✓ | 图像分析、多图对比 |
| [05-extended-thinking](./05-extended-thinking/) | 扩展思维 | ★★☆ | ✓ | 深度推理、思维预算控制 |
| [06-advanced-patterns](./06-advanced-patterns/) | 高级模式 | ★★★ | ✓ | Prompt 缓存、批处理、成本控制 |
| [07-mcp](./07-mcp/) | MCP 协议 | ★★★ | | 构建 MCP Server、工具与资源 |
| [08-agent-patterns](./08-agent-patterns/) | Agent 模式 | ★★★ | ✓ | ReAct、多 Agent、记忆管理 |
| [09-claude-code](./09-claude-code/) | Claude Code | ★★☆ | | CLI 使用、Hooks、MCP 配置 |
| [10-projects](./10-projects/) | 综合实战 | ★★★ | ✓ | 聊天机器人、代码审查、RAG 助手 |
| [11-cli-mastery](./11-cli-mastery/) | CLI 高级技巧 | ★★☆ | | 会话管理、输出解析、工具限制、预算控制 |
| [12-skills](./12-skills/) | Skills 技能扩展 | ★★☆ | | 自定义 Skill 开发、技能测试 |

## 技术栈

- **语言**: Go 1.22+
- **SDK**: [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go)
- **MCP**: [mcp-go](https://github.com/mark3labs/mcp-go)
- **默认模型**: `claude-sonnet-4-6-20250514`

## 运行示例

```bash
# 运行单个示例
cd 01-api-basics
go run 01_hello_claude.go

# 运行实战项目
cd 10-projects/chatbot
go run main.go
```
