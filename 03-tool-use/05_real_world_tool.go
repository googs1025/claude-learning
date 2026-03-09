// 第三章示例 5：真实 API 调用 —— HTTP 请求工具
// 运行方式: go run 05_real_world_tool.go
//
// 本示例演示如何：
// 1. 构建一个调用真实 HTTP API 的工具
// 2. 使用 httpbin.org 作为测试 API（公开、无需认证）
// 3. 让 Claude 通过工具获取真实的网络数据
// 4. 完整展示从工具定义到执行到回复的全流程
//
// httpbin.org 是一个免费的 HTTP 测试服务，支持多种端点：
// - /ip: 返回请求者的 IP 地址
// - /user-agent: 返回请求的 User-Agent
// - /headers: 返回请求头信息
// - /get?key=value: 回显 GET 请求参数
// - /uuid: 生成一个随机 UUID

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// ========================================
// HTTP 工具的执行函数
// ========================================

// executeHTTPGet 执行 HTTP GET 请求
// 这是一个真实的网络调用，会访问 httpbin.org
func executeHTTPGet(input json.RawMessage) (string, bool) {
	var params struct {
		Endpoint    string            `json:"endpoint"`
		QueryParams map[string]string `json:"query_params"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return fmt.Sprintf("参数解析失败: %v", err), true
	}

	// 构造完整的 URL
	baseURL := "https://httpbin.org"
	fullURL := baseURL + params.Endpoint

	// 添加查询参数
	if len(params.QueryParams) > 0 {
		queryValues := url.Values{}
		for k, v := range params.QueryParams {
			queryValues.Set(k, v)
		}
		fullURL += "?" + queryValues.Encode()
	}

	fmt.Printf("  [HTTP] 请求 URL: %s\n", fullURL)

	// 创建带超时的 HTTP 客户端
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送 GET 请求
	resp, err := httpClient.Get(fullURL)
	if err != nil {
		return fmt.Sprintf("HTTP 请求失败: %v", err), true
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("读取响应失败: %v", err), true
	}

	fmt.Printf("  [HTTP] 响应状态: %d\n", resp.StatusCode)

	// 检查状态码
	if resp.StatusCode != 200 {
		return fmt.Sprintf("HTTP 错误，状态码: %d，响应: %s", resp.StatusCode, string(body)), true
	}

	return string(body), false
}

// executeGetPublicIP 获取公网 IP（简化版工具，无参数）
func executeGetPublicIP() (string, bool) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	resp, err := httpClient.Get("https://httpbin.org/ip")
	if err != nil {
		return fmt.Sprintf("获取 IP 失败: %v", err), true
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("读取响应失败: %v", err), true
	}

	return string(body), false
}

// executeGenerateUUID 生成随机 UUID
func executeGenerateUUID() (string, bool) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	resp, err := httpClient.Get("https://httpbin.org/uuid")
	if err != nil {
		return fmt.Sprintf("生成 UUID 失败: %v", err), true
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("读取响应失败: %v", err), true
	}

	return string(body), false
}

// executeTool 工具分发器
func executeTool(name string, input json.RawMessage) (string, bool) {
	switch name {
	case "http_get":
		return executeHTTPGet(input)
	case "get_public_ip":
		return executeGetPublicIP()
	case "generate_uuid":
		return executeGenerateUUID()
	default:
		return fmt.Sprintf("未知工具: %s", name), true
	}
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// ========================================
	// 定义真实 API 工具
	// ========================================
	tools := []anthropic.ToolUnionParam{
		// 工具 1：通用 HTTP GET 请求
		// 这是一个功能强大的工具，可以向 httpbin.org 的各种端点发送请求
		{
			OfTool: &anthropic.ToolParam{
				Name: "http_get",
				Description: anthropic.String(
					"向 httpbin.org 发送 HTTP GET 请求。" +
						"可用端点包括：/headers（查看请求头）、/get（回显参数）、" +
						"/user-agent（查看UA）、/ip（查看IP）、/uuid（生成UUID）。",
				),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"endpoint": map[string]any{
							"type":        "string",
							"description": "API 端点路径，如 /headers, /get, /user-agent",
						},
						"query_params": map[string]any{
							"type":        "object",
							"description": "可选的查询参数，键值对形式",
							"additionalProperties": map[string]any{
								"type": "string",
							},
						},
					},
					Required: []string{"endpoint"},
				},
			},
		},
		// 工具 2：获取公网 IP（无参数的简单工具）
		{
			OfTool: &anthropic.ToolParam{
				Name:        "get_public_ip",
				Description: anthropic.String("获取当前网络的公网 IP 地址。这个工具不需要任何参数。"),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{},
				},
			},
		},
		// 工具 3：生成 UUID
		{
			OfTool: &anthropic.ToolParam{
				Name:        "generate_uuid",
				Description: anthropic.String("生成一个随机的 UUID（通用唯一标识符）。不需要参数。"),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{},
				},
			},
		},
	}

	// ========================================
	// 发送请求，让 Claude 使用真实 API
	// ========================================
	userQuestion := "请帮我做以下几件事：1. 查看我当前的公网IP地址；2. 生成一个随机UUID；3. 用 httpbin 的 /get 端点测试一下，发送参数 name=Claude 和 language=Go。"
	fmt.Printf("用户: %s\n\n", userQuestion)

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userQuestion)),
	}

	// ========================================
	// Agentic Loop：处理可能的多轮工具调用
	// ========================================
	maxRounds := 10
	round := 0

	for round < maxRounds {
		round++
		fmt.Printf("\n========== 第 %d 轮 ==========\n", round)

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

		// 任务完成
		if message.StopReason == "end_turn" {
			fmt.Println("\n========== Claude 的最终回复 ==========")
			for _, block := range message.Content {
				if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
					fmt.Println(tb.Text)
				}
			}
			break
		}

		// 需要执行工具
		messages = append(messages, message.ToParam())

		var toolResults []anthropic.ContentBlockParamUnion

		for _, block := range message.Content {
			switch v := block.AsAny().(type) {
			case anthropic.TextBlock:
				fmt.Printf("[Claude] %s\n", v.Text)

			case anthropic.ToolUseBlock:
				rawInput := json.RawMessage(v.Input)
				fmt.Printf("\n[调用工具] %s\n", v.Name)
				fmt.Printf("  参数: %s\n", string(rawInput))

				// 执行真实的 HTTP 工具调用
				result, isError := executeTool(v.Name, rawInput)

				if isError {
					fmt.Printf("  错误: %s\n", result)
				} else {
					// 美化 JSON 输出（如果是 JSON）
					var prettyJSON json.RawMessage
					if err := json.Unmarshal([]byte(result), &prettyJSON); err == nil {
						formatted, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
						fmt.Printf("  结果:\n  %s\n", string(formatted))
					} else {
						fmt.Printf("  结果: %s\n", result)
					}
				}

				toolResults = append(toolResults,
					anthropic.NewToolResultBlock(v.ID, result, isError),
				)
			}
		}

		if len(toolResults) > 0 {
			messages = append(messages,
				anthropic.NewUserMessage(toolResults...),
			)
		}
	}

	if round >= maxRounds {
		fmt.Printf("\n警告: 达到最大轮数 %d\n", maxRounds)
	}

	// ========================================
	// 总结说明
	// ========================================
	fmt.Println("\n========== 关键要点 ==========")
	fmt.Println("1. 真实工具调用涉及网络请求，要设置合理的超时时间")
	fmt.Println("2. 始终处理 HTTP 错误和异常情况")
	fmt.Println("3. 工具的错误应通过 isError=true 返回，让 Claude 知道调用失败")
	fmt.Println("4. Claude 能理解 JSON 格式的 API 响应并提取关键信息")
	fmt.Println("5. 在生产环境中，还需要考虑速率限制、认证、日志等")
}
