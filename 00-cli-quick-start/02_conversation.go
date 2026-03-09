//go:build ignore

// 第零章示例 2：多轮对话
// 运行方式: go run 02_conversation.go
//
// 本示例演示如何：
// 1. 发送第一条消息开始对话
// 2. 使用 --continue 参数继续同一对话
// 3. Claude 会记住之前的对话上下文
//
// 注意：--continue 会继续最近一次的对话会话

package main

import (
	"fmt"
	"log"
	"os/exec"
)

// callClaude 调用 claude CLI 并返回输出
func callClaude(args ...string) string {
	cmd := exec.Command("claude", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("调用 claude CLI 失败: %v\n输出: %s", err, string(output))
	}
	return string(output)
}

func main() {
	// 第一轮对话：提出一个话题
	fmt.Println("=== 第一轮对话 ===")
	fmt.Println("用户: 请记住这个数字：42。然后告诉我它在数学中有什么特别之处？")
	reply1 := callClaude("-p", "请记住这个数字：42。然后告诉我它在数学中有什么特别之处？", "--model", "sonnet")
	fmt.Println("Claude:", reply1)

	// 第二轮对话：使用 --continue 继续上一次对话
	// Claude 应该记得我们之前提到的数字 42
	fmt.Println("=== 第二轮对话（继续） ===")
	fmt.Println("用户: 我之前让你记住的数字是什么？")
	reply2 := callClaude("-p", "我之前让你记住的数字是什么？", "--continue", "--model", "sonnet")
	fmt.Println("Claude:", reply2)

	// 第三轮对话：继续追问
	fmt.Println("=== 第三轮对话（继续） ===")
	fmt.Println("用户: 把那个数字乘以 2 是多少？")
	reply3 := callClaude("-p", "把那个数字乘以 2 是多少？", "--continue", "--model", "sonnet")
	fmt.Println("Claude:", reply3)
}
