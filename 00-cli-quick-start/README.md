# 第零章：CLI 快速入门（无需 API Key）

本章使用 `claude` CLI 命令（通过 Go 的 `os/exec` 调用），**无需 API Key**，只需安装 Claude Code CLI 即可运行。

## CLI 模式 vs API 模式

| | CLI 模式（本章） | API 模式（第 1-8 章） |
|---|---|---|
| **前置条件** | 安装 `claude` CLI + Pro/Max 订阅 | Anthropic API Key |
| **调用方式** | `os/exec` 调用 `claude -p` | `anthropic-sdk-go` SDK |
| **支持功能** | 对话、系统提示、流式输出、JSON Schema | Tool Use、Vision、Extended Thinking 等全部功能 |
| **适用场景** | 快速上手、脚本集成、无 API Key 用户 | 生产环境、复杂应用 |

## 前置条件

1. 安装 Go 1.22+
2. 安装 Claude Code CLI：
   ```bash
   npm install -g @anthropic-ai/claude-code
   ```
3. 确认 CLI 可用：
   ```bash
   claude --version
   ```

> **注意**：无需设置 `ANTHROPIC_API_KEY` 环境变量，CLI 使用你的 Claude Pro/Max 订阅认证。

## 课程内容

| 文件 | 主题 | 说明 |
|------|------|------|
| `01_hello_claude.go` | 最简 CLI 调用 | 通过 `exec.Command` 调用 `claude -p`，获取回复 |
| `02_conversation.go` | 多轮对话 | 使用 `--continue` 模拟连续对话 |
| `03_streaming.go` | 流式输出 | 使用 `--output-format stream-json` 逐行读取 |
| `04_system_prompt.go` | 系统提示词 | 使用 `--append-system-prompt` 控制行为 |
| `05_structured_output.go` | 结构化输出 | 使用 `--output-format json` 和 `--json-schema` |
| `06_prompt_chaining.go` | 链式调用 | 一个 Claude 的输出作为下一个的输入 |

## 运行方式

```bash
# 每个文件都可以独立运行
go run 01_hello_claude.go
go run 02_conversation.go
go run 03_streaming.go
go run 04_system_prompt.go
go run 05_structured_output.go
go run 06_prompt_chaining.go
```

## CLI 常用参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `-p "prompt"` | 非交互模式，传入提示词 | `claude -p "你好"` |
| `--continue` | 继续上一次对话 | `claude -p "继续" --continue` |
| `--output-format json` | JSON 格式输出 | `claude -p "问题" --output-format json` |
| `--output-format stream-json` | 流式 JSON 输出 | `claude -p "问题" --output-format stream-json` |
| `--append-system-prompt "..."` | 追加系统提示词 | `claude -p "翻译" --append-system-prompt "你是翻译专家"` |
| `--model sonnet` | 选择模型 | `claude -p "问题" --model sonnet` |
| `--json-schema '{...}'` | 指定输出 JSON Schema | 见 `05_structured_output.go` |
