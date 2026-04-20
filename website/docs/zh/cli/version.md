# `pyscn version`

输出版本和构建信息。

```text
pyscn version [flags]
```

## 选项

| 选项 | 说明 |
| --- | --- |
| `-s, --short` | 仅输出版本号，便于在脚本中使用。 |

## 示例

```bash
$ pyscn version
pyscn v0.2.0
  commit:   a3671f4
  built:    2026-02-14T08:42:11Z
  go:       go1.24.6
  platform: darwin/arm64

$ pyscn version --short
0.2.0
```

## 在脚本中使用

```bash
# 保存版本号供后续使用
VERSION=$(pyscn version --short)

# 检查版本兼容性
if [ "$(pyscn version --short)" != "0.2.0" ]; then
  echo "Expected pyscn 0.2.0" >&2
  exit 1
fi
```

`pyscn version` 始终以退出码 `0` 退出。
