#!/bin/bash
# Hook: 04_sensitive_data.sh
# 事件: PostToolUse (Write)
# 作用: 检测 Claude 写入文件中是否包含 API Key、密码、Token 等敏感数据
#       发现后输出警告（不阻断，但提醒用户复查）

# 从 stdin 读取 Claude Code 传入的 JSON
INPUT=$(cat)

# 提取工具名和目标文件路径
TOOL_NAME=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_name',''))" 2>/dev/null)
FILE_PATH=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_input',{}).get('file_path',''))" 2>/dev/null)

# 仅处理 Write 工具
if [ "$TOOL_NAME" != "Write" ]; then
    exit 0
fi

# 文件必须存在
if [ -z "$FILE_PATH" ] || [ ! -f "$FILE_PATH" ]; then
    exit 0
fi

# 敏感数据正则模式
SENSITIVE_PATTERNS=(
    # API Keys
    "sk-[a-zA-Z0-9]{20,}"                    # Anthropic / OpenAI API Key
    "AKIA[0-9A-Z]{16}"                        # AWS Access Key
    "AIza[0-9A-Za-z_-]{35}"                   # Google API Key
    "ghp_[a-zA-Z0-9]{36}"                     # GitHub Personal Access Token
    "ghs_[a-zA-Z0-9]{36}"                     # GitHub App Token
    # 密码和密钥
    "password\s*=\s*['\"][^'\"]{6,}"          # 明文密码赋值
    "passwd\s*=\s*['\"][^'\"]{6,}"
    "secret\s*=\s*['\"][^'\"]{6,}"
    "private_key\s*="
    # 数据库连接串
    "mongodb\+srv://[^:]+:[^@]+"              # MongoDB Atlas
    "postgresql://[^:]+:[^@]+"                # PostgreSQL
    "mysql://[^:]+:[^@]+"                     # MySQL
)

AUDIT_LOG="/tmp/claude_hooks_audit.log"
FOUND=0

for pattern in "${SENSITIVE_PATTERNS[@]}"; do
    if grep -qE "$pattern" "$FILE_PATH" 2>/dev/null; then
        FOUND=1
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] [sensitive_data] ⚠️  在 $FILE_PATH 中检测到敏感数据模式: $pattern" >> "$AUDIT_LOG"
    fi
done

if [ $FOUND -eq 1 ]; then
    # PostToolUse 不能阻断（工具已执行），但可通过 stderr 向 Claude 发出警告
    cat >&2 <<EOF
⚠️  [敏感数据警告] 在文件 $FILE_PATH 中检测到可能的 API Key 或密码。
请确认：
1. 是否为示例代码（用占位符替换真实值）
2. 是否已将该文件加入 .gitignore
3. 如为生产密钥，请立即撤销并重新生成
EOF
fi

exit 0
