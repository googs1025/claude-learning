# 第9章：Claude Code CLI

## 概述

Claude Code 是 Anthropic 官方的命令行 AI 编程助手，直接在终端中使用 Claude 进行代码开发。

## 核心功能

### 1. 基本使用
```bash
# 启动交互式对话
claude

# 一次性提问
claude -p "解释这段代码的作用"

# 指定模型
claude --model claude-opus-4-6-20250925
```

### 2. 常用斜杠命令
| 命令 | 说明 |
|------|------|
| `/help` | 查看帮助信息 |
| `/clear` | 清空对话历史 |
| `/compact` | 压缩对话上下文 |
| `/model` | 切换模型 |
| `/cost` | 查看当前会话费用 |

### 3. 权限模式
- **Ask 模式**: 每次操作都需要确认（最安全）
- **Auto 模式**: 自动执行允许列表中的操作（高效）

### 4. Hooks 系统
Hooks 允许你在特定事件发生时执行自定义脚本：
- `PreToolUse`: 工具调用前
- `PostToolUse`: 工具调用后
- `Notification`: 通知事件

### 5. MCP Server 集成
Claude Code 可以连接外部 MCP Server 获得额外能力（数据库查询、API 调用等）。

## 文件说明

| 文件 | 说明 |
|------|------|
| `claude_code_cheatsheet.md` | 常用命令和快捷键速查表 |
| `hooks_examples/pre_tool_hook.sh` | Hook 脚本示例 |
| `mcp_config_example.json` | MCP Server 配置示例 |

## 配置文件位置

- 全局设置: `~/.claude/settings.json`
- 项目设置: `.claude/settings.json`
- MCP 配置: `.claude/settings.json` 中的 `mcpServers` 字段
- 快捷键: `~/.claude/keybindings.json`
