//go:build ignore

// 第零章示例 5：结构化 JSON 输出
// 运行方式: go run 05_structured_output.go
//
// 本示例演示如何：
// 1. 使用 --output-format json 获取 JSON 格式的完整响应
// 2. 使用 --json-schema 约束输出的 JSON 结构
// 3. 将 Claude 的输出解析为 Go 结构体
//
// --json-schema 确保 Claude 的回复严格遵循指定的 JSON 格式

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
	cmd := exec.Command("claude",
		"-p", prompt,
		"--output-format", "json",
		"--json-schema", schema,
		"--model", "sonnet",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("调用 claude CLI 失败: %v\n输出: %s", err, string(output))
	}

	// 解析 JSON 输出
	// --output-format json 会返回一个包含 result 字段的 JSON 对象
	var response struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		// 如果整体不是 JSON 包装，尝试直接解析为 BookList
		var books BookList
		if err2 := json.Unmarshal(output, &books); err2 != nil {
			log.Fatalf("解析 JSON 失败: %v\n原始输出: %s", err, string(output))
		}
		printBooks(books)
		return
	}

	// 解析 result 字段中的实际内容
	var books BookList
	if err := json.Unmarshal([]byte(response.Result), &books); err != nil {
		log.Fatalf("解析书籍数据失败: %v\n原始数据: %s", err, response.Result)
	}

	printBooks(books)
}

func printBooks(books BookList) {
	fmt.Printf("=== %s ===\n\n", books.Topic)
	for i, book := range books.Books {
		fmt.Printf("%d. 《%s》- %s\n", i+1, book.Title, book.Author)
		fmt.Printf("   推荐理由: %s\n\n", book.Reason)
	}
}
