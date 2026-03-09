# 第六章：高级模式

本章介绍 Claude API 的高级使用模式，帮助你构建更高效、更稳定的生产级应用。

## 前置条件

1. 完成前五章的学习
2. 设置环境变量 `ANTHROPIC_API_KEY`
3. 安装依赖：
   ```bash
   go get github.com/anthropics/anthropic-sdk-go
   ```

## 课程内容

| 文件 | 主题 | 说明 |
|------|------|------|
| `01_prompt_caching.go` | Prompt 缓存 | 缓存大型系统提示词，减少重复计算和费用 |
| `02_batch_processing.go` | 批量并发处理 | 使用 goroutine 并发处理多个请求 |
| `03_retry_strategy.go` | 重试策略 | 指数退避 + 抖动，处理速率限制和瞬时错误 |
| `04_token_counting.go` | Token 计数与成本估算 | 跟踪 token 使用量，计算 API 调用费用 |

## 运行方式

```bash
export ANTHROPIC_API_KEY="your-api-key"

go run 01_prompt_caching.go
go run 02_batch_processing.go
go run 03_retry_strategy.go
go run 04_token_counting.go
```

## 核心概念

### Prompt 缓存
当你有一个很长的系统提示词（如详细的规则文档），每次请求都会重新处理这些内容。Prompt 缓存允许你标记这些内容为可缓存的，后续请求可以复用已处理的结果，显著减少延迟和成本。

### 并发处理
Go 语言的 goroutine 天然适合并发调用 API。使用 `sync.WaitGroup` 和 channel 可以优雅地并发处理多个请求，大幅提高吞吐量。

### 重试策略
生产环境中必须处理 API 的速率限制（429）和服务端过载（529）。指数退避加随机抖动是业界标准的重试策略，避免"惊群效应"。

### 成本控制
了解 token 消耗情况是控制 API 成本的基础。通过跟踪每次调用的输入/输出 token 数量，你可以准确估算费用并优化提示词长度。
