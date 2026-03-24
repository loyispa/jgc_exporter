# JGC Exporter

Real-time Java GC log collector and Prometheus exporter with health visualization.

## Features

- **Multi-collector support**: G1, CMS, ZGC with PreUnified (JDK8) and Unified (JDK9+) log formats
- **Real-time tail**: Glob-based file discovery with automatic rotation handling
- **Prometheus metrics**: `jgc_*` prefix, compatible with existing dashboards
- **Health dashboard**: Web UI at `/ui` with overall health status (Healthy/Warning/Critical), Full GC alerts, heap trends, and throughput trend chart

## Quick Start

```bash
# Build
make build

# Run with glob pattern (quote the pattern so the shell does not expand it)
./jgc_exporter --glob-path '/var/logs/**/gc.log'
```

## Configuration

| Flag  | Default | Description |
|------|---------|-------------|
| `--glob-path` | *required* | Glob pattern for GC log files |
| `--port` | `5898` | HTTP port for /metrics and /ui |
| `--idle-timeout` | `3600000` | Idle file close timeout (ms) |
| `--watch-interval` | `10000` | File scan interval (ms) |


## HTTP Endpoints

| Path | Description |
|------|-------------|
| `GET /metrics` | Prometheus exposition format |
| `GET /ui` | GC health dashboard |
| `GET /api/dashboard` | Health data as JSON (`multi_window` fields use `last_1m`, `last_5m`, `last_1h`) |

## Health Status Rules

| Status | Condition |
|--------|-----------|
| **Critical** | ≥2 Full GCs in 1min, OR max STW pause >500ms, OR heap >90% |
| **Warning** | Any Full GC in 1min, OR max STW pause >200ms, OR heap >80% |
| **Healthy** | None of the above |

## Prometheus Metrics

Core metrics with `jgc_` prefix and labels `path`, `host`:

- `jgc_event_duration_seconds` — GC event duration summary
- `jgc_event_pause_duration_seconds` — Pause-only duration summary
- `jgc_heap_used_before_collection_bytes` / `jgc_heap_used_after_collection_bytes` — Heap **used** bytes before/after collection
- `jgc_heap_size_*_bytes` — Heap **committed capacity** (total) from the log; not the same as used bytes
- `jgc_metaspace_*_bytes` — class metadata (non-heap)
- `jgc_g1_*_references` / `jgc_cms_*_references` and matching `*_pause_duration_seconds` — Reference processing from Remark (PrintReferenceGC); G1 and CMS are separate metric names
- `jgc_g1_*_heap_*_regions` — G1 per-region-type counts (Eden / Survivor / Old / Humongous / Archive); no generic young/old gauges
- `jgc_zgc_*` — ZGC-specific (MMU, load, phase durations)
- `jgc_log_lines_total` — Processed log lines (labels: path, host, gc_type)
