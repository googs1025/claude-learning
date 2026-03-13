#!/usr/bin/env python3
"""抓取 code.claude.com 中文文档（.md 格式）并生成 PDF"""

import os
import re
import subprocess
import markdown

PROJECT_ROOT = os.path.dirname(os.path.abspath(__file__))

# 所有文档页面（从 llms.txt 提取的 61 个页面）
PAGES = [
    ("overview", "概述"),
    ("quickstart", "快速开始"),
    ("how-claude-code-works", "Claude Code 工作原理"),
    ("features-overview", "功能概览"),
    ("setup", "安装设置"),
    ("authentication", "认证"),
    ("interactive-mode", "交互模式"),
    ("cli-reference", "CLI 参考"),
    ("common-workflows", "常见工作流"),
    ("best-practices", "最佳实践"),
    ("memory", "记忆与 CLAUDE.md"),
    ("skills", "Skills 技能"),
    ("hooks", "Hooks"),
    ("hooks-guide", "Hooks 指南"),
    ("mcp", "MCP 协议"),
    ("sub-agents", "子代理"),
    ("agent-teams", "Agent 团队"),
    ("permissions", "权限"),
    ("security", "安全"),
    ("sandboxing", "沙箱"),
    ("settings", "设置"),
    ("model-config", "模型配置"),
    ("costs", "成本"),
    ("monitoring-usage", "监控使用"),
    ("analytics", "分析"),
    ("output-styles", "输出样式"),
    ("fast-mode", "快速模式"),
    ("keybindings", "快捷键"),
    ("terminal-config", "终端配置"),
    ("statusline", "状态栏"),
    ("checkpointing", "检查点"),
    ("vs-code", "VS Code"),
    ("jetbrains", "JetBrains"),
    ("desktop", "桌面应用"),
    ("desktop-quickstart", "桌面快速开始"),
    ("chrome", "Chrome 浏览器"),
    ("claude-code-on-the-web", "Web 版"),
    ("remote-control", "远程控制"),
    ("scheduled-tasks", "定时任务"),
    ("github-actions", "GitHub Actions"),
    ("gitlab-ci-cd", "GitLab CI/CD"),
    ("slack", "Slack 集成"),
    ("headless", "无头模式"),
    ("code-review", "代码审查"),
    ("plugins", "插件"),
    ("plugins-reference", "插件参考"),
    ("plugin-marketplaces", "插件市场"),
    ("discover-plugins", "发现插件"),
    ("third-party-integrations", "第三方集成"),
    ("amazon-bedrock", "Amazon Bedrock"),
    ("google-vertex-ai", "Google Vertex AI"),
    ("microsoft-foundry", "Microsoft Foundry"),
    ("llm-gateway", "LLM 网关"),
    ("network-config", "网络配置"),
    ("server-managed-settings", "服务端管理设置"),
    ("devcontainer", "Dev Container"),
    ("data-usage", "数据使用"),
    ("legal-and-compliance", "法律合规"),
    ("zero-data-retention", "零数据保留"),
    ("troubleshooting", "故障排除"),
    # changelog 跳过（.md 端点重定向到 GitHub 返回 1.3MB HTML）
    # ("changelog", "更新日志"),
]

BASE_URL = "https://code.claude.com/docs/zh-CN"

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

h4 {
    font-size: 13px;
    color: #555;
    margin-top: 16px;
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
    font-size: 13px;
}

.toc .toc-num {
    color: #999;
    margin-right: 8px;
}

.page-section {
    page-break-before: always;
}

.page-source {
    color: #999;
    font-size: 11px;
    margin-bottom: 20px;
}

a {
    color: #2980b9;
    text-decoration: none;
}

img {
    max-width: 100%;
}

hr {
    border: none;
    border-top: 1px solid #ddd;
    margin: 20px 0;
}
"""


def fetch_md(slug):
    """用 curl 抓取 .md 格式的文档页面"""
    url = f"{BASE_URL}/{slug}.md"
    try:
        result = subprocess.run(
            ['curl', '-s', '-f', '--max-time', '60', url],
            capture_output=True, text=True
        )
        if result.returncode == 0 and result.stdout.strip():
            return result.stdout
        return None
    except Exception as e:
        print(f"Error: {e}")
        return None


def clean_markdown(md_text):
    """清理 Markdown 中的 Mintlify 特有标签"""
    # 移除文件开头的 llms.txt 提示行
    md_text = re.sub(r'^> ## Documentation Index.*?(?=\n[^>])', '', md_text, flags=re.DOTALL)

    # 转换 Mintlify 组件为普通 Markdown
    # <Note> → blockquote
    md_text = re.sub(r'<Note>\s*', '\n> **注意**: ', md_text)
    md_text = re.sub(r'</Note>', '\n', md_text)

    # <Tip> → blockquote
    md_text = re.sub(r'<Tip>\s*', '\n> **提示**: ', md_text)
    md_text = re.sub(r'</Tip>', '\n', md_text)

    # <Info> → blockquote
    md_text = re.sub(r'<Info>\s*', '\n> **信息**: ', md_text)
    md_text = re.sub(r'</Info>', '\n', md_text)

    # <Warning> → blockquote
    md_text = re.sub(r'<Warning>\s*', '\n> **警告**: ', md_text)
    md_text = re.sub(r'</Warning>', '\n', md_text)

    # <Tabs>/<Tab> → 用标题标注
    md_text = re.sub(r'<Tabs>\s*', '\n', md_text)
    md_text = re.sub(r'</Tabs>\s*', '\n', md_text)
    md_text = re.sub(r'<Tab\s+title="([^"]*)">\s*', r'\n**\1**\n\n', md_text)
    md_text = re.sub(r'</Tab>\s*', '\n', md_text)

    # <Accordion> → 用标题标注
    md_text = re.sub(r'<AccordionGroup>\s*', '\n', md_text)
    md_text = re.sub(r'</AccordionGroup>\s*', '\n', md_text)
    md_text = re.sub(r'<Accordion\s+title="([^"]*)"[^>]*>\s*', r'\n**\1**\n\n', md_text)
    md_text = re.sub(r'</Accordion>\s*', '\n', md_text)

    # <Card> → 链接
    md_text = re.sub(
        r'<Card\s+title="([^"]*)"[^>]*href="([^"]*)"[^>]*>\s*(.*?)\s*</Card>',
        r'- **\1** (\2): \3',
        md_text, flags=re.DOTALL
    )

    # <CardGroup> → 移除
    md_text = re.sub(r'<CardGroup[^>]*>\s*', '\n', md_text)
    md_text = re.sub(r'</CardGroup>\s*', '\n', md_text)

    # <Steps>/<Step>
    md_text = re.sub(r'<Steps>\s*', '\n', md_text)
    md_text = re.sub(r'</Steps>\s*', '\n', md_text)
    md_text = re.sub(r'<Step\s+title="([^"]*)"[^>]*>\s*', r'\n#### \1\n\n', md_text)
    md_text = re.sub(r'</Step>\s*', '\n', md_text)

    # <Frame> → 移除包装
    md_text = re.sub(r'<Frame[^>]*>\s*', '\n', md_text)
    md_text = re.sub(r'</Frame>\s*', '\n', md_text)

    # <CodeGroup> → 移除包装
    md_text = re.sub(r'<CodeGroup>\s*', '\n', md_text)
    md_text = re.sub(r'</CodeGroup>\s*', '\n', md_text)

    # 移除代码块中的 theme={null}
    md_text = re.sub(r'```(\w+)\s+theme=\{null\}', r'```\1', md_text)

    # 移除剩余的 HTML 标签（保守处理）
    md_text = re.sub(r'</?(?:br|div|span|p)\s*/?>', '', md_text)

    # 清理多余空行
    md_text = re.sub(r'\n{4,}', '\n\n\n', md_text)

    return md_text.strip()


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


def build_html(pages_content):
    """构建完整的 HTML 文档"""
    parts = []

    # 封面
    parts.append("""
    <div class="cover">
        <h1>Claude Code 中文文档</h1>
        <p>code.claude.com/docs/zh-CN</p>
        <p>完整离线版</p>
        <br><br>
        <p style="color: #999; font-size: 13px;">共 {} 个文档页面</p>
    </div>
    """.format(len(pages_content)))

    # 目录
    toc_items = []
    for i, (slug, title, _) in enumerate(pages_content, 1):
        toc_items.append(
            f'<li><span class="toc-num">{i:02d}.</span> '
            f'<strong>{title}</strong></li>'
        )

    parts.append(f"""
    <div class="toc">
        <h1>目录</h1>
        <ul>{"".join(toc_items)}</ul>
    </div>
    """)

    # 各页面内容
    for slug, title, md_content in pages_content:
        parts.append('<div class="page-section">')
        if md_content:
            html_content = markdown_to_html(md_content)
            parts.append(html_content)
        else:
            parts.append(f'<h1>{title}</h1>')
            parts.append('<p style="color:red;">（此页面内容未能获取）</p>')
        parts.append(f'<p class="page-source">来源: {BASE_URL}/{slug}</p>')
        parts.append('</div>')

    from pygments.formatters import HtmlFormatter
    pygments_css = HtmlFormatter(style='friendly').get_style_defs('.highlight')

    html = f"""<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <title>Claude Code 中文文档</title>
    <style>{CSS}
    {pygments_css}
    </style>
</head>
<body>
{"".join(parts)}
</body>
</html>"""

    return html


def main():
    print("=" * 60)
    print("Claude Code 中文文档 PDF 生成器")
    print("=" * 60)
    print(f"\n共 {len(PAGES)} 个页面待抓取\n")

    pages_content = []
    success = 0
    failed = 0

    for i, (slug, title) in enumerate(PAGES, 1):
        print(f"[{i:02d}/{len(PAGES)}] {title} ({slug})...", end=" ", flush=True)

        md = fetch_md(slug)
        if md:
            # 检查是否返回了 HTML 而非 Markdown（如 changelog 重定向到 GitHub）
            if md.strip().startswith('<!DOCTYPE') or md.strip().startswith('<html'):
                md = None
                print("SKIP (HTML redirect)")
                pages_content.append((slug, title, None))
                failed += 1
                continue
            md = clean_markdown(md)
            pages_content.append((slug, title, md))
            print("OK")
            success += 1
        else:
            pages_content.append((slug, title, None))
            print("FAIL")
            failed += 1

    print(f"\n抓取完成: 成功 {success}, 失败 {failed}")

    # 构建 HTML
    print("\n正在构建 HTML...")
    html = build_html(pages_content)

    html_path = os.path.join(PROJECT_ROOT, "claude-code-docs-zh.html")
    with open(html_path, 'w', encoding='utf-8') as f:
        f.write(html)
    print(f"HTML 已保存: {html_path}")

    # 生成 PDF
    print("\n正在生成 PDF（使用 weasyprint）...")
    try:
        from weasyprint import HTML as WeasyHTML
        pdf_path = os.path.join(PROJECT_ROOT, "claude-code-docs-zh.pdf")
        WeasyHTML(string=html).write_pdf(pdf_path)
        size_mb = os.path.getsize(pdf_path) / (1024 * 1024)
        print(f"\nPDF 生成完成: {pdf_path} ({size_mb:.1f} MB)")
    except ImportError:
        print("\nweasyprint 未安装，仅生成了 HTML 文件。")
        print("安装: source /tmp/pdf-venv/bin/activate && pip install weasyprint")
    except Exception as e:
        print(f"\nPDF 生成失败: {e}")
        print("HTML 文件可在浏览器中打开并手动打印为 PDF")


if __name__ == "__main__":
    main()
