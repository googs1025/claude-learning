#!/bin/bash
# Hook: 03_backup_before_edit.sh
# 事件: PreToolUse (Write|Edit)
# 作用: 在文件被覆写或编辑前，自动备份到 /tmp/claude_backup/
#       让你随时可以恢复意外修改的文件

# 从 stdin 读取 Claude Code 传入的 JSON
INPUT=$(cat)

# 提取工具名和目标文件路径
TOOL_NAME=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_name',''))" 2>/dev/null)
FILE_PATH=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin).get('tool_input',{}); print(d.get('file_path', d.get('path','')))" 2>/dev/null)

# 仅处理 Write 和 Edit 工具
if [ "$TOOL_NAME" != "Write" ] && [ "$TOOL_NAME" != "Edit" ] && [ "$TOOL_NAME" != "MultiEdit" ]; then
    exit 0
fi

# 文件必须存在才能备份（新建文件无需备份）
if [ -z "$FILE_PATH" ] || [ ! -f "$FILE_PATH" ]; then
    exit 0
fi

# 创建备份目录（按日期组织）
BACKUP_DIR="/tmp/claude_backup/$(date '+%Y%m%d')"
mkdir -p "$BACKUP_DIR"

# 生成备份文件名：原路径中 / 替换为 _，加上时间戳
SAFE_PATH=$(echo "$FILE_PATH" | sed 's|/|_|g' | sed 's|^_||')
TIMESTAMP=$(date '+%H%M%S')
BACKUP_FILE="$BACKUP_DIR/${TIMESTAMP}_${SAFE_PATH}"

# 执行备份
if cp "$FILE_PATH" "$BACKUP_FILE" 2>/dev/null; then
    # 记录到审计日志（stderr 会显示在 Claude 的日志中）
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [backup] ✅ 已备份: $FILE_PATH → $BACKUP_FILE" >&2
    AUDIT_LOG="/tmp/claude_hooks_audit.log"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [backup] $FILE_PATH → $BACKUP_FILE" >> "$AUDIT_LOG"
fi

# 备份是辅助操作，不阻断工具执行
exit 0
