# `pyscn version`

バージョンとビルド情報を表示します。

```text
pyscn version [flags]
```

## フラグ

| フラグ | 説明 |
| --- | --- |
| `-s, --short` | バージョン番号のみを出力します。スクリプトでの利用に便利です。 |

## 使用例

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

## スクリプトでの利用

```bash
# Save for later use
VERSION=$(pyscn version --short)

# Guard against incompatible versions
if [ "$(pyscn version --short)" != "0.2.0" ]; then
  echo "Expected pyscn 0.2.0" >&2
  exit 1
fi
```

`pyscn version` は常に終了コード `0` で終了します。
