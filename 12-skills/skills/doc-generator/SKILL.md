---
name: doc-generator
description: 根据模板为 Go 包生成文档
argument-hint: "[package-path]"
user-invocable: true
allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
---

# Go 包文档生成器

为 `$ARGUMENTS` 指定的 Go 包生成标准化文档。

## 步骤

1. **分析包结构**: 读取目标包中的所有 `.go` 文件
2. **提取公开 API**: 列出所有导出的类型、函数和方法
3. **应用模板**: 使用 [template.md](template.md) 中的模板格式
4. **生成文档**: 将文档写入目标包的 `DOC.md` 文件

## 分析要点

- 包的整体用途和设计意图
- 核心类型及其关系
- 主要函数的输入输出
- 使用示例（从测试文件或注释中提取）
- 依赖关系

## 注意事项

- 只为导出的（大写开头的）符号生成文档
- 如果有 `_test.go` 文件，从中提取使用示例
- 如果有 `doc.go` 文件，优先使用其中的包描述
- 保持文档简洁，避免重复 godoc 已有的内容
