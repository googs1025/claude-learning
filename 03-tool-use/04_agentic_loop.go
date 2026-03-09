// 第三章示例 4：Agentic 工具循环 —— 知识库查询助手
// 运行方式: go run 04_agentic_loop.go
//
// 本示例演示如何：
// 1. 实现完整的 agentic loop（代理循环）
// 2. Claude 自主决定调用工具的顺序和次数
// 3. 处理多轮工具调用直到 Claude 认为任务完成
// 4. 模拟一个知识库查询系统，包含搜索和详情查看两个工具
//
// Agentic Loop 是构建 AI Agent 的核心模式：
//   发送消息 → 检查是否需要工具 → 执行工具 → 返回结果 → 重复
//   直到 Claude 返回 end_turn，表示任务完成

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// ========================================
// 模拟知识库数据
// ========================================

// Article 代表知识库中的一篇文章
type Article struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Tags     []string `json:"tags"`
}

// 模拟的知识库
var knowledgeBase = []Article{
	{
		ID:       "kb001",
		Title:    "Go 语言并发模型",
		Summary:  "介绍 Go 语言的 goroutine 和 channel 并发模型",
		Content:  "Go 语言使用 goroutine 实现轻量级并发。goroutine 是由 Go 运行时管理的轻量级线程，创建成本极低（约 2KB 栈空间）。通过 channel 可以在 goroutine 之间安全地传递数据，实现 CSP（通信顺序进程）并发模型。使用 select 语句可以同时等待多个 channel 操作。",
		Category: "编程语言",
		Tags:     []string{"Go", "并发", "goroutine", "channel"},
	},
	{
		ID:       "kb002",
		Title:    "Python 异步编程",
		Summary:  "Python 的 asyncio 异步编程框架",
		Content:  "Python 3.5+ 引入了 async/await 语法，配合 asyncio 库实现异步编程。asyncio 使用事件循环（event loop）来调度协程（coroutine）。与多线程相比，异步编程避免了线程切换的开销，适合 I/O 密集型任务。常用的异步框架包括 aiohttp、FastAPI 等。",
		Category: "编程语言",
		Tags:     []string{"Python", "异步", "asyncio", "协程"},
	},
	{
		ID:       "kb003",
		Title:    "Docker 容器基础",
		Summary:  "Docker 容器技术入门指南",
		Content:  "Docker 是一个开源的容器化平台，允许开发者将应用程序及其依赖打包到一个轻量级、可移植的容器中。核心概念包括：镜像（Image）是只读模板，容器（Container）是镜像的运行实例，Dockerfile 定义镜像的构建步骤。Docker Compose 用于定义和运行多容器应用。",
		Category: "DevOps",
		Tags:     []string{"Docker", "容器", "DevOps", "部署"},
	},
	{
		ID:       "kb004",
		Title:    "Go 语言错误处理",
		Summary:  "Go 语言的错误处理机制和最佳实践",
		Content:  "Go 语言使用显式错误返回而非异常机制。函数通过返回 error 接口来表示错误。最佳实践包括：使用 errors.New() 或 fmt.Errorf() 创建错误，使用 errors.Is() 和 errors.As() 进行错误比较，使用 %w 包装错误以保留错误链。Go 1.13 引入了错误包装，使错误处理更加灵活。",
		Category: "编程语言",
		Tags:     []string{"Go", "错误处理", "最佳实践"},
	},
	{
		ID:       "kb005",
		Title:    "Kubernetes 入门",
		Summary:  "Kubernetes 容器编排平台概述",
		Content:  "Kubernetes（K8s）是一个开源的容器编排平台，用于自动化容器的部署、扩展和管理。核心概念：Pod 是最小部署单元，Deployment 管理 Pod 的副本，Service 提供网络访问，Ingress 管理外部访问。K8s 支持自动扩缩容、滚动更新、服务发现等功能。",
		Category: "DevOps",
		Tags:     []string{"Kubernetes", "K8s", "容器编排", "DevOps"},
	},
}

// searchKnowledgeBase 搜索知识库
// 根据关键词在标题、摘要、标签中搜索匹配的文章
func searchKnowledgeBase(input json.RawMessage) (string, bool) {
	var params struct {
		Query    string `json:"query"`
		Category string `json:"category"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return fmt.Sprintf("参数解析失败: %v", err), true
	}

	query := strings.ToLower(params.Query)
	var results []map[string]string

	for _, article := range knowledgeBase {
		// 检查分类过滤
		if params.Category != "" && !strings.EqualFold(article.Category, params.Category) {
			continue
		}

		// 在标题、摘要、标签中搜索关键词
		matched := strings.Contains(strings.ToLower(article.Title), query) ||
			strings.Contains(strings.ToLower(article.Summary), query)

		if !matched {
			for _, tag := range article.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					matched = true
					break
				}
			}
		}

		if matched {
			results = append(results, map[string]string{
				"id":       article.ID,
				"title":    article.Title,
				"summary":  article.Summary,
				"category": article.Category,
			})
		}
	}

	if len(results) == 0 {
		return fmt.Sprintf("未找到与 '%s' 相关的文章", params.Query), false
	}

	data, _ := json.MarshalIndent(results, "", "  ")
	return string(data), false
}

// getArticleDetail 获取文章详情
// 根据文章 ID 返回完整内容
func getArticleDetail(input json.RawMessage) (string, bool) {
	var params struct {
		ArticleID string `json:"article_id"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return fmt.Sprintf("参数解析失败: %v", err), true
	}

	for _, article := range knowledgeBase {
		if article.ID == params.ArticleID {
			data, _ := json.MarshalIndent(article, "", "  ")
			return string(data), false
		}
	}

	return fmt.Sprintf("未找到 ID 为 '%s' 的文章", params.ArticleID), true
}

// executeTool 根据工具名称分发执行
func executeTool(name string, input json.RawMessage) (string, bool) {
	switch name {
	case "search_knowledge_base":
		return searchKnowledgeBase(input)
	case "get_article_detail":
		return getArticleDetail(input)
	default:
		return fmt.Sprintf("未知工具: %s", name), true
	}
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// ========================================
	// 定义知识库工具
	// ========================================
	tools := []anthropic.ToolUnionParam{
		// 工具 1：搜索知识库
		{
			OfTool: &anthropic.ToolParam{
				Name:        "search_knowledge_base",
				Description: anthropic.String("在知识库中搜索文章。根据关键词搜索标题、摘要和标签。返回匹配文章的ID、标题和摘要列表。"),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "搜索关键词，如 'Go并发'、'Docker'、'错误处理'",
						},
						"category": map[string]any{
							"type":        "string",
							"description": "可选的分类过滤，如 '编程语言'、'DevOps'",
						},
					},
					Required: []string{"query"},
				},
			},
		},
		// 工具 2：获取文章详情
		{
			OfTool: &anthropic.ToolParam{
				Name:        "get_article_detail",
				Description: anthropic.String("根据文章ID获取文章的完整内容。需要先通过搜索获取文章ID。"),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"article_id": map[string]any{
							"type":        "string",
							"description": "文章的唯一标识符，如 'kb001'",
						},
					},
					Required: []string{"article_id"},
				},
			},
		},
	}

	// ========================================
	// 发送一个需要多步工具调用的问题
	// ========================================
	// 这个问题需要 Claude：
	// 1. 搜索 Go 相关文章
	// 2. 查看感兴趣的文章详情
	// 3. 综合信息给出回答
	userQuestion := "请帮我找一下知识库中关于 Go 语言的文章，我想了解 Go 的并发模型和错误处理的详细内容。"
	fmt.Printf("用户: %s\n\n", userQuestion)

	// 初始化消息列表
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userQuestion)),
	}

	// ========================================
	// 核心：Agentic Loop（代理循环）
	// ========================================
	// 这是构建 AI Agent 的核心模式
	// 循环逻辑：
	//   1. 发送消息给 Claude
	//   2. 检查 stop_reason
	//      - "end_turn" → 任务完成，退出循环
	//      - "tool_use" → 执行工具，将结果加入消息，继续循环
	//   3. 重复直到完成

	maxRounds := 10 // 设置最大轮数，防止无限循环
	round := 0

	for round < maxRounds {
		round++
		fmt.Printf("========== 第 %d 轮 ==========\n", round)

		// 1. 发送消息给 Claude
		message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_5_20250929,
			MaxTokens: 4096,
			Tools:     tools,
			Messages:  messages,
		})
		if err != nil {
			log.Fatalf("第 %d 轮 API 调用失败: %v", round, err)
		}

		fmt.Printf("停止原因: %s\n", message.StopReason)

		// 2. 检查是否任务完成
		if message.StopReason == "end_turn" {
			// Claude 认为任务完成，输出最终回复
			fmt.Println("\n========== Claude 的最终回复 ==========")
			for _, block := range message.Content {
				if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
					fmt.Println(tb.Text)
				}
			}
			fmt.Printf("\n共经历了 %d 轮工具调用\n", round)
			return // 退出程序
		}

		// 3. stop_reason 为 "tool_use"，需要执行工具
		// 将 Claude 的响应加入消息历史
		messages = append(messages, message.ToParam())

		// 4. 遍历所有内容块，执行工具调用
		var toolResults []anthropic.ContentBlockParamUnion

		for _, block := range message.Content {
			switch v := block.AsAny().(type) {
			case anthropic.TextBlock:
				// Claude 可能在工具调用前输出思考文本
				fmt.Printf("[Claude 思考] %s\n", v.Text)

			case anthropic.ToolUseBlock:
				// 解析并执行工具
				rawInput := json.RawMessage(v.JSON.Input.Raw())
				fmt.Printf("[调用工具] %s\n", v.Name)
				fmt.Printf("  参数: %s\n", string(rawInput))

				// 执行工具
				result, isError := executeTool(v.Name, rawInput)
				if isError {
					fmt.Printf("  错误: %s\n", result)
				} else {
					// 截断过长的结果用于显示
					displayResult := result
					if len(displayResult) > 200 {
						displayResult = displayResult[:200] + "..."
					}
					fmt.Printf("  结果: %s\n", displayResult)
				}

				// 创建工具结果
				toolResults = append(toolResults,
					anthropic.NewToolResultBlock(v.ID, result, isError),
				)
			}
		}

		// 5. 将所有工具结果发回给 Claude
		if len(toolResults) > 0 {
			messages = append(messages,
				anthropic.NewUserMessage(toolResults...),
			)
		}

		fmt.Println() // 轮次之间加空行
	}

	// 如果达到最大轮数仍未完成
	fmt.Printf("警告: 达到最大轮数 %d，循环终止\n", maxRounds)
}
