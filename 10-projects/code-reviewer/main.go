// 运行方式: go run main.go <文件路径>
// 综合项目2：代码审查助手
// 功能：读取源代码文件，使用 Claude + Tool Use 进行结构化审查

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// ReviewResult 审查结果的结构化格式
type ReviewResult struct {
	OverallScore int           `json:"overall_score"` // 1-10 综合评分
	Summary      string        `json:"summary"`       // 总体评价
	Issues       []ReviewIssue `json:"issues"`         // 发现的问题列表
	Suggestions  []string      `json:"suggestions"`    // 改进建议
}

// ReviewIssue 单个审查问题
type ReviewIssue struct {
	Severity    string `json:"severity"`    // critical/warning/info
	Category    string `json:"category"`    // security/performance/style/logic
	Line        int    `json:"line"`        // 行号（大约）
	Description string `json:"description"` // 问题描述
	Fix         string `json:"fix"`         // 修复建议
}

// 定义审查工具
var reviewTool = []anthropic.ToolUnionParam{
	{
		OfTool: &anthropic.ToolParam{
			Name:        "submit_review",
			Description: anthropic.String("提交代码审查结果，包含评分、问题列表和改进建议"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"overall_score": map[string]any{
						"type":        "integer",
						"description": "综合评分 1-10，10 为最好",
						"minimum":     1,
						"maximum":     10,
					},
					"summary": map[string]any{
						"type":        "string",
						"description": "总体评价，2-3句话概括代码质量",
					},
					"issues": map[string]any{
						"type":        "array",
						"description": "发现的问题列表",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"severity": map[string]any{
									"type": "string",
									"enum": []string{"critical", "warning", "info"},
								},
								"category": map[string]any{
									"type": "string",
									"enum": []string{"security", "performance", "style", "logic", "error_handling"},
								},
								"line":        map[string]any{"type": "integer", "description": "大约行号"},
								"description": map[string]any{"type": "string"},
								"fix":         map[string]any{"type": "string"},
							},
						},
					},
					"suggestions": map[string]any{
						"type":        "array",
						"description": "改进建议列表",
						"items":       map[string]any{"type": "string"},
					},
				},
				Required: []string{"overall_score", "summary", "issues", "suggestions"},
			},
		},
	},
}

// 严重程度对应的图标
var severityIcons = map[string]string{
	"critical": "🔴",
	"warning":  "🟡",
	"info":     "🔵",
}

// 分类对应的图标
var categoryIcons = map[string]string{
	"security":       "🔒",
	"performance":    "⚡",
	"style":          "🎨",
	"logic":          "🧠",
	"error_handling": "⚠️",
}

func reviewCode(filePath string) error {
	// 读取源代码文件
	code, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("无法读取文件 %s: %w", filePath, err)
	}

	fmt.Printf("📄 正在审查文件: %s (%d 字节)\n", filePath, len(code))
	fmt.Println(strings.Repeat("=", 50))

	client := anthropic.NewClient()
	ctx := context.Background()

	// 系统提示：定义审查标准
	systemPrompt := `你是一位资深的 Go 语言代码审查专家。请从以下维度审查代码：

1. **安全性 (security)**: SQL 注入、命令注入、XSS、敏感信息泄露等
2. **性能 (performance)**: 不必要的内存分配、goroutine 泄漏、低效的算法等
3. **代码风格 (style)**: 命名规范、注释质量、代码组织等
4. **逻辑 (logic)**: 边界条件、空指针、并发安全等
5. **错误处理 (error_handling)**: 错误是否被正确处理、是否有遗漏

请使用 submit_review 工具提交你的审查结果。评分标准：
- 9-10: 优秀，几乎没有问题
- 7-8: 良好，有小问题
- 5-6: 一般，有明显问题需要修复
- 3-4: 较差，有严重问题
- 1-2: 很差，需要重写`

	// 发送审查请求，强制使用工具
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 4096,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				fmt.Sprintf("请审查以下 Go 代码:\n\n```go\n%s\n```", string(code)),
			)),
		},
		Tools: reviewTool,
		ToolChoice: anthropic.ToolChoiceUnionParam{
			OfTool: &anthropic.ToolChoiceToolParam{Name: "submit_review"},
		},
	})
	if err != nil {
		return fmt.Errorf("API 调用失败: %w", err)
	}

	// 解析工具调用结果
	for _, block := range message.Content {
		if toolUse, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
			var result ReviewResult
			if err := json.Unmarshal(toolUse.Input, &result); err != nil {
				return fmt.Errorf("解析结果失败: %w", err)
			}

			// 打印审查报告
			printReport(result)
		}
	}

	return nil
}

func printReport(result ReviewResult) {
	// 评分颜色
	scoreBar := strings.Repeat("█", result.OverallScore) + strings.Repeat("░", 10-result.OverallScore)

	fmt.Printf("\n📊 综合评分: [%s] %d/10\n", scoreBar, result.OverallScore)
	fmt.Printf("\n📝 总体评价:\n   %s\n", result.Summary)

	// 问题列表
	if len(result.Issues) > 0 {
		fmt.Printf("\n🔍 发现 %d 个问题:\n", len(result.Issues))
		fmt.Println(strings.Repeat("-", 50))

		for i, issue := range result.Issues {
			sevIcon := severityIcons[issue.Severity]
			catIcon := categoryIcons[issue.Category]
			if sevIcon == "" {
				sevIcon = "⚪"
			}
			if catIcon == "" {
				catIcon = "📌"
			}

			fmt.Printf("\n  %s %s 问题 #%d (行 %d) [%s/%s]\n",
				sevIcon, catIcon, i+1, issue.Line, issue.Severity, issue.Category)
			fmt.Printf("     描述: %s\n", issue.Description)
			fmt.Printf("     修复: %s\n", issue.Fix)
		}
	} else {
		fmt.Println("\n✅ 未发现问题!")
	}

	// 改进建议
	if len(result.Suggestions) > 0 {
		fmt.Printf("\n💡 改进建议:\n")
		for i, suggestion := range result.Suggestions {
			fmt.Printf("   %d. %s\n", i+1, suggestion)
		}
	}

	fmt.Println(strings.Repeat("=", 50))
}

func main() {
	if len(os.Args) < 2 {
		// 没有参数时，审查自身作为演示
		fmt.Println("用法: go run main.go <Go源码文件路径>")
		fmt.Println("示例: go run main.go main.go")
		fmt.Println("\n未指定文件，将审查自身代码作为演示...")
		if err := reviewCode("main.go"); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
		return
	}

	filePath := os.Args[1]
	if err := reviewCode(filePath); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
