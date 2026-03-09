# 第10章：综合实战项目

## 概述

本章包含三个完整的实战项目，综合运用前面章节学到的所有知识。

## 项目列表

### 1. CLI 聊天机器人 (`chatbot/`)
一个功能完整的终端聊天机器人，支持：
- 多轮对话与上下文记忆
- 流式输出（实时显示响应）
- System Prompt 自定义
- 对话历史管理

**涉及知识点**: 第1章 API 基础、第2章 Prompt 工程

```bash
cd chatbot && go run main.go
```

### 2. 代码审查助手 (`code-reviewer/`)
一个自动代码审查工具，支持：
- 读取 Go 源代码文件
- 从多个维度审查代码（安全性、性能、可读性）
- 使用 Tool Use 进行结构化输出
- 生成审查报告

**涉及知识点**: 第3章 Tool Use、第4章 结构化输出

```bash
cd code-reviewer && go run main.go <文件路径>
```

### 3. RAG 问答助手 (`rag-assistant/`)
基于检索增强生成（RAG）的问答系统，支持：
- 加载本地文本文档
- 简单的文本分块与相似度搜索
- 基于检索结果回答问题
- 引用来源标注

**涉及知识点**: 第2章 Prompt 工程、第3章 Tool Use、第8章 Agent 模式

```bash
cd rag-assistant && go run main.go
```
