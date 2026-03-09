// 第七章示例 2：MCP 工具定义
// 运行方式: go run 02_mcp_tools.go
//
// 本示例演示如何：
// 1. 定义多种工具，包含复杂的参数 schema
// 2. 实现计算器工具（支持四则运算）
// 3. 实现文件信息工具（读取文件元数据）
// 4. 实现文本分析工具（统计字数、行数等）
// 5. 进行参数验证和错误处理
//
// 测试方法：
//   go build -o tools-server 02_mcp_tools.go

package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 创建 MCP Server
	s := server.NewMCPServer(
		"tools-demo-server",
		"1.0.0",
	)

	// ==================== 工具 1：计算器 ====================
	// 支持加、减、乘、除四则运算
	calculatorTool := mcp.NewTool(
		"calculator",
		mcp.WithDescription("执行基本数学运算（加、减、乘、除、乘方）"),
		mcp.WithString(
			"operation",
			mcp.Required(),
			mcp.Description("运算类型：add（加）、subtract（减）、multiply（乘）、divide（除）、power（乘方）"),
			mcp.Enum("add", "subtract", "multiply", "divide", "power"), // 限定可选值
		),
		mcp.WithNumber(
			"a",
			mcp.Required(),
			mcp.Description("第一个操作数"),
		),
		mcp.WithNumber(
			"b",
			mcp.Required(),
			mcp.Description("第二个操作数"),
		),
	)
	s.AddTool(calculatorTool, calculatorHandler)

	// ==================== 工具 2：文件信息 ====================
	// 读取文件的基本信息（大小、修改时间等）
	fileInfoTool := mcp.NewTool(
		"file_info",
		mcp.WithDescription("获取指定文件的基本信息（大小、修改时间、权限等）"),
		mcp.WithString(
			"path",
			mcp.Required(),
			mcp.Description("文件的绝对路径或相对路径"),
		),
	)
	s.AddTool(fileInfoTool, fileInfoHandler)

	// ==================== 工具 3：文本分析 ====================
	// 分析文本的基本统计信息
	textAnalysisTool := mcp.NewTool(
		"text_analysis",
		mcp.WithDescription("分析文本内容，返回字符数、单词数、行数等统计信息"),
		mcp.WithString(
			"text",
			mcp.Required(),
			mcp.Description("要分析的文本内容"),
		),
		mcp.WithBoolean(
			"include_frequency",
			mcp.Description("是否包含字符频率统计（默认 false）"),
		),
	)
	s.AddTool(textAnalysisTool, textAnalysisHandler)

	// ==================== 工具 4：字符串工具 ====================
	// 常用字符串操作
	stringTool := mcp.NewTool(
		"string_utils",
		mcp.WithDescription("常用字符串操作工具"),
		mcp.WithString(
			"action",
			mcp.Required(),
			mcp.Description("操作类型：reverse（反转）、upper（大写）、lower（小写）、repeat（重复）"),
			mcp.Enum("reverse", "upper", "lower", "repeat"),
		),
		mcp.WithString(
			"text",
			mcp.Required(),
			mcp.Description("要处理的文本"),
		),
		mcp.WithNumber(
			"count",
			mcp.Description("重复次数（仅 repeat 操作使用，默认为 2）"),
		),
	)
	s.AddTool(stringTool, stringUtilsHandler)

	// 启动 Stdio Server
	fmt.Fprintln(os.Stderr, "Tools Demo MCP Server 已启动")
	fmt.Fprintln(os.Stderr, "已注册工具: calculator, file_info, text_analysis, string_utils")

	stdioServer := server.NewStdioServer(s)
	ctx := context.Background()
	if err := stdioServer.Listen(ctx, os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server 错误: %v", err)
	}
}

// calculatorHandler 处理计算器工具的调用
func calculatorHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 提取参数
	operation := request.GetString("operation", "")
	a := request.GetFloat("a", math.NaN())
	if math.IsNaN(a) {
		return mcp.NewToolResultError("参数错误：a 必须是数字"), nil
	}
	b := request.GetFloat("b", math.NaN())
	if math.IsNaN(b) {
		return mcp.NewToolResultError("参数错误：b 必须是数字"), nil
	}

	var result float64
	var description string

	switch operation {
	case "add":
		result = a + b
		description = fmt.Sprintf("%.6g + %.6g = %.6g", a, b, result)
	case "subtract":
		result = a - b
		description = fmt.Sprintf("%.6g - %.6g = %.6g", a, b, result)
	case "multiply":
		result = a * b
		description = fmt.Sprintf("%.6g × %.6g = %.6g", a, b, result)
	case "divide":
		if b == 0 {
			return mcp.NewToolResultError("错误：除数不能为零"), nil
		}
		result = a / b
		description = fmt.Sprintf("%.6g ÷ %.6g = %.6g", a, b, result)
	case "power":
		result = math.Pow(a, b)
		description = fmt.Sprintf("%.6g ^ %.6g = %.6g", a, b, result)
	default:
		return mcp.NewToolResultError(
			fmt.Sprintf("不支持的运算: %s，可选: add, subtract, multiply, divide, power", operation),
		), nil
	}

	return mcp.NewToolResultText(description), nil
}

// fileInfoHandler 处理文件信息工具的调用
func fileInfoHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := request.GetString("path", "")
	if path == "" {
		return mcp.NewToolResultError("参数错误：path 不能为空"), nil
	}

	// 获取文件信息
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("文件不存在: %s", path)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("无法读取文件信息: %v", err)), nil
	}

	// 格式化文件大小
	size := info.Size()
	sizeStr := formatFileSize(size)

	// 构建结果文本
	result := fmt.Sprintf(
		"文件信息:\n"+
			"  名称: %s\n"+
			"  大小: %s (%d 字节)\n"+
			"  修改时间: %s\n"+
			"  权限: %s\n"+
			"  是否目录: %v",
		info.Name(),
		sizeStr, size,
		info.ModTime().Format("2006-01-02 15:04:05"),
		info.Mode().String(),
		info.IsDir(),
	)

	return mcp.NewToolResultText(result), nil
}

// textAnalysisHandler 处理文本分析工具的调用
func textAnalysisHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text := request.GetString("text", "")
	if text == "" {
		return mcp.NewToolResultError("参数错误：text 不能为空"), nil
	}

	includeFrequency := request.GetBool("include_frequency", false)

	// 基本统计
	chars := len([]rune(text))
	bytes := len(text)
	lines := strings.Count(text, "\n") + 1
	words := len(strings.Fields(text))
	sentences := strings.Count(text, "。") + strings.Count(text, "！") +
		strings.Count(text, "？") + strings.Count(text, ".") +
		strings.Count(text, "!") + strings.Count(text, "?")

	var sb strings.Builder
	sb.WriteString("文本分析结果:\n")
	sb.WriteString(fmt.Sprintf("  字符数: %d\n", chars))
	sb.WriteString(fmt.Sprintf("  字节数: %d\n", bytes))
	sb.WriteString(fmt.Sprintf("  行数: %d\n", lines))
	sb.WriteString(fmt.Sprintf("  词/字数: %d\n", words))
	sb.WriteString(fmt.Sprintf("  句子数: %d\n", sentences))

	// 可选：字符频率统计
	if includeFrequency {
		sb.WriteString("\n  字符频率（前 10）:\n")
		freq := make(map[rune]int)
		for _, r := range text {
			if r != ' ' && r != '\n' && r != '\t' {
				freq[r]++
			}
		}

		// 简单排序：取频率最高的 10 个字符
		type charFreq struct {
			char  rune
			count int
		}
		sorted := make([]charFreq, 0, len(freq))
		for r, c := range freq {
			sorted = append(sorted, charFreq{r, c})
		}
		// 冒泡排序（简单实现）
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].count > sorted[i].count {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		limit := 10
		if len(sorted) < limit {
			limit = len(sorted)
		}
		for _, cf := range sorted[:limit] {
			sb.WriteString(fmt.Sprintf("    '%c': %d 次\n", cf.char, cf.count))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// stringUtilsHandler 处理字符串工具的调用
func stringUtilsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := request.GetString("action", "")
	text := request.GetString("text", "")
	if text == "" {
		return mcp.NewToolResultError("参数错误：text 不能为空"), nil
	}

	var result string

	switch action {
	case "reverse":
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = fmt.Sprintf("反转结果: %s", string(runes))

	case "upper":
		result = fmt.Sprintf("大写结果: %s", strings.ToUpper(text))

	case "lower":
		result = fmt.Sprintf("小写结果: %s", strings.ToLower(text))

	case "repeat":
		count := request.GetInt("count", 2)
		if count <= 0 {
			count = 2
		}
		if count > 100 {
			return mcp.NewToolResultError("重复次数不能超过 100"), nil
		}
		result = fmt.Sprintf("重复结果（%d 次）: %s", count, strings.Repeat(text, count))

	default:
		return mcp.NewToolResultError(
			fmt.Sprintf("不支持的操作: %s，可选: reverse, upper, lower, repeat", action),
		), nil
	}

	return mcp.NewToolResultText(result), nil
}

// formatFileSize 将字节数格式化为人类可读的大小
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}
