#!/usr/bin/env python3
"""将所有章节合并为一个 PDF 学习手册"""

import os
import re
import markdown
from pygments import highlight
from pygments.lexers import get_lexer_by_name, TextLexer
from pygments.formatters import HtmlFormatter
from weasyprint import HTML

PROJECT_ROOT = os.path.dirname(os.path.abspath(__file__))

# 章节顺序
CHAPTERS = [
    "00-cli-quick-start",
    "01-api-basics",
    "02-prompt-engineering",
    "03-tool-use",
    "04-vision",
    "05-extended-thinking",
    "06-advanced-patterns",
    "07-mcp",
    "08-agent-patterns",
    "09-claude-code",
    "10-projects",
    "11-cli-mastery",
    "12-skills",
    "13-hooks-system",
]

# 代码文件扩展名
CODE_EXTENSIONS = {".go", ".sh", ".json"}
# Markdown 文档扩展名（非 README）
DOC_EXTENSIONS = {".md"}

CSS = """
@page {
    size: A4;
    margin: 2cm 1.5cm;
    @bottom-center {
        content: counter(page);
        font-size: 10px;
        color: #888;
    }
}

body {
    font-family: "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei",
                 "Noto Sans CJK SC", "Source Han Sans SC",
                 -apple-system, BlinkMacSystemFont, sans-serif;
    font-size: 12px;
    line-height: 1.7;
    color: #1a1a1a;
    max-width: 100%;
}

h1 {
    font-size: 24px;
    border-bottom: 2px solid #333;
    padding-bottom: 8px;
    margin-top: 40px;
    page-break-before: always;
}

h1:first-of-type {
    page-break-before: avoid;
}

h2 {
    font-size: 18px;
    color: #2c3e50;
    margin-top: 28px;
    border-bottom: 1px solid #eee;
    padding-bottom: 4px;
}

h3 {
    font-size: 15px;
    color: #34495e;
    margin-top: 20px;
}

pre {
    background: #f5f5f5;
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 12px;
    font-family: Menlo, Monaco, "Courier New", monospace,
                 "PingFang SC", "Hiragino Sans GB", "Noto Sans CJK SC";
    font-size: 10px;
    line-height: 1.5;
    overflow-wrap: break-word;
    white-space: pre-wrap;
    word-break: break-all;
}

code {
    font-family: Menlo, Monaco, "Courier New", monospace,
                 "PingFang SC", "Hiragino Sans GB", "Noto Sans CJK SC";
    font-size: 10.5px;
}

p code, li code, td code {
    background: #f0f0f0;
    padding: 1px 5px;
    border-radius: 3px;
    font-size: 11px;
}

table {
    border-collapse: collapse;
    width: 100%;
    margin: 12px 0;
    font-size: 11px;
}

th, td {
    border: 1px solid #ddd;
    padding: 6px 10px;
    text-align: left;
}

th {
    background: #f8f9fa;
    font-weight: 600;
}

tr:nth-child(even) {
    background: #fafafa;
}

blockquote {
    border-left: 4px solid #3498db;
    margin: 12px 0;
    padding: 8px 16px;
    background: #f0f7ff;
    color: #2c3e50;
}

.cover {
    text-align: center;
    padding: 120px 40px 60px;
    page-break-after: always;
}

.cover h1 {
    font-size: 36px;
    border: none;
    page-break-before: avoid;
    margin-bottom: 20px;
}

.cover p {
    font-size: 16px;
    color: #666;
}

.toc {
    page-break-after: always;
}

.toc h1 {
    page-break-before: avoid;
}

.toc ul {
    list-style: none;
    padding-left: 0;
}

.toc li {
    padding: 4px 0;
    border-bottom: 1px dotted #ddd;
}

.file-header {
    background: #2c3e50;
    color: white;
    padding: 6px 12px;
    border-radius: 4px 4px 0 0;
    font-size: 11px;
    font-family: monospace;
    margin-top: 20px;
    margin-bottom: 0;
}

.file-header + pre {
    margin-top: 0;
    border-top: none;
    border-radius: 0 0 4px 4px;
}

.chapter-divider {
    page-break-before: always;
}

""" + HtmlFormatter(style='friendly').get_style_defs('.highlight')


def collect_files(chapter_dir):
    """收集章节中的所有文件，按类型和名称排序"""
    readme = None
    code_files = []
    doc_files = []
    skill_files = []

    for root, dirs, files in os.walk(chapter_dir):
        # 排除隐藏目录
        dirs[:] = [d for d in dirs if not d.startswith('.')]

        for f in sorted(files):
            filepath = os.path.join(root, f)
            relpath = os.path.relpath(filepath, chapter_dir)
            _, ext = os.path.splitext(f)

            if f == "README.md":
                if root == chapter_dir:
                    readme = filepath
                else:
                    doc_files.append((relpath, filepath))
            elif f == "SKILL.md" or (ext == ".md" and "skills" in root):
                skill_files.append((relpath, filepath))
            elif ext in CODE_EXTENSIONS:
                code_files.append((relpath, filepath))
            elif ext in DOC_EXTENSIONS and f != "README.md":
                doc_files.append((relpath, filepath))

    return readme, code_files, doc_files, skill_files


def read_file(filepath):
    """读取文件内容"""
    with open(filepath, 'r', encoding='utf-8') as f:
        return f.read()


def get_lexer_for_file(filename):
    """根据文件扩展名获取语法高亮 lexer"""
    ext = os.path.splitext(filename)[1]
    mapping = {
        '.go': 'go',
        '.sh': 'bash',
        '.json': 'json',
        '.yaml': 'yaml',
        '.yml': 'yaml',
        '.md': 'markdown',
    }
    lang = mapping.get(ext, 'text')
    try:
        return get_lexer_by_name(lang)
    except:
        return TextLexer()


def format_code_block(content, filename):
    """将代码文件格式化为带高亮的 HTML"""
    lexer = get_lexer_for_file(filename)
    formatter = HtmlFormatter(nowrap=False, style='friendly')
    highlighted = highlight(content, lexer, formatter)
    return f'<div class="file-header">📄 {filename}</div>\n{highlighted}'


def format_skill_file(relpath, content):
    """格式化 SKILL.md 文件"""
    return f'<div class="file-header">🎯 {relpath}</div>\n<pre><code>{escape_html(content)}</code></pre>'


def escape_html(text):
    """转义 HTML 特殊字符"""
    return (text
            .replace('&', '&amp;')
            .replace('<', '&lt;')
            .replace('>', '&gt;')
            .replace('"', '&quot;'))


def markdown_to_html(md_text):
    """将 Markdown 转换为 HTML"""
    return markdown.markdown(
        md_text,
        extensions=['tables', 'fenced_code', 'codehilite', 'toc'],
        extension_configs={
            'codehilite': {
                'guess_lang': False,
                'css_class': 'highlight',
            }
        }
    )


def build_html():
    """构建完整的 HTML 文档"""
    parts = []

    # 封面
    parts.append("""
    <div class="cover">
        <h1>Claude 全面学习指南</h1>
        <p>使用 Go 语言系统学习 Claude API 及相关生态</p>
        <p>从入门到实战</p>
        <br><br>
        <p style="color: #999; font-size: 13px;">包含 13 个章节 · API 基础 · Prompt 工程 · Tool Use · MCP · Agent · Claude Code · Skills</p>
    </div>
    """)

    # 目录
    toc_items = []
    for chapter in CHAPTERS:
        chapter_dir = os.path.join(PROJECT_ROOT, chapter)
        if not os.path.isdir(chapter_dir):
            continue
        readme_path = os.path.join(chapter_dir, "README.md")
        if os.path.exists(readme_path):
            content = read_file(readme_path)
            # 提取第一个 h1 标题
            match = re.search(r'^#\s+(.+)$', content, re.MULTILINE)
            title = match.group(1) if match else chapter
        else:
            title = chapter
        toc_items.append(f'<li><strong>{chapter}</strong> — {title}</li>')

    parts.append(f"""
    <div class="toc">
        <h1>目录</h1>
        <ul>{"".join(toc_items)}</ul>
    </div>
    """)

    # 先添加根 README
    root_readme = os.path.join(PROJECT_ROOT, "README.md")
    if os.path.exists(root_readme):
        parts.append('<div class="chapter-divider"></div>')
        parts.append(markdown_to_html(read_file(root_readme)))

    # 各章节
    for chapter in CHAPTERS:
        chapter_dir = os.path.join(PROJECT_ROOT, chapter)
        if not os.path.isdir(chapter_dir):
            print(f"  跳过: {chapter}（目录不存在）")
            continue

        print(f"  处理: {chapter}")
        readme, code_files, doc_files, skill_files = collect_files(chapter_dir)

        parts.append('<div class="chapter-divider"></div>')

        # README
        if readme:
            parts.append(markdown_to_html(read_file(readme)))

        # 文档文件
        for relpath, filepath in doc_files:
            parts.append(f'<h3>📝 {relpath}</h3>')
            parts.append(markdown_to_html(read_file(filepath)))

        # 代码文件
        for relpath, filepath in code_files:
            content = read_file(filepath)
            parts.append(format_code_block(content, relpath))

        # Skill 文件
        for relpath, filepath in skill_files:
            content = read_file(filepath)
            parts.append(format_skill_file(relpath, content))

    # 组装完整 HTML
    html = f"""<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <title>Claude 全面学习指南 (Golang)</title>
    <style>{CSS}</style>
</head>
<body>
{"".join(parts)}
</body>
</html>"""

    return html


def main():
    print("正在生成 PDF 学习手册...")
    print()

    # 生成 HTML
    html = build_html()

    # 保存 HTML（方便调试）
    html_path = os.path.join(PROJECT_ROOT, "claude-learning-guide.html")
    with open(html_path, 'w', encoding='utf-8') as f:
        f.write(html)
    print(f"\nHTML 已保存: {html_path}")

    # 生成 PDF
    pdf_path = os.path.join(PROJECT_ROOT, "claude-learning-guide.pdf")
    print(f"正在生成 PDF: {pdf_path}")
    HTML(string=html).write_pdf(pdf_path)

    # 文件大小
    size_mb = os.path.getsize(pdf_path) / (1024 * 1024)
    print(f"\n✅ PDF 生成完成: {pdf_path} ({size_mb:.1f} MB)")


if __name__ == "__main__":
    main()
