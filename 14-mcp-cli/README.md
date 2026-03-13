# 第14章：MCP CLI 实战教学

用命令行方式学习和测试 MCP（Model Context Protocol），无需 GUI，全程终端操作。

> **前置知识**: 第7章（MCP 协议理论）、第11章（CLI 高级用法）
> **是否需要 API Key**: 第1-3节不需要，第4节需要

---

## 文件结构

```
14-mcp-cli/
├── README.md                 ← 本教学文档
└── servers/
    ├── calculator.go         ← MCP Server 源码（Go）
    └── calculator            ← 编译后的二进制（先运行第0步编译）
```

---

## 第 0 步：编译 MCP Server

所有后续步骤都依赖这个二进制文件。

```bash
cd 14-mcp-cli
go build -o servers/calculator servers/calculator.go
```

Calculator Server 提供四个工具：`add`（加）、`subtract`（减）、`multiply`（乘）、`divide`（除）

---

## 第 1 节：理解 MCP 底层协议（JSON-RPC 2.0）

MCP 的本质是 **JSON-RPC 2.0 over stdio**。每次 Claude 调用 MCP 工具时，
它实际上是通过 stdin 向 Server 发送 JSON 消息，从 stdout 读取响应。

你完全可以手动模拟这个过程。

### 1.1 建立连接（initialize 握手）

每个 MCP 会话必须先发送 `initialize` 请求：

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  | ./servers/calculator
```

**输出：**
```json
{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{"listChanged":true}},"serverInfo":{"name":"calculator","version":"1.0.0"}}}
```

解读：
- `protocolVersion`: 协议版本协商成功
- `capabilities.tools`: Server 声明支持工具功能
- `serverInfo`: Server 自报名称和版本

### 1.2 发现工具（tools/list）

查询 Server 提供哪些工具及其参数 schema：

```bash
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | ./servers/calculator
```

**输出（格式化后）：**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "add",
        "description": "计算两个数字的和",
        "inputSchema": {
          "type": "object",
          "properties": {
            "a": {"type": "number", "description": "第一个数字"},
            "b": {"type": "number", "description": "第二个数字"}
          },
          "required": ["a", "b"]
        }
      },
      {"name": "divide", "description": "计算两个数字的商", ...},
      {"name": "multiply", "description": "计算两个数字的积", ...},
      {"name": "subtract", "description": "计算两个数字的差", ...}
    ]
  }
}
```

> **这正是 Claude 在决定"何时、如何调用工具"前发送的第一个请求。**

### 1.3 调用工具（tools/call）

直接调用工具，传入参数：

```bash
# 加法：123 + 456
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"add","arguments":{"a":123,"b":456}}}' \
  | ./servers/calculator
```

**输出：**
```json
{"jsonrpc":"2.0","id":3,"result":{"content":[{"type":"text","text":"123 + 456 = 579"}]}}
```

```bash
# 乘法：7 × 8
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"multiply","arguments":{"a":7,"b":8}}}' \
  | ./servers/calculator
```

**输出：**
```json
{"jsonrpc":"2.0","id":4,"result":{"content":[{"type":"text","text":"7 × 8 = 56"}]}}
```

```bash
# 除以零（错误处理）
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"divide","arguments":{"a":5,"b":0}}}' \
  | ./servers/calculator
```

**输出：**
```json
{"jsonrpc":"2.0","id":5,"result":{"content":[{"type":"text","text":"错误：除数不能为零"}]}}
```

### MCP 协议流程图

```
你（或 Claude）                    MCP Server
     │                                 │
     │── initialize ──────────────────►│  建立连接，协商版本
     │◄── result（版本+能力）──────────│
     │                                 │
     │── tools/list ─────────────────►│  发现有哪些工具
     │◄── result（工具列表）──────────│
     │                                 │
     │── tools/call（name + args）────►│  调用具体工具
     │◄── result（content）───────────│  返回结果
     │                                 │
  stdin 关闭                       进程退出
```

---

## 第 2 节：用 `claude mcp` 管理 Server

`claude mcp` 是一组子命令，用于注册和管理 MCP Server。
注册后，Claude 会话会自动发现并使用这些工具，无需每次手动指定。

### 2.1 查看已注册的 Server

```bash
claude mcp list
```

**输出（初始状态）：**
```
No MCP servers configured. Use `claude mcp add` to add a server.
```

### 2.2 注册 Server

```bash
claude mcp add calculator $(pwd)/servers/calculator
```

**输出：**
```
Added stdio MCP server calculator with command: /path/to/14-mcp-cli/servers/calculator to local config
File modified: /Users/yourname/.claude.json [project: /path/to/14-mcp-cli]
Checking MCP server health...

calculator: /path/to/14-mcp-cli/servers/calculator - ✓ Connected
```

> `claude mcp add` 默认写入**本地配置**（私有，不提交 git）。

### 2.3 查看注册后的列表

```bash
claude mcp list
```

**输出：**
```
calculator: /path/to/14-mcp-cli/servers/calculator  - ✓ Connected
```

### 2.4 查看 Server 详情

```bash
claude mcp get calculator
```

**输出：**
```
calculator:
  Scope: Local config (private to you in this project)
  Status: ✓ Connected
  Type: stdio
  Command: /path/to/14-mcp-cli/servers/calculator
  Args:
  Environment:

To remove this server, run: claude mcp remove "calculator" -s local
```

### 2.5 删除注册

```bash
claude mcp remove calculator
```

**输出：**
```
Removed MCP server "calculator" from local config
```

### 作用域（Scope）说明

| 命令 | 写入位置 | 适用场景 |
|------|---------|---------|
| `claude mcp add <name> <cmd>` | `~/.claude.json`（本项目本地） | 默认，私有 |
| `claude mcp add --scope global <name> <cmd>` | `~/.claude/settings.json` | 所有项目都能用 |
| `claude mcp add --scope project <name> <cmd>` | `.claude/settings.json` | 团队共享（可提交 git） |
| `claude mcp add --scope local <name> <cmd>` | `.claude/settings.local.json` | 项目内私有 |

---

## 第 3 节：用 `--mcp-config` 临时加载 Server

不需要永久注册，可以用 `--mcp-config` 在单次 `claude -p` 调用中临时加载 MCP Server。
适合 CI/CD、脚本自动化、或一次性测试。

### 3.1 准备配置文件

```bash
# 创建 MCP 配置文件（注意替换为绝对路径）
cat > /tmp/calc-mcp.json << EOF
{
  "mcpServers": {
    "calculator": {
      "command": "$(pwd)/servers/calculator"
    }
  }
}
EOF
```

### 3.2 使用 `--mcp-config` 调用

```bash
claude -p "请用 calculator 工具计算 (123 + 456) × 789" \
  --mcp-config /tmp/calc-mcp.json \
  --model sonnet
```

Claude 会自动：
1. 启动 calculator Server
2. 发现 `add`、`multiply` 等工具
3. 决定调用顺序：先 `add(123, 456)` → 再 `multiply(579, 789)`
4. 返回最终结果

### 3.3 `--strict-mcp-config`：隔离模式

```bash
claude -p "计算 100 ÷ 7" \
  --mcp-config /tmp/calc-mcp.json \
  --strict-mcp-config \
  --model sonnet
```

加上 `--strict-mcp-config` 后：
- **只**使用 `--mcp-config` 中指定的 Server
- 忽略全局/项目级已注册的其他 Server
- 适用于 CI/CD 隔离，防止意外使用生产环境 Server

### 3.4 直接传 JSON 字符串（无需文件）

```bash
CALC_BIN=$(pwd)/servers/calculator
claude -p "计算 7 乘以 8" \
  --mcp-config "{\"mcpServers\":{\"calculator\":{\"command\":\"$CALC_BIN\"}}}" \
  --strict-mcp-config \
  --model sonnet
```

---

## 第 4 节：与 Hooks 结合（高级）

可以在 `settings.json` 中同时配置 MCP Server 和 Hooks：

```json
{
  "mcpServers": {
    "calculator": {
      "command": "/path/to/14-mcp-cli/servers/calculator"
    }
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "calculator:.*",
        "hooks": [
          {
            "type": "command",
            "command": "echo \"即将调用 calculator 工具\" >&2"
          }
        ]
      }
    ]
  }
}
```

> Hook 的 matcher 可以用 `服务器名:工具名` 格式精确匹配 MCP 工具。

---

## 核心命令速查

```bash
# ===== 编译 Server =====
go build -o servers/calculator servers/calculator.go

# ===== 手动 JSON-RPC 测试 =====
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./servers/calculator
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./servers/calculator
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"add","arguments":{"a":1,"b":2}}}' | ./servers/calculator

# ===== claude mcp 管理 =====
claude mcp list
claude mcp add calculator $(pwd)/servers/calculator
claude mcp get calculator
claude mcp remove calculator

# ===== --mcp-config 临时加载 =====
claude -p "计算 7+8" --mcp-config /tmp/calc-mcp.json --model sonnet
claude -p "计算 7+8" --mcp-config /tmp/calc-mcp.json --strict-mcp-config --model sonnet
```

---

## 与第 7 章的对比

| | 第7章 (`07-mcp/`) | 第14章 (`14-mcp-cli/`) |
|-|------------------|----------------------|
| 重点 | 如何**编写** MCP Server | 如何**使用**和**测试** MCP |
| 工具 | Go SDK | `echo` 管道 / `claude mcp` CLI |
| 是否需要 API Key | 是 | 第1-3节不需要 |
| 学习目标 | Server 实现原理 | CLI 集成与调试方法 |