//go:build ignore

// 第零章示例 5：结构化 JSON 输出
// 运行方式: go run 05_structured_output.go
//
// 本示例演示如何：
// 1. 使用 --output-format json 获取 JSON 格式的完整响应
// 2. 使用 --json-schema 约束输出的 JSON 结构
// 3. 将 Claude 的输出解析为 Go 结构体
//
// --output-format json 返回格式：
// {"type":"result","result":"...","structured_output":{...},"session_id":"...","total_cost_usd":...}
// 使用 --json-schema 时，structured_output 字段包含符合 schema 的已解析 JSON 对象

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

// BookRecommendation 表示一本推荐书籍
type BookRecommendation struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Reason string `json:"reason"`
}

// BookList 表示推荐书籍列表
type BookList struct {
	Topic string               `json:"topic"`
	Books []BookRecommendation `json:"books"`
}

func main() {
	prompt := "推荐 3 本学习 Go 语言的书籍"

	// 定义期望的 JSON Schema
	// Claude 会严格按照这个结构输出 JSON
	schema := `{
		"type": "object",
		"required": ["topic", "books"],
		"properties": {
			"topic": {
				"type": "string",
				"description": "推荐主题"
			},
			"books": {
				"type": "array",
				"items": {
					"type": "object",
					"required": ["title", "author", "reason"],
					"properties": {
						"title": {"type": "string", "description": "书名"},
						"author": {"type": "string", "description": "作者"},
						"reason": {"type": "string", "description": "推荐理由（一句话）"}
					}
				}
			}
		}
	}`

	// 调用 claude CLI，使用 --json-schema 参数
	// 注意：使用 Output() 而不是 CombinedOutput()，避免 stderr 混入 JSON
	cmd := exec.Command("claude",
		"-p", prompt,
		"--output-format", "json",
		"--json-schema", schema,
		"--model", "sonnet",
	)

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("调用 claude CLI 失败: %v", err)
	}

	// 解析 CLI 的 JSON 响应
	// --output-format json 返回：{"type":"result", "result":"...", "structured_output":{...}, ...}
	// 使用 --json-schema 时，structured_output 包含符合 schema 的 JSON 对象
	var response struct {
		Type             string   `json:"type"`
		Result           string   `json:"result"`
		StructuredOutput BookList `json:"structured_output"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		log.Fatalf("解析 JSON 响应失败: %v\n原始输出: %s", err, string(output))
	}

	printBooks(response.StructuredOutput)
}

func printBooks(books BookList) {
	fmt.Printf("=== %s ===\n\n", books.Topic)
	for i, book := range books.Books {
		fmt.Printf("%d. 《%s》- %s\n", i+1, book.Title, book.Author)
		fmt.Printf("   推荐理由: %s\n\n", book.Reason)
	}
}
