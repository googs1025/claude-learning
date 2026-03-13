// 计算器 MCP Server
// 编译方式: go build -o calculator servers/calculator.go
//
// 本文件是一个完整的 MCP Server，提供四则运算工具：
//   add      - 加法
//   subtract - 减法
//   multiply - 乘法
//   divide   - 除法
//
// 通过 stdio 传输，供 Claude Code / Claude Desktop 调用

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer("calculator", "1.0.0")

	// ===== 加法工具 =====
	s.AddTool(
		mcp.NewTool("add",
			mcp.WithDescription("计算两个数字的和"),
			mcp.WithNumber("a", mcp.Required(), mcp.Description("第一个数字")),
			mcp.WithNumber("b", mcp.Required(), mcp.Description("第二个数字")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			a := req.GetFloat("a", 0)
			b := req.GetFloat("b", 0)
			return mcp.NewToolResultText(fmt.Sprintf("%.6g + %.6g = %.6g", a, b, a+b)), nil
		},
	)

	// ===== 减法工具 =====
	s.AddTool(
		mcp.NewTool("subtract",
			mcp.WithDescription("计算两个数字的差"),
			mcp.WithNumber("a", mcp.Required(), mcp.Description("被减数")),
			mcp.WithNumber("b", mcp.Required(), mcp.Description("减数")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			a := req.GetFloat("a", 0)
			b := req.GetFloat("b", 0)
			return mcp.NewToolResultText(fmt.Sprintf("%.6g - %.6g = %.6g", a, b, a-b)), nil
		},
	)

	// ===== 乘法工具 =====
	s.AddTool(
		mcp.NewTool("multiply",
			mcp.WithDescription("计算两个数字的积"),
			mcp.WithNumber("a", mcp.Required(), mcp.Description("第一个数字")),
			mcp.WithNumber("b", mcp.Required(), mcp.Description("第二个数字")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			a := req.GetFloat("a", 0)
			b := req.GetFloat("b", 0)
			return mcp.NewToolResultText(fmt.Sprintf("%.6g × %.6g = %.6g", a, b, a*b)), nil
		},
	)

	// ===== 除法工具 =====
	s.AddTool(
		mcp.NewTool("divide",
			mcp.WithDescription("计算两个数字的商"),
			mcp.WithNumber("a", mcp.Required(), mcp.Description("被除数")),
			mcp.WithNumber("b", mcp.Required(), mcp.Description("除数（不能为 0）")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			a := req.GetFloat("a", 0)
			b := req.GetFloat("b", 0)
			if b == 0 {
				return mcp.NewToolResultText("错误：除数不能为零"), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("%.6g ÷ %.6g = %.6g", a, b, a/b)), nil
		},
	)

	fmt.Fprintln(os.Stderr, "Calculator MCP Server 已启动（stdio 模式）")
	stdioServer := server.NewStdioServer(s)
	if err := stdioServer.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server 错误: %v", err)
	}
}
