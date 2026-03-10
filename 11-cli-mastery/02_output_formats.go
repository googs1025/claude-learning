//go:build ignore

// 第十一章示例 2：输出格式解析
// 运行方式: go run 02_output_formats.go
//
// 本示例演示如何：
// 1. 使用 text 格式获取纯文本输出
// 2. 使用 json 格式获取结构化响应（含 cost、usage）
// 3. 使用 stream-json 格式实时处理流式 JSON
//
// 三种格式对比：
// - text:        纯文本，最简单，适合直接展示
// - json:        完整 JSON，包含 cost/usage/session_id，适合自动化
// - stream-json: 逐行 JSON 事件流，适合实时处理

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// FullCLIResponse 表示 --output-format json 的完整响应
type FullCLIResponse struct {
	Type         string  `json:"type"`           // "result"
	Subtype      string  `json:"subtype"`        // "success" 或 "error"
	Result       string  `json:"result"`         // Claude 的文本回复
	SessionID    string  `json:"session_id"`     // 会话 UUID
	TotalCostUSD float64 `json:"total_cost_usd"` // 本次调用总费用（美元）
	Usage        struct {
		InputTokens       int `json:"input_tokens"`        // 输入 token 数
		OutputTokens      int `json:"output_tokens"`       // 输出 token 数
		CacheReadTokens   int `json:"cache_read_tokens"`   // 缓存读取 token 数
		CacheCreatedTokens int `json:"cache_created_tokens"` // 缓存创建 token 数
	} `json:"usage"`
}

// StreamEvent 表示 stream-json 中的单个事件
type StreamEvent struct {
	Type    string `json:"type"`    // "assistant", "result" 等
	Message string `json:"message"` // 文本内容（assistant 类型时）
	Result  string `json:"result"`  // 最终结果（result 类型时）
}

func main() {
	prompt := "用一句话解释什么是 goroutine"

	// ========== 格式 1：text（默认） ==========
	fmt.Println("=== 格式 1: text（默认） ===")
	demoTextFormat(prompt)

	// ========== 格式 2：json ==========
	fmt.Println("\n=== 格式 2: json ===")
	demoJSONFormat(prompt)

	// ========== 格式 3：stream-json ==========
	fmt.Println("\n=== 格式 3: stream-json ===")
	demoStreamJSONFormat(prompt)
}

// demoTextFormat 演示纯文本输出
func demoTextFormat(prompt string) {
	cmd := exec.Command("claude",
		"-p", prompt,
		"--output-format", "text",
		"--model", "sonnet",
	)

	// text 格式可以用 CombinedOutput，输出就是纯文本
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("text 格式调用失败: %v", err)
	}

	fmt.Printf("纯文本回复: %s\n", strings.TrimSpace(string(output)))
	fmt.Println("特点: 最简单，但没有 cost/usage 等元信息")
}

// demoJSONFormat 演示 JSON 格式输出并提取费用信息
func demoJSONFormat(prompt string) {
	cmd := exec.Command("claude",
		"-p", prompt,
		"--output-format", "json",
		"--model", "sonnet",
	)

	// JSON 格式必须用 Output()，避免 stderr 混入
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("json 格式调用失败: %v", err)
	}

	var resp FullCLIResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		log.Fatalf("解析 JSON 失败: %v\n原始: %s", err, string(output))
	}

	fmt.Printf("回复: %s\n", resp.Result)
	fmt.Printf("费用: $%.6f\n", resp.TotalCostUSD)
	fmt.Printf("Token 使用: 输入=%d, 输出=%d\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)
	fmt.Printf("会话 ID: %s\n", resp.SessionID)
	fmt.Println("特点: 包含完整元信息，适合自动化脚本")
}

// demoStreamJSONFormat 演示流式 JSON 格式
func demoStreamJSONFormat(prompt string) {
	cmd := exec.Command("claude",
		"-p", prompt,
		"--output-format", "stream-json",
		"--model", "sonnet",
	)

	// stream-json 需要逐行读取 stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("创建 stdout pipe 失败: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("启动命令失败: %v", err)
	}

	// 逐行扫描 JSON 事件
	scanner := bufio.NewScanner(stdout)
	eventCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		eventCount++

		// 解析每一行为 JSON 事件
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // 跳过非 JSON 行
		}

		eventType, _ := event["type"].(string)
		fmt.Printf("  事件 %d: type=%s\n", eventCount, eventType)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("命令执行失败: %v", err)
	}

	fmt.Printf("共收到 %d 个事件\n", eventCount)
	fmt.Println("特点: 实时获取中间事件，适合进度展示和流式 UI")
}
