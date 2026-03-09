# Claude Code 速查表

## 启动与基本用法

```bash
# 交互式启动
claude

# 一次性提问（不进入交互模式）
claude -p "你的问题"

# 管道输入
cat file.go | claude -p "review this code"

# 从文件读取 prompt
claude -p "$(cat prompt.txt)"

# 恢复上次对话
claude --resume

# 指定模型
claude --model claude-opus-4-6-20250925
```

## 斜杠命令

| 命令 | 说明 |
|------|------|
| `/help` | 帮助信息 |
| `/clear` | 清空对话 |
| `/compact` | 压缩上下文（节省 token） |
| `/model` | 切换模型 |
| `/cost` | 查看费用 |
| `/permissions` | 管理权限 |
| `/fast` | 切换快速模式 |

## 键盘快捷键

| 快捷键 | 功能 |
|--------|------|
| `Enter` | 发送消息 |
| `Shift+Enter` | 换行 |
| `Ctrl+C` | 中断当前操作 |
| `Ctrl+D` | 退出 |
| `Esc` | 取消当前输入 |

## 权限模式

```bash
# Ask 模式（默认，每次操作需确认）
claude --permission-mode ask

# Auto 模式（自动执行白名单操作）
claude --permission-mode auto
```

## Hooks 配置

在 `.claude/settings.json` 中配置：

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "./hooks/pre_bash.sh"
          }
        ]
      }
    ],
    "PostToolUse": [],
    "Notification": []
  }
}
```

## MCP Server 配置

在 `.claude/settings.json` 中添加：

```json
{
  "mcpServers": {
    "my-server": {
      "command": "go",
      "args": ["run", "./my-mcp-server/main.go"],
      "env": {
        "API_KEY": "xxx"
      }
    }
  }
}
```

## 常用工作流

### 代码审查
```bash
claude -p "review the changes in this PR" < <(git diff main)
```

### 生成测试
```bash
claude -p "为 main.go 中的函数生成单元测试"
```

### 修复 Bug
```bash
claude -p "这个测试失败了，帮我找出原因并修复" < <(go test ./... 2>&1)
```

### 重构代码
```bash
claude -p "将这个文件重构为更符合 Go 惯用写法"
```
