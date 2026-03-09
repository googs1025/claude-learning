# 第四章：视觉能力（Vision）

本章介绍如何使用 Claude 的视觉能力，让 Claude 理解和分析图片内容。

## 前置条件

1. 安装 Go 1.21+
2. 设置环境变量 `ANTHROPIC_API_KEY`
3. 安装依赖：
   ```bash
   go get github.com/anthropics/anthropic-sdk-go
   ```

## 课程内容

| 文件 | 主题 | 说明 |
|------|------|------|
| `01_image_url.go` | URL 图片分析 | 通过公开 URL 发送图片，让 Claude 描述图片内容 |
| `02_image_base64.go` | Base64 图片分析 | 将本地图片编码为 Base64 后发送给 Claude |
| `03_multi_images.go` | 多图对比 | 在一条消息中发送多张图片，让 Claude 进行对比分析 |
| `04_image_with_tools.go` | 视觉 + 工具 | 结合视觉能力和工具调用，实现图片自动分类 |

## 运行方式

```bash
# 确保设置了 API 密钥
export ANTHROPIC_API_KEY="your-api-key"

# 运行任意示例
go run 01_image_url.go
go run 02_image_base64.go
go run 03_multi_images.go
go run 04_image_with_tools.go
```

## 核心概念

### 图片输入方式
Claude 支持两种图片输入方式：
- **URL 方式**：直接传入可公开访问的图片 URL，API 会自动下载图片
- **Base64 方式**：将图片编码为 Base64 字符串后传入，适合本地文件或私有图片

### 支持的图片格式
- JPEG (`image/jpeg`)
- PNG (`image/png`)
- GIF (`image/gif`)
- WebP (`image/webp`)

### 多图分析
一条消息中可以包含多个图片内容块（ImageBlock），Claude 能够同时理解和对比多张图片。

### 视觉 + 工具
Claude 的视觉能力可以和工具调用结合使用，实现更结构化的图片分析结果。例如，让 Claude 调用分类工具将图片归入预定义类别。
