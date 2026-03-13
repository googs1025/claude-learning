# 第13章：Hooks 自动化守门人

## 概述

Claude Code Hooks 是一套事件驱动的自动化系统，让你在 Claude 执行工具的前后插入自定义逻辑。通过 Hooks，你可以：

- **拦截危险命令**：在 `rm -rf /` 执行前自动阻断
- **保护敏感文件**：阻止 AI 覆写 `.env`、`credentials` 等配置
- **自动化质量保证**：写代码后自动运行 `gofmt`、`prettier`、测试套件
- **注入环境上下文**：会话启动时把 git 状态、依赖版本告知 Claude
- **任务完成通知**：长任务完成后发送 macOS 桌面通知

> **前置知识**: 本章假设你已了解第 9 章的 Claude Code 基础和第 11 章的 CLI 高级功能（特别是 `05_hooks_integration.go`）。

## 核心机制

### Hook 的工作原理

```
[Claude 决定执行工具]
        ↓
[PreToolUse Hook 触发]
  → Claude Code 将 JSON 写入 hook 脚本的 stdin
  → hook 读取 JSON，分析，决策
  → exit 0          : 允许执行（静默通过）
  → exit 2          : 阻断执行（快速拒绝）
  → JSON输出+exit 0  : 携带原因的阻断/允许
        ↓
[工具执行]
        ↓
[PostToolUse Hook 触发]
  → 同上读取 JSON（包含工具输出）
  → 不能阻断，用于：格式化、测试、日志
```

### stdin JSON 格式（Hook 接收的数据）

```json
{
  "tool_name": "Bash",
  "tool_input": {
    "command": "rm -rf /tmp/test"
  },
  "session_id": "sess_abc123"
}
```

Write 工具的 JSON：
```json
{
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/project/.env",
    "content": "..."
  }
}
```

### Hook 脚本核心模式

```bash
#!/bin/bash
# 关键：从 stdin 读取 JSON（不是环境变量！）
INPUT=$(cat)

# 用 python3 解析 JSON
TOOL_NAME=$(echo "$INPUT" | python3 -c \
    "import sys,json; print(json.load(sys.stdin).get('tool_name',''))")
COMMAND=$(echo "$INPUT" | python3 -c \
    "import sys,json; print(json.load(sys.stdin).get('tool_input',{}).get('command',''))")

# 决策
if echo "$COMMAND" | grep -q "危险模式"; then
    # 推荐方式：JSON 输出（可携带原因）
    echo '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"原因"}}'
    exit 0
fi
exit 0
```

### Hook 事件类型

| 事件 | 触发时机 | 能否阻断 | stdout 用途 |
|------|---------|---------|-------------|
| `PreToolUse` | 工具执行前 | ✅ 可以 | 忽略 |
| `PostToolUse` | 工具执行后 | ❌ 不能 | 忽略 |
| `SessionStart` | 会话启动时 | ❌ 不能 | **注入 Claude 上下文** |
| `Stop` | 会话结束时 | ❌ 不能 | 忽略 |

### 退出码含义

| 退出码 | 含义 |
|-------|------|
| `exit 0` | 正常（允许执行，或 JSON deny 决策） |
| `exit 1` | 错误（非阻断，Claude 收到错误信息继续执行） |
| `exit 2` | **阻断**（PreToolUse 专用，拒绝工具执行） |

## 文件说明

### Go 示例

| 文件 | 主题 | 关键知识点 |
|------|------|-----------|
| `01_hook_basics.go` | Hook 机制原理 | stdin JSON 解析，退出码，基础配置 |
| `02_security_hooks.go` | 安全防御 Hooks | 危险命令拦截，敏感文件保护，自动备份 |
| `03_quality_hooks.go` | 代码质量 Hooks | 敏感数据检测，自动格式化，自动测试 |
| `04_context_hooks.go` | 上下文增强 Hooks | SessionStart 注入，Stop 通知 |
| `05_advanced_patterns.go` | 高级组合模式 | 综合配置，matcher 正则，多 hooks 机制 |
| `06_cli_demo.go` | 命令行直接测试 | os/exec 调用脚本，模拟 stdin JSON，验证所有 hook |

### Shell 脚本（`hooks/` 目录）

| 脚本 | 事件 | 功能 |
|------|------|------|
| `01_security_guard.sh` | PreToolUse Bash | 拦截 `rm -rf /`、`sudo rm` 等危险命令 |
| `02_file_protector.sh` | PreToolUse Write | 阻止覆写 `.env`、`credentials`、`*.pem` 等 |
| `03_backup_before_edit.sh` | PreToolUse Write/Edit | 覆写前自动备份到 `/tmp/claude_backup/` |
| `04_sensitive_data.sh` | PostToolUse Write | 检测写入文件中的 API Key / 密码 |
| `05_auto_format.sh` | PostToolUse Write | 按扩展名运行 gofmt / prettier / black |
| `06_run_tests.sh` | PostToolUse Write | 源码变更后自动运行测试套件 |
| `07_session_check.sh` | SessionStart | 体检 git 状态、Go 版本、环境变量，注入上下文 |
| `08_stop_notify.sh` | Stop | macOS 桌面通知 + 会话摘要 |
| `09_audit_log.sh` | PostToolUse Bash | Bash 命令写入 `/tmp/claude_audit.log` |

## 运行方式

```bash
cd 13-hooks-system

# 查看 Hook 机制原理
go run 01_hook_basics.go

# 安全防御 Hooks 演示
go run 02_security_hooks.go

# 代码质量 Hooks 演示
go run 03_quality_hooks.go

# 上下文增强 Hooks 演示
go run 04_context_hooks.go

# 高级组合（生成完整配置）
go run 05_advanced_patterns.go

# 命令行直接测试所有 hooks（模拟 Claude Code 行为）
go run 06_cli_demo.go
```

## 手动测试 Hook 脚本

```bash
# 验证脚本语法
bash -n hooks/01_security_guard.sh

# 测试危险命令拦截（应返回 JSON deny）
echo '{"tool_name":"Bash","tool_input":{"command":"rm -rf /"}}' \
  | bash hooks/01_security_guard.sh

# 测试安全命令（应静默通过，exit 0）
echo '{"tool_name":"Bash","tool_input":{"command":"ls -la"}}' \
  | bash hooks/01_security_guard.sh; echo "退出码: $?"

# 测试敏感文件保护（应返回 JSON deny）
echo '{"tool_name":"Write","tool_input":{"file_path":"/project/.env"}}' \
  | bash hooks/02_file_protector.sh

# 查看审计日志
tail -f /tmp/claude_audit.log
```

## 部署到全局

```bash
# 1. 给脚本添加执行权限
chmod +x hooks/*.sh

# 2. 复制到用户 hooks 目录
mkdir -p ~/.claude/hooks
cp hooks/*.sh ~/.claude/hooks/

# 3. 生成完整配置
go run 05_advanced_patterns.go

# 4. 将生成的 complete_hooks_config.json 合并到全局配置
# （手动编辑 ~/.claude/settings.json 或直接替换）
cat complete_hooks_config.json
```

## 学习要点

- **核心错误**：Hook 通过 stdin 接收 JSON，不要用环境变量获取工具信息
- **阻断方式**：PreToolUse 用 `exit 2` 或输出 JSON `{permissionDecision: deny}`
- **PostToolUse 限制**：工具已执行，只能发警告（stderr），不能阻断
- **SessionStart 特权**：stdout 会成为 Claude 的初始上下文（动态注入环境信息）
- **多 hooks 并发**：不同 matcher 匹配的 hooks 会并发执行
- **配置层级**：`--settings` > `.claude/settings.local.json` > `.claude/settings.json` > `~/.claude/settings.json`
