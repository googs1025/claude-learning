// 运行方式: go run main.go
// 综合项目3：RAG 问答助手
// 功能：加载文档 → 文本分块 → 简单检索 → 基于上下文回答问题

package main

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"unicode"

	"github.com/anthropics/anthropic-sdk-go"
)

// =====================
// 文档与分块
// =====================

// Chunk 表示一个文本块
type Chunk struct {
	ID      int
	Content string
	Source  string // 来源标识
}

// KnowledgeBase 知识库
type KnowledgeBase struct {
	Chunks []Chunk
}

// NewKnowledgeBase 创建知识库
func NewKnowledgeBase() *KnowledgeBase {
	return &KnowledgeBase{
		Chunks: make([]Chunk, 0),
	}
}

// AddDocument 将文档分块后加入知识库
func (kb *KnowledgeBase) AddDocument(content string, source string, chunkSize int) {
	// 按段落分块
	paragraphs := strings.Split(content, "\n\n")
	var currentChunk strings.Builder
	chunkID := len(kb.Chunks)

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// 如果当前块加上新段落不超过限制，合并
		if currentChunk.Len()+len(para) < chunkSize {
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n\n")
			}
			currentChunk.WriteString(para)
		} else {
			// 保存当前块，开始新块
			if currentChunk.Len() > 0 {
				kb.Chunks = append(kb.Chunks, Chunk{
					ID:      chunkID,
					Content: currentChunk.String(),
					Source:  source,
				})
				chunkID++
			}
			currentChunk.Reset()
			currentChunk.WriteString(para)
		}
	}

	// 保存最后一个块
	if currentChunk.Len() > 0 {
		kb.Chunks = append(kb.Chunks, Chunk{
			ID:      chunkID,
			Content: currentChunk.String(),
			Source:  source,
		})
	}
}

// Search 简单的关键词搜索（实际项目中应使用向量搜索）
// 使用 TF 相似度对分块进行排序
func (kb *KnowledgeBase) Search(query string, topK int) []Chunk {
	// 提取查询关键词
	queryWords := tokenize(query)

	type scoredChunk struct {
		chunk Chunk
		score float64
	}

	var scored []scoredChunk
	for _, chunk := range kb.Chunks {
		chunkWords := tokenize(chunk.Content)
		score := computeRelevance(queryWords, chunkWords)
		if score > 0 {
			scored = append(scored, scoredChunk{chunk: chunk, score: score})
		}
	}

	// 按分数降序排序
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// 返回 top-K 结果
	results := make([]Chunk, 0, topK)
	for i := 0; i < len(scored) && i < topK; i++ {
		results = append(results, scored[i].chunk)
	}
	return results
}

// tokenize 简单分词
func tokenize(text string) []string {
	text = strings.ToLower(text)
	var words []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

// computeRelevance 计算查询与文档的相关度
func computeRelevance(queryWords, docWords []string) float64 {
	if len(queryWords) == 0 || len(docWords) == 0 {
		return 0
	}

	// 构建文档词频
	docFreq := make(map[string]int)
	for _, w := range docWords {
		docFreq[w]++
	}

	// 计算匹配分数
	var score float64
	for _, qw := range queryWords {
		if count, ok := docFreq[qw]; ok {
			// TF 分数 + 长度归一化
			score += math.Log(1+float64(count)) / math.Log(1+float64(len(docWords)))
		}
	}

	return score
}

// =====================
// RAG 助手
// =====================

// RAGAssistant RAG 问答助手
type RAGAssistant struct {
	client anthropic.Client
	kb     *KnowledgeBase
}

// NewRAGAssistant 创建 RAG 助手
func NewRAGAssistant() *RAGAssistant {
	return &RAGAssistant{
		client: anthropic.NewClient(),
		kb:     NewKnowledgeBase(),
	}
}

// Answer 根据知识库回答问题
func (ra *RAGAssistant) Answer(ctx context.Context, question string) error {
	// 第1步：检索相关文档
	fmt.Println("🔍 正在检索相关文档...")
	chunks := ra.kb.Search(question, 3)

	if len(chunks) == 0 {
		fmt.Println("⚠️  未找到相关文档，将基于通用知识回答。")
	}

	// 第2步：构建增强的 prompt
	var contextBuilder strings.Builder
	contextBuilder.WriteString("以下是与问题相关的参考资料：\n\n")
	for i, chunk := range chunks {
		contextBuilder.WriteString(fmt.Sprintf("--- 参考资料 %d (来源: %s) ---\n", i+1, chunk.Source))
		contextBuilder.WriteString(chunk.Content)
		contextBuilder.WriteString("\n\n")
	}

	fmt.Printf("📚 找到 %d 个相关文档块\n", len(chunks))

	// 第3步：调用 Claude 回答
	systemPrompt := `你是一个基于文档的问答助手。请根据提供的参考资料来回答问题。

规则：
1. 优先使用参考资料中的信息来回答
2. 如果参考资料中没有相关信息，请明确说明
3. 在回答中标注信息来源，格式如 [来源: xxx]
4. 回答要准确、简洁`

	message, err := ra.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 2048,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				fmt.Sprintf("%s\n\n问题: %s", contextBuilder.String(), question),
			)),
		},
	})
	if err != nil {
		return fmt.Errorf("API 调用失败: %w", err)
	}

	// 显示回答
	fmt.Println("\n💬 回答:")
	for _, block := range message.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			fmt.Println(tb.Text)
		}
	}
	fmt.Printf("\n[tokens: 输入=%d 输出=%d]\n",
		message.Usage.InputTokens, message.Usage.OutputTokens)

	return nil
}

// =====================
// 示例知识库数据
// =====================

func loadSampleDocuments(kb *KnowledgeBase) {
	// 文档1：Go 语言基础
	kb.AddDocument(`Go 语言简介

Go（又称 Golang）是 Google 开发的开源编程语言。它由 Robert Griesemer、Rob Pike 和 Ken Thompson 于 2007 年开始设计，2009 年正式发布。

Go 语言的核心特点包括：
- 静态类型和编译型语言，编译速度极快
- 内置垃圾回收（GC），开发者无需手动管理内存
- 原生支持并发编程，通过 goroutine 和 channel 实现
- 简洁的语法设计，关键字只有 25 个
- 强大的标准库，包含 HTTP、JSON、加密等常用功能

Go 的并发模型

Go 使用 goroutine 实现轻量级并发。goroutine 是由 Go 运行时管理的轻量级线程，创建成本极低（约 2KB 栈空间）。

Channel 是 goroutine 之间通信的管道，遵循 CSP（Communicating Sequential Processes）模型。可以是有缓冲的或无缓冲的。

select 语句可以同时等待多个 channel 操作，类似于 switch 但用于通道。`, "go_basics", 500)

	// 文档2：Claude API
	kb.AddDocument(`Claude API 使用指南

Claude 是 Anthropic 开发的 AI 助手。通过 API 可以在应用程序中集成 Claude 的能力。

API 认证
使用 API 需要一个 API Key，可以在 console.anthropic.com 获取。所有请求需要在 Header 中包含 x-api-key。

Messages API
Messages API 是与 Claude 交互的主要接口。核心参数包括：
- model: 指定使用的模型（如 claude-sonnet-4-5-20250929）
- max_tokens: 最大输出 token 数
- messages: 对话消息列表
- system: 系统提示（可选）
- tools: 工具定义（可选）

Tool Use 功能
Claude 支持函数调用（Tool Use），可以定义工具让 Claude 在需要时调用。工作流程：
1. 在请求中定义可用工具
2. Claude 分析是否需要使用工具
3. 如果需要，返回工具调用请求
4. 开发者执行工具并返回结果
5. Claude 基于结果生成最终回答

流式响应
使用 Server-Sent Events（SSE）可以实现流式输出，适合实时显示 Claude 的响应。`, "claude_api", 500)

	// 文档3：MCP 协议
	kb.AddDocument(`MCP（Model Context Protocol）

MCP 是 Anthropic 提出的开放协议，旨在标准化 AI 模型与外部工具和数据源之间的连接方式。

MCP 的核心概念：
- Server: 提供工具和资源的服务端
- Client: 连接 Server 的客户端（如 Claude Code、Claude Desktop）
- Tools: Server 暴露的可调用功能
- Resources: Server 暴露的只读数据
- Prompts: Server 提供的预定义提示模板

传输方式：
MCP 支持多种传输协议：
- stdio: 通过标准输入/输出通信，适合本地进程
- SSE: Server-Sent Events，适合远程 HTTP 通信
- Streamable HTTP: 新的 HTTP 流式传输

使用 Go 构建 MCP Server：
推荐使用 mcp-go 库（github.com/mark3labs/mcp-go）。它提供了简洁的 API 来定义工具和资源。`, "mcp_protocol", 500)
}

func main() {
	assistant := NewRAGAssistant()
	ctx := context.Background()

	// 加载示例文档到知识库
	fmt.Println("📖 正在加载知识库...")
	loadSampleDocuments(assistant.kb)
	fmt.Printf("✅ 已加载 %d 个文档块\n", len(assistant.kb.Chunks))
	fmt.Println(strings.Repeat("=", 45))

	// 交互式问答
	fmt.Println("🤖 RAG 问答助手（输入 'quit' 退出）")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n❓ 你的问题: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}
		if input == "quit" {
			fmt.Println("👋 再见！")
			break
		}

		if err := assistant.Answer(ctx, input); err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
		}
	}
}
