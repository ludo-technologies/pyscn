# Output Formats

pyscn writes analysis results to files under an output directory. All formats share stable field semantics across patch releases.

## Output directory

Default: `.pyscn/reports/` under the current working directory.

Configurable via `.pyscn.toml`:

```toml
[output]
directory = "build/reports"
```

## Filename pattern

```
{command}_YYYYMMDD_HHMMSS.{ext}
```

`{command}` is `analyze` (the only pyscn command that writes reports). The timestamp is local time. Existing files are never overwritten.

## Supported formats

| Format | Extension | Flag          | Specification                    |
| ------ | --------- | ------------- | -------------------------------- |
| text   | —         | (terminal)    | human-readable, not stable       |
| json   | `.json`   | `--json`      | [schemas.md](schemas.md)         |
| yaml   | `.yaml`   | `--yaml`      | [schemas.md](schemas.md)         |
| csv    | `.csv`    | `--csv`       | [schemas.md](schemas.md)         |
| html   | `.html`   | `--html` (default) | [html-report.md](html-report.md) |

The `text` format is intended for terminal display and has no stability contract; its layout may change between any releases.

## Stability contract

Across patch and minor releases within the same major version:

- **Stable**: field names, types, and semantic meaning in `json`, `yaml`, and `csv`.
- **May change**: ordering of array elements, addition of new fields, addition of new top-level sections, cosmetic changes to `text` and `html`.
- **Breaking changes**: restricted to major version bumps (removal or rename of fields, change of field types).

Third-party integrations should ignore unknown fields and not rely on field ordering within objects.
