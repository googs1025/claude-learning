#!/bin/bash
# Hook: 06_run_tests.sh
# 事件: PostToolUse (Write)
# 作用: 检测到源码文件变更后，自动运行对应的测试套件
#       避免 Claude 修改代码后遗忘运行测试

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

EXT="${FILE_PATH##*.}"
FILENAME=$(basename "$FILE_PATH")
AUDIT_LOG="/tmp/claude_hooks_audit.log"

# 跳过测试文件本身（避免循环）和配置文件
if echo "$FILENAME" | grep -qE "_test\.(go|py|js|ts)$|\.test\.(js|ts)$|spec\.(js|ts)$"; then
    exit 0
fi

# 获取文件所在目录，在该目录运行测试
FILE_DIR=$(dirname "$FILE_PATH")

case "$EXT" in
    go)
        # 查找包含测试文件的最近目录
        if ls "$FILE_DIR"/*_test.go &>/dev/null 2>&1; then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [run_tests] 运行 Go 测试: $FILE_DIR" >> "$AUDIT_LOG"
            echo "[run_tests] 检测到 Go 源码变更，运行测试..." >&2
            cd "$FILE_DIR" && go test ./... -timeout 30s 2>&1 | tail -5 >&2
        fi
        ;;
    py)
        if command -v pytest &>/dev/null && ls "$FILE_DIR"/test_*.py &>/dev/null 2>&1; then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [run_tests] 运行 pytest: $FILE_DIR" >> "$AUDIT_LOG"
            echo "[run_tests] 检测到 Python 源码变更，运行测试..." >&2
            cd "$FILE_DIR" && pytest -q --tb=short 2>&1 | tail -10 >&2
        fi
        ;;
    js|ts)
        if [ -f "$FILE_DIR/package.json" ] || [ -f "$(git rev-parse --show-toplevel 2>/dev/null)/package.json" ]; then
            ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "$FILE_DIR")
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] [run_tests] 运行 npm test: $ROOT" >> "$AUDIT_LOG"
            echo "[run_tests] 检测到 JS/TS 源码变更，运行测试..." >&2
            cd "$ROOT" && npm test --silent 2>&1 | tail -10 >&2
        fi
        ;;
esac

exit 0
