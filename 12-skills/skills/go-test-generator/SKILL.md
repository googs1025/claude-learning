---
name: go-test-generator
description: 为指定的 Go 文件生成表驱动测试
argument-hint: "[filename]"
user-invocable: true
disable-model-invocation: true
allowed-tools:
  - Read
  - Grep
  - Glob
  - Write
---

# Go 测试生成器

为 `$ARGUMENTS` 生成全面的 Go 单元测试。

## 要求

1. **读取目标文件**: 分析 `$ARGUMENTS` 中所有导出的函数和方法
2. **使用表驱动测试**: 所有测试都必须使用 `[]struct{ name string; ... }` 的表驱动模式
3. **测试命名**: 使用 `Test_函数名` 格式
4. **覆盖场景**:
   - 正常输入
   - 边界值（空字符串、零值、nil）
   - 错误情况
5. **文件命名**: 输出文件名为原文件名加 `_test` 后缀

## 模板

```go
func Test_FunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name:  "正常情况",
            input: ...,
            want:  ...,
        },
        {
            name:    "错误情况",
            input:   ...,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("错误 = %v, 期望错误 = %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("结果 = %v, 期望 = %v", got, tt.want)
            }
        })
    }
}
```

## 注意事项

- 不要 mock 外部依赖，只测试纯逻辑函数
- 如果函数依赖外部服务，添加 `t.Skip("需要外部服务")` 的测试桩
- 使用 `t.Helper()` 标记辅助函数
