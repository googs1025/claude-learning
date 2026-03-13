#!/bin/bash
# Hook: 05_auto_format.sh
# 事件: PostToolUse (Write)
# 作用: Claude 写入文件后，根据扩展名自动运行代码格式化工具
#       支持 Go (gofmt)、JavaScript/TypeScript (prettier)、Python (black)

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

# 获取文件扩展名
EXT="${FILE_PATH##*.}"
AUDIT_LOG="/tmp/claude_hooks_audit.log"

case "$EXT" in
    go)
        if command -v gofmt &>/dev/null; then
            gofmt -w "$FILE_PATH" 2>/dev/null
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [auto_format] ✅ gofmt: $FILE_PATH" >> "$AUDIT_LOG"
            echo "[auto_format] 已对 $FILE_PATH 运行 gofmt" >&2
        fi
        ;;
    js|jsx|ts|tsx|json|css|scss|html)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE_PATH" 2>/dev/null
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [auto_format] ✅ prettier: $FILE_PATH" >> "$AUDIT_LOG"
            echo "[auto_format] 已对 $FILE_PATH 运行 prettier" >&2
        fi
        ;;
    py)
        if command -v black &>/dev/null; then
            black --quiet "$FILE_PATH" 2>/dev/null
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [auto_format] ✅ black: $FILE_PATH" >> "$AUDIT_LOG"
            echo "[auto_format] 已对 $FILE_PATH 运行 black" >&2
        elif command -v autopep8 &>/dev/null; then
            autopep8 --in-place "$FILE_PATH" 2>/dev/null
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [auto_format] ✅ autopep8: $FILE_PATH" >> "$AUDIT_LOG"
        fi
        ;;
    sh|bash)
        if command -v shfmt &>/dev/null; then
            shfmt -w "$FILE_PATH" 2>/dev/null
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [auto_format] ✅ shfmt: $FILE_PATH" >> "$AUDIT_LOG"
        fi
        ;;
    *)
        # 不支持的扩展名，静默跳过
        ;;
esac

exit 0
