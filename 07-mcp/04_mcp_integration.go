// 第七章示例 4：MCP 集成指南
// 运行方式: go run 04_mcp_integration.go
//
// 本示例是一个完整的 MCP Server，同时包含详细的集成配置说明。
// 它演示如何将自己编写的 MCP Server 与 Claude Desktop 和 Claude Code 集成。
//
// ============================================================
// 第一部分：集成配置说明
// ============================================================
//
// 【步骤 1】编译你的 MCP Server
//
//   go build -o my-mcp-server 04_mcp_integration.go
//
//   编译后会得到一个可执行文件 my-mcp-server，记录其绝对路径。
//   例如: /Users/yourname/claude-learning/07-mcp/my-mcp-server
//
// 【步骤 2a】配置 Claude Desktop
//
//   编辑配置文件（macOS 路径）:
//   ~/Library/Application Support/Claude/claude_desktop_config.json
//
//   Windows 路径:
//   %APPDATA%\Claude\claude_desktop_config.json
//
//   配置内容:
//   {
//     "mcpServers": {
//       "my-demo-server": {
//         "command": "/absolute/path/to/my-mcp-server"
//       }
//     }
//   }
//
//   如果你的 Server 需要环境变量（如 API Key）:
//   {
//     "mcpServers": {
//       "my-demo-server": {
//         "command": "/absolute/path/to/my-mcp-server",
//         "env": {
//           "DATABASE_URL": "postgres://localhost:5432/mydb",
//           "API_KEY": "your-api-key"
//         }
//       }
//     }
//   }
//
//   如果你的 Server 需要命令行参数:
//   {
//     "mcpServers": {
//       "my-demo-server": {
//         "command": "/absolute/path/to/my-mcp-server",
//         "args": ["--port", "8080", "--verbose"]
//       }
//     }
//   }
//
// 【步骤 2b】配置 Claude Code
//
//   方法一：项目级配置（推荐）
//   在项目根目录创建 .claude/settings.json:
//   {
//     "mcpServers": {
//       "my-demo-server": {
//         "command": "/absolute/path/to/my-mcp-server"
//       }
//     }
//   }
//
//   方法二：使用 claude mcp 命令
//   claude mcp add my-demo-server /absolute/path/to/my-mcp-server
//
// 【步骤 3】验证集成
//
//   Claude Desktop: 重启应用后，在对话框中应看到工具图标
//   Claude Code: 重启后使用 /mcp 命令查看已连接的 Server
//
// 【常见问题排查】
//
//   1. Server 没有出现:
//      - 确认可执行文件路径正确且有执行权限 (chmod +x)
//      - 检查 JSON 配置文件格式是否正确
//      - 重启 Claude Desktop / Claude Code
//
//   2. 工具调用失败:
//      - 检查 Server 的 stderr 输出（日志）
//      - 确认所需的环境变量已正确设置
//      - 手动运行 Server 测试是否正常
//
//   3. 调试技巧:
//      - Server 的 fmt.Fprintln(os.Stderr, ...) 输出会显示在 Client 的日志中
//      - 使用 MCP Inspector 工具进行调试: npx @modelcontextprotocol/inspector
//
// ============================================================
// 第二部分：示例 MCP Server 代码
// ============================================================

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 创建 MCP Server
	s := server.NewMCPServer(
		"integration-demo",
		"1.0.0",
	)

	// ==================== 注册工具 ====================

	// 工具 1: 获取当前时间
	s.AddTool(
		mcp.NewTool(
			"current_time",
			mcp.WithDescription("获取当前的日期和时间"),
			mcp.WithString(
				"format",
				mcp.Description("时间格式（simple: 简单格式，full: 完整格式，unix: Unix 时间戳）"),
			),
		),
		currentTimeHandler,
	)

	// 工具 2: JSON 格式化
	s.AddTool(
		mcp.NewTool(
			"format_json",
			mcp.WithDescription("格式化 JSON 字符串，使其更易读"),
			mcp.WithString(
				"json_string",
				mcp.Required(),
				mcp.Description("要格式化的 JSON 字符串"),
			),
			mcp.WithNumber(
				"indent",
				mcp.Description("缩进空格数（默认 2）"),
			),
		),
		formatJSONHandler,
	)

	// 工具 3: 项目信息
	s.AddTool(
		mcp.NewTool(
			"project_info",
			mcp.WithDescription("获取当前项目的基本信息"),
		),
		projectInfoHandler,
	)

	// ==================== 注册资源 ====================

	// 资源: Server 自身信息
	s.AddResource(
		mcp.Resource{
			URI:         "info://server/about",
			Name:        "Server 信息",
			Description: "关于此 MCP Server 的说明和使用指南",
			MIMEType:    "text/markdown",
		},
		serverInfoHandler,
	)

	// ==================== 启动 Server ====================

	// 输出启动信息到 stderr（不干扰 JSON-RPC 通信）
	fmt.Fprintln(os.Stderr, "============================================")
	fmt.Fprintln(os.Stderr, "Integration Demo MCP Server")
	fmt.Fprintln(os.Stderr, "版本: 1.0.0")
	fmt.Fprintln(os.Stderr, "============================================")
	fmt.Fprintln(os.Stderr, "已注册工具:")
	fmt.Fprintln(os.Stderr, "  - current_time: 获取当前时间")
	fmt.Fprintln(os.Stderr, "  - format_json: JSON 格式化")
	fmt.Fprintln(os.Stderr, "  - project_info: 项目信息")
	fmt.Fprintln(os.Stderr, "已注册资源:")
	fmt.Fprintln(os.Stderr, "  - info://server/about: Server 说明")
	fmt.Fprintln(os.Stderr, "============================================")
	fmt.Fprintln(os.Stderr, "等待 MCP Client 连接...")

	stdioServer := server.NewStdioServer(s)
	ctx := context.Background()
	if err := stdioServer.Listen(ctx, os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server 错误: %v", err)
	}
}

// currentTimeHandler 返回当前时间
func currentTimeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format := request.GetString("format", "simple")

	now := time.Now()
	var result string

	switch format {
	case "unix":
		result = fmt.Sprintf("Unix 时间戳: %d", now.Unix())
	case "full":
		result = fmt.Sprintf(
			"当前时间（完整格式）:\n"+
				"  日期: %s\n"+
				"  时间: %s\n"+
				"  时区: %s\n"+
				"  星期: %s\n"+
				"  今年第 %d 天\n"+
				"  Unix 时间戳: %d",
			now.Format("2006年01月02日"),
			now.Format("15:04:05"),
			now.Location().String(),
			translateWeekday(now.Weekday()),
			now.YearDay(),
			now.Unix(),
		)
	default: // simple
		result = fmt.Sprintf("当前时间: %s", now.Format("2006-01-02 15:04:05"))
	}

	return mcp.NewToolResultText(result), nil
}

// formatJSONHandler 格式化 JSON 字符串
func formatJSONHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jsonStr := request.GetString("json_string", "")
	if jsonStr == "" {
		return mcp.NewToolResultError("参数错误：json_string 不能为空"), nil
	}

	indent := request.GetInt("indent", 2)
	if indent <= 0 {
		indent = 2
	}
	if indent > 8 {
		indent = 8
	}

	// 解析 JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("JSON 解析错误: %v", err)), nil
	}

	// 格式化输出
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += " "
	}
	formatted, err := json.MarshalIndent(parsed, "", indentStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("JSON 格式化错误: %v", err)), nil
	}

	return mcp.NewToolResultText(string(formatted)), nil
}

// projectInfoHandler 返回项目信息
func projectInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "无法获取"
	}

	// 获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "无法获取"
	}

	info := fmt.Sprintf(
		"项目信息:\n"+
			"  工作目录: %s\n"+
			"  主机名: %s\n"+
			"  进程 ID: %d\n"+
			"  Server 名称: integration-demo\n"+
			"  Server 版本: 1.0.0\n"+
			"  启动时间: %s",
		cwd,
		hostname,
		os.Getpid(),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	return mcp.NewToolResultText(info), nil
}

// serverInfoHandler 返回 Server 的说明文档
func serverInfoHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	doc := `# Integration Demo MCP Server

## 概述
这是一个用于演示 MCP 集成的 Server，包含常用的实用工具。

## 可用工具

### current_time
获取当前的日期和时间。
- 参数 format: simple（默认）、full、unix

### format_json
格式化 JSON 字符串，使其更易读。
- 参数 json_string: 要格式化的 JSON（必填）
- 参数 indent: 缩进空格数（默认 2）

### project_info
获取当前项目和运行环境的基本信息。

## 集成方式

### Claude Desktop
将以下内容添加到 claude_desktop_config.json:
` + "```json" + `
{
  "mcpServers": {
    "integration-demo": {
      "command": "/path/to/this/server"
    }
  }
}
` + "```" + `

### Claude Code
运行以下命令:
` + "```bash" + `
claude mcp add integration-demo /path/to/this/server
` + "```" + `
`

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "info://server/about",
			MIMEType: "text/markdown",
			Text:     doc,
		},
	}, nil
}

// translateWeekday 将英文星期翻译为中文
func translateWeekday(day time.Weekday) string {
	weekdays := map[time.Weekday]string{
		time.Sunday:    "星期日",
		time.Monday:    "星期一",
		time.Tuesday:   "星期二",
		time.Wednesday: "星期三",
		time.Thursday:  "星期四",
		time.Friday:    "星期五",
		time.Saturday:  "星期六",
	}
	return weekdays[day]
}
