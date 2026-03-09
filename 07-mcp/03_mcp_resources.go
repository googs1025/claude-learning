// 第七章示例 3：MCP 资源（Resources）
// 运行方式: go run 03_mcp_resources.go
//
// 本示例演示如何：
// 1. 使用 MCP Resources 暴露数据给 AI 模型
// 2. 定义静态资源（固定 URI）和动态资源（URI 模板）
// 3. 暴露配置信息、系统状态等数据
// 4. AI 模型可以通过 URI 读取这些资源
//
// Resources 与 Tools 的区别：
//   - Tools: AI 主动调用，执行操作（有副作用）
//   - Resources: AI 主动读取，获取数据（只读，无副作用）
//
// 测试方法：
//   go build -o resource-server 03_mcp_resources.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// AppConfig 模拟应用配置数据
type AppConfig struct {
	AppName    string
	Version    string
	Debug      bool
	MaxRetries int
	APIBaseURL string
	Features   map[string]bool
}

// 全局配置（模拟真实应用的配置）
var appConfig = AppConfig{
	AppName:    "MyApp",
	Version:    "2.1.0",
	Debug:      false,
	MaxRetries: 3,
	APIBaseURL: "https://api.example.com",
	Features: map[string]bool{
		"dark_mode":     true,
		"notifications": true,
		"beta_features": false,
		"analytics":     true,
	},
}

func main() {
	// 创建 MCP Server
	s := server.NewMCPServer(
		"resource-demo-server",
		"1.0.0",
	)

	// ==================== 资源 1：应用配置 ====================
	// 静态资源：固定的 URI，返回应用配置信息
	s.AddResource(
		mcp.Resource{
			URI:         "config://app/settings",
			Name:        "应用配置",
			Description: "当前应用的所有配置信息",
			MIMEType:    "application/json",
		},
		appConfigHandler,
	)

	// ==================== 资源 2：系统状态 ====================
	// 静态资源：返回实时系统状态
	s.AddResource(
		mcp.Resource{
			URI:         "status://system/info",
			Name:        "系统状态",
			Description: "当前系统的运行状态信息（Go 运行时、内存使用等）",
			MIMEType:    "text/plain",
		},
		systemStatusHandler,
	)

	// ==================== 资源 3：功能开关 ====================
	// 静态资源：返回功能开关列表
	s.AddResource(
		mcp.Resource{
			URI:         "config://app/features",
			Name:        "功能开关",
			Description: "所有功能开关的启用/禁用状态",
			MIMEType:    "text/plain",
		},
		featureFlagsHandler,
	)

	// ==================== 资源 4：帮助文档 ====================
	// 静态资源：返回帮助信息
	s.AddResource(
		mcp.Resource{
			URI:         "docs://help/getting-started",
			Name:        "快速入门指南",
			Description: "应用的快速入门使用指南",
			MIMEType:    "text/markdown",
		},
		helpDocHandler,
	)

	// ==================== 资源 5：环境变量 ====================
	// 静态资源：返回安全的环境变量信息
	s.AddResource(
		mcp.Resource{
			URI:         "env://variables",
			Name:        "环境变量",
			Description: "当前进程的部分环境变量（已过滤敏感信息）",
			MIMEType:    "text/plain",
		},
		envVarsHandler,
	)

	// 启动 Stdio Server
	fmt.Fprintln(os.Stderr, "Resource Demo MCP Server 已启动")
	fmt.Fprintln(os.Stderr, "已注册资源:")
	fmt.Fprintln(os.Stderr, "  - config://app/settings (应用配置)")
	fmt.Fprintln(os.Stderr, "  - status://system/info (系统状态)")
	fmt.Fprintln(os.Stderr, "  - config://app/features (功能开关)")
	fmt.Fprintln(os.Stderr, "  - docs://help/getting-started (帮助文档)")
	fmt.Fprintln(os.Stderr, "  - env://variables (环境变量)")

	stdioServer := server.NewStdioServer(s)
	ctx := context.Background()
	if err := stdioServer.Listen(ctx, os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Server 错误: %v", err)
	}
}

// appConfigHandler 返回应用配置信息
func appConfigHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// 将配置格式化为 JSON 风格的文本
	config := fmt.Sprintf(`{
  "app_name": "%s",
  "version": "%s",
  "debug": %v,
  "max_retries": %d,
  "api_base_url": "%s",
  "features": {
    "dark_mode": %v,
    "notifications": %v,
    "beta_features": %v,
    "analytics": %v
  }
}`,
		appConfig.AppName,
		appConfig.Version,
		appConfig.Debug,
		appConfig.MaxRetries,
		appConfig.APIBaseURL,
		appConfig.Features["dark_mode"],
		appConfig.Features["notifications"],
		appConfig.Features["beta_features"],
		appConfig.Features["analytics"],
	)

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "config://app/settings",
			MIMEType: "application/json",
			Text:     config,
		},
	}, nil
}

// systemStatusHandler 返回系统运行状态
func systemStatusHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status := fmt.Sprintf(
		"系统状态报告\n"+
			"生成时间: %s\n"+
			"====================\n"+
			"Go 版本: %s\n"+
			"操作系统: %s\n"+
			"架构: %s\n"+
			"CPU 核数: %d\n"+
			"Goroutine 数量: %d\n"+
			"====================\n"+
			"内存使用:\n"+
			"  已分配: %s\n"+
			"  总分配: %s\n"+
			"  系统内存: %s\n"+
			"  GC 次数: %d\n",
		time.Now().Format("2006-01-02 15:04:05"),
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		runtime.NumCPU(),
		runtime.NumGoroutine(),
		formatBytes(m.Alloc),
		formatBytes(m.TotalAlloc),
		formatBytes(m.Sys),
		m.NumGC,
	)

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "status://system/info",
			MIMEType: "text/plain",
			Text:     status,
		},
	}, nil
}

// featureFlagsHandler 返回功能开关列表
func featureFlagsHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	var sb strings.Builder
	sb.WriteString("功能开关列表\n")
	sb.WriteString("====================\n")

	for name, enabled := range appConfig.Features {
		status := "关闭"
		if enabled {
			status = "开启"
		}
		sb.WriteString(fmt.Sprintf("  %-20s: %s\n", name, status))
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "config://app/features",
			MIMEType: "text/plain",
			Text:     sb.String(),
		},
	}, nil
}

// helpDocHandler 返回帮助文档
func helpDocHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	doc := `# 快速入门指南

## 安装

1. 下载最新版本
2. 配置环境变量
3. 运行初始化命令

## 基本使用

### 启动应用
` + "```bash" + `
./myapp start
` + "```" + `

### 查看状态
` + "```bash" + `
./myapp status
` + "```" + `

### 配置管理
- 配置文件位于 ` + "`~/.myapp/config.yaml`" + `
- 使用 ` + "`./myapp config set key value`" + ` 修改配置
- 使用 ` + "`./myapp config get key`" + ` 查看配置

## 常见问题

**Q: 如何重置配置？**
A: 删除配置文件后重启应用即可自动生成默认配置。

**Q: 如何查看日志？**
A: 日志文件位于 ` + "`~/.myapp/logs/`" + ` 目录下。
`

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "docs://help/getting-started",
			MIMEType: "text/markdown",
			Text:     doc,
		},
	}, nil
}

// envVarsHandler 返回过滤后的环境变量
// 安全起见，过滤掉包含 KEY、SECRET、TOKEN、PASSWORD 的变量
func envVarsHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	var sb strings.Builder
	sb.WriteString("环境变量（已过滤敏感信息）\n")
	sb.WriteString("====================\n")

	// 定义敏感关键词列表
	sensitiveKeywords := []string{"KEY", "SECRET", "TOKEN", "PASSWORD", "CREDENTIAL", "AUTH"}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name := parts[0]

		// 检查是否包含敏感关键词
		isSensitive := false
		upperName := strings.ToUpper(name)
		for _, keyword := range sensitiveKeywords {
			if strings.Contains(upperName, keyword) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sb.WriteString(fmt.Sprintf("  %s = [已隐藏]\n", name))
		} else {
			value := parts[1]
			// 截断过长的值
			if len(value) > 100 {
				value = value[:100] + "..."
			}
			sb.WriteString(fmt.Sprintf("  %s = %s\n", name, value))
		}
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "env://variables",
			MIMEType: "text/plain",
			Text:     sb.String(),
		},
	}, nil
}

// formatBytes 将字节数格式化为人类可读的字符串
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
