#!/bin/bash
# Hook: 02_file_protector.sh
# 事件: PreToolUse (Write)
# 作用: 阻止 Claude 直接覆写敏感配置文件（.env、credentials、secrets 等）

# 从 stdin 读取 Claude Code 传入的 JSON
INPUT=$(cat)

# 提取工具名和目标文件路径
TOOL_NAME=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_name',''))" 2>/dev/null)
FILE_PATH=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_input',{}).get('file_path',''))" 2>/dev/null)

# 仅处理 Write 工具
if [ "$TOOL_NAME" != "Write" ]; then
    exit 0
fi

# 敏感文件模式（文件名匹配）
SENSITIVE_PATTERNS=(
    "\.env$"
    "\.env\."
    "credentials"
    "secrets"
    "\.pem$"
    "\.key$"
    "\.p12$"
    "\.pfx$"
    "id_rsa"
    "id_ed25519"
    "\.htpasswd$"
    "auth\.json$"
    "config\.secret"
    "keystore"
)

FILENAME=$(basename "$FILE_PATH")
AUDIT_LOG="/tmp/claude_hooks_audit.log"
echo "[$(date '+%Y-%m-%d %H:%M:%S')] [file_protector] 检查文件: $FILE_PATH" >> "$AUDIT_LOG"

for pattern in "${SENSITIVE_PATTERNS[@]}"; do
    if echo "$FILENAME" | grep -qiE "$pattern" 2>/dev/null || echo "$FILE_PATH" | grep -qiE "$pattern" 2>/dev/null; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] [file_protector] 🔒 已阻断敏感文件写入: $FILE_PATH" >> "$AUDIT_LOG"

        cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "deny",
    "permissionDecisionReason": "文件保护守卫：'$FILE_PATH' 是敏感配置文件，禁止 AI 自动覆写。如需修改，请手动编辑。"
  }
}
EOF
        exit 0
    fi
done

# 文件安全，允许写入
exit 0
