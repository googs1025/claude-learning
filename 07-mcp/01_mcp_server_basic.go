// 第七章示例 1：基础 MCP Server
// 运行方式: go run 01_mcp_server_basic.go
//
// 本示例演示如何：
// 1. 创建一个最简单的 MCP Server
// 2. 注册一个 "greet"（问候）工具
// 3. 使用 stdio 传输方式运行 Server
// 4. 处理工具调用请求并返回结果
//
// MCP Server 通过标准输入/输出（stdio）与 MCP Client 通信，
// 使用 JSON-RPC 2.0 协议交换消息。
//
// 测试方法：
//   1. 编译: go build -o greet-server 01_mcp_server_basic.go
//   2. 手动测试（发送初始化请求）:
//      echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":{},"clientInfo":{"name":"test","version":"1.0"},"protocolVersion":"2024-11-05"}}' | ./greet-server

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// ==================== 创建 MCP Server ====================
	// NewMCPServer 创建一个新的 MCP Server 实例
	// 参数 1: Server 名称（标识你的服务）
	// 参数 2: 版本号
	s := server.NewMCPServer(
		"greeting-server", // 服务器名称
		"1.0.0",           // 版本号
	)

	// ==================== 定义工具 ====================
	// 使用 mcp.NewTool 创建工具定义
	// 工具定义包括：名称、描述、参数 schema
	greetTool := mcp.NewTool(
		"greet", // 工具名称，Client 用这个名称来调用
		mcp.WithDescription("向指定的人发送问候语"), // 工具描述，帮助 AI 理解何时使用此工具
		mcp.WithString(
			"name",                             // 参数名称
			mcp.Required(),                     // 标记为必填参数
			mcp.Description("要问候的人的名字"), // 参数描述
		),
		mcp.WithString(
			"language",                                      // 可选参数：问候语言
			mcp.Description("问候使用的语言（zh/en/ja）"),   // 参数描述
		),
	)

	// ==================== 注册工具和处理函数 ====================
	// AddTool 将工具定义与处理函数绑定
	// 当 Client 调用 "greet" 工具时，greetHandler 函数会被执行
	s.AddTool(greetTool, greetHandler)

	// ==================== 启动 Stdio Server ====================
	// NewStdioServer 创建一个基于标准输入/输出的传输层
	// 这是与 Claude Desktop / Claude Code 集成的标准方式
	stdioServer := server.NewStdioServer(s)

	// 日志输出到 stderr，避免干扰 stdout 上的 JSON-RPC 通信
	fmt.Fprintln(os.Stderr, "Greeting MCP Server 已启动")
	fmt.Fprintln(os.Stderr, "等待 MCP Client 连接...")

	// Listen 开始监听 stdin 输入并通过 stdout 返回响应
	// 这个调用会阻塞，直到 stdin 关闭或发生错误
	ctx := context.Background()
	if err := stdioServer.Listen(ctx, os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server 错误: %v", err)
	}

	fmt.Fprintln(os.Stderr, "Server 已停止")
}

// greetHandler 处理 "greet" 工具的调用请求
//
// 参数:
//   - ctx: 上下文，可用于超时控制
//   - request: 工具调用请求，包含客户端传入的参数
//
// 返回:
//   - *mcp.CallToolResult: 工具执行结果，包含返回给 AI 的文本
//   - error: 如果处理过程中发生错误
func greetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 从请求中提取参数
	// 使用 helper 方法安全地获取参数
	name := request.GetString("name", "")
	if name == "" {
		return mcp.NewToolResultError("参数错误：name 是必填参数"), nil
	}

	// 获取可选参数 language，默认为中文
	language := request.GetString("language", "zh")

	// 根据语言生成问候语
	var greeting string
	switch language {
	case "zh":
		greeting = fmt.Sprintf("你好，%s！欢迎使用 MCP Server。当前时间：%s",
			name, time.Now().Format("2006-01-02 15:04:05"))
	case "en":
		greeting = fmt.Sprintf("Hello, %s! Welcome to MCP Server. Current time: %s",
			name, time.Now().Format("2006-01-02 15:04:05"))
	case "ja":
		greeting = fmt.Sprintf("こんにちは、%sさん！MCP Server へようこそ。現在時刻：%s",
			name, time.Now().Format("2006-01-02 15:04:05"))
	default:
		greeting = fmt.Sprintf("Hello, %s! (Unsupported language: %s, using English)",
			name, language)
	}

	// 返回工具执行结果
	// NewToolResultText 创建一个包含文本内容的成功结果
	return mcp.NewToolResultText(greeting), nil
}
