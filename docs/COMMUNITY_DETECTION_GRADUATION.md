# Community Detection Graduation

Issue #605 graduates module community detection into the default `pyscn analyze`
run. The default-on decision is based on the Phase 2 scoring and output work:
community risk is deterministic, bounded, and does not penalize users when
community detection is disabled or produces fewer than two communities.

## Default Behavior

- `pyscn analyze <path>` runs community detection with the other default
  analyses.
- `pyscn analyze --select ...` runs communities only when `communities` appears
  in the selected list.
- `pyscn analyze --skip-communities <path>` disables community detection for the
  run.
- `[communities] enabled = false` in `.pyscn.toml` or
  `[tool.pyscn.communities]` disables community detection for default analyze
  runs. `--select communities` remains an explicit per-run request.
- MCP `analyze_code` follows the same selection rule: omitted `analyses` runs
  communities; an explicit list must include `communities`.

## Graduation Criteria

Run these checks before changing community defaults, thresholds, graph
construction, or score weights.

### Performance

Reference target:

- `testdata/python/community_bridge`
- One medium open-source Python snapshot used by release validation

Budget:

- Community detection p95 latency must stay within 10% of the previous release
  baseline on both references.
- The benchmark should not allocate asymptotically more memory on the medium
  graph fixture.

Commands:

```sh
make bench-communities
go test -run '^$' -bench 'BenchmarkDetectCommunitiesLeiden_MediumGraph' -benchmem -count=10 ./internal/analyzer
```

For end-to-end fixture timing, build once and run the same command repeatedly
against the fixed fixture:

```sh
make build
for i in 1 2 3 4 5 6 7 8 9 10; do
  ./pyscn analyze --json testdata/python/community_bridge >/tmp/pyscn-community-bridge-$i.json
done
```

Use the previous release's numbers as the baseline; record the p95 and the
machine class in the issue or release notes.

### Signal Quality

Community mismatch and risk scores must keep the expected ordering on known
fixtures:

- `testdata/python/community_separated` should remain lower risk than
  `testdata/python/community_bridge`.
- Package mismatch fixtures should populate `split_packages` and
  `mixed_communities`.
- Layer mismatch fixtures should populate `cross_layer_communities` and
  `layer_bridge_modules` when architecture layers are configured.

Commands:

```sh
go test ./internal/analyzer ./service ./domain -run 'Community|AnalyzeScore'
go test ./app ./cmd/pyscn ./mcp -run 'Community|Communities'
```

### Backward Compatibility

- Disabled communities must not change Health Score category penalties.
- Trivial community results with fewer than two communities must score as
  `community_score=100` and `community_risk_score=0`.
- `--skip-communities` and explicit `[communities] enabled = false` must keep
  community output absent.
- Report size impact should be checked on the medium OSS snapshot when JSON or
  HTML report size is release-sensitive.

### CI Noise

- JSON/YAML community fields must remain sorted by stable identifiers.
- HTML graph payloads must be deterministic for a fixed fixture and
  configuration.
- Repeated runs on `testdata/python/community_bridge` must produce stable
  community ids, bridge modules, mismatch lists, and rounded numeric fields.

Command:

```sh
go test ./service ./cmd/pyscn ./mcp -run 'Community|Communities|AnalyzeCommand'
```

## Decision Record

Decision for issue #605: flip the default to on.

Rationale:

- Community scoring is bounded and only affects the Health Score when the
  analyzer ran and detected at least two communities.
- The CLI, config, and MCP selection semantics now have explicit opt-out paths.
- Existing Phase 2 fixtures cover package mismatch, layer mismatch, bridge
  modules, risk-score ordering, and deterministic JSON/HTML output.
