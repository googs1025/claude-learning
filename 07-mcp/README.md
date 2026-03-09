# 第七章：MCP（Model Context Protocol）

本章介绍 MCP 协议的核心概念和实现方式，学习如何使用 Go 构建 MCP Server。

## 什么是 MCP？

MCP（Model Context Protocol）是 Anthropic 推出的开放协议，用于标准化 AI 模型与外部工具/数据源之间的通信。它定义了一种统一的方式，让 AI 应用（如 Claude Desktop、Claude Code）能够与外部服务交互。

### 核心架构

```
┌─────────────────┐     MCP 协议      ┌─────────────────┐
│   MCP Client    │ ◄──────────────► │   MCP Server    │
│  (Claude Code,  │                   │  (你的 Go 程序)  │
│   Claude Desktop│    JSON-RPC 2.0   │                 │
│   等 AI 应用)    │                   │  提供工具、资源、 │
└─────────────────┘                   │  提示词等能力     │
                                      └─────────────────┘
```

### 通信方式（Transports）

| 传输方式 | 说明 | 适用场景 |
|---------|------|---------|
| **stdio** | 通过标准输入/输出通信 | 本地进程，Claude Desktop/Claude Code 集成 |
| **SSE** | 通过 HTTP Server-Sent Events 通信 | 远程服务，Web 应用集成 |

### 三大核心概念

#### 1. Tools（工具）
- 让 AI 模型调用外部功能（类似函数调用）
- 例如：计算器、文件操作、数据库查询、API 调用
- 每个工具有名称、描述和参数 schema

#### 2. Resources（资源）
- 让 AI 模型读取外部数据
- 使用 URI 标识（如 `file:///path/to/file`、`config://app/settings`）
- 例如：文件内容、配置信息、数据库记录

#### 3. Prompts（提示词模板）
- 预定义的提示词模板，可由用户选择使用
- 支持参数化，动态生成提示词
- 例如：代码审查模板、翻译模板

## 前置条件

1. 完成前六章的学习
2. 安装 mcp-go 库：
   ```bash
   go get github.com/mark3labs/mcp-go
   ```

## 课程内容

| 文件 | 主题 | 说明 |
|------|------|------|
| `01_mcp_server_basic.go` | 基础 MCP Server | 创建最简单的 MCP Server，注册问候工具 |
| `02_mcp_tools.go` | 工具定义 | 定义多种工具（计算器、文件读取），参数验证 |
| `03_mcp_resources.go` | 资源暴露 | 通过 URI 暴露文件内容和配置数据 |
| `04_mcp_integration.go` | 集成指南 | 与 Claude Desktop / Claude Code 集成配置 |

## 运行方式

MCP Server 通常不直接运行，而是由 MCP Client（如 Claude Desktop）启动。但你可以手动测试：

```bash
# 编译 MCP Server
go build -o my-server 01_mcp_server_basic.go

# 手动测试（通过 stdin 发送 JSON-RPC 消息）
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":{},"clientInfo":{"name":"test","version":"1.0"},"protocolVersion":"2024-11-05"}}' | ./my-server
```

## 与 Claude 集成

### Claude Desktop 配置

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`（macOS）：

```json
{
  "mcpServers": {
    "my-server": {
      "command": "/path/to/your/compiled/server"
    }
  }
}
```

### Claude Code 配置

编辑 `.claude/settings.json`：

```json
{
  "mcpServers": {
    "my-server": {
      "command": "/path/to/your/compiled/server"
    }
  }
}
```

## 参考资源

- [MCP 官方文档](https://modelcontextprotocol.io/)
- [mcp-go 库](https://github.com/mark3labs/mcp-go)
- [MCP 规范](https://spec.modelcontextprotocol.io/)
