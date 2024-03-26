# jgc_exporter
[![Build Status][maven-build-image]][maven-build-url]
[![CodeCov][codecov-image]][codecov-url]

An exporter that can continuously analyze hotspot gc logs(Parallel, CMS, G1, ZGC, etc.)
# Running the exporter

Run the exporter as standalone HTTP server, prefer native executable if available:

start with jar:
```shell
# jdk require: 11+
#  os require: linux, windows
sh bin/run.sh --jar

# background running
sh bin/run.sh --jar --daemon
```

start with native executable (only linux now, windows is on the way):
```shell
# os require: linux
sh bin/run.sh --native

# background running
sh bin/run.sh --native --daemon
```

A simple `config.yml` looks like below:
```yaml

# wildcard pattern is allowed in any level of paths, most of the time you only need to modify this configuration
fileGlobPattern: /path/to/some*/*.log
```

Fetch the metrics:
```
http://0.0.0.0:5898/metrics
```

# Configuration
| Name            | Description                                                                  |
|-----------------|------------------------------------------------------------------------------|
| hostPort        | Host and port that http server binds, default is 0.0.0.0:5898                |
| fileGlobPattern | Wildcard pattern of gc log file path, separate multiple paths with commas(,) |
| idleTimeout     | Milliseconds before closing idle(no update) files, default is 1 hour         |
| watchInterval   | Time interval for scanning matching files (ms)                               |
| readInterval    | Time to sleep between files reading empty (ms)                               |

# Metric
| Name                                       | type    | labels               | Description                      |
|--------------------------------------------|---------|----------------------|----------------------------------|
| jgc_log_lines_total                        | counter | path, host           | Number of process log lines      |
| jgc_event_duration_seconds                 | summary | path, host, category | Duration of GC events            |
| jgc_event_pause_duration_seconds           | summary | path, host, category | Duration of GC pause events      |
| jgc_heap_occupancy_before_collection_bytes | gauge   | path, host           | Heap occupancy before collection |
| jgc_heap_occupancy_after_collection_bytes  | gauge   | path, host           | Heap occupancy after collection  |

See more [metrics](https://github.com/loyispa/jgc_exporter/blob/main/src/main/java/prometheus/exporter/jgc/metric/MetricRegistry.java) related to specific garbage-collection algorithms.

# Build
``` shell
./mvnw clean package
```

# Native build
The exporter can be converted into native executables by installing [graalvm](https://www.graalvm.org/downloads/)(17 and 21).

- native dynamic link:
``` shell
./mvnw -Pnative clean package
```

- native static link with [musl](https://musl.cc/):
``` shell
./mvnw -Pnative-static-musl clean package
```

# Suggestions

*Note: The lack of essential jvm flags may miss some indicators, so we recommend that you set up your target java process as follows:*

- jdk8 and previous versions
```
-verbose:gc -XX:+PrintGCDetails -XX:+PrintGCDateStamps
```

- jdk9 and later versions
```
-verbose:gc -Xlog:gc*=info,gc+heap=debug,gc+phases=debug:file=xxx/gc.log:t,tags
```

# Contributing
The exporter is still iterating frequently, all contributions are welcome. If you want a new feature, please raise an issue first.

[maven-build-image]: https://github.com/loyispa/jgc_exporter/workflows/Java%20CI%20with%20Maven/badge.svg
[maven-build-url]: https://github.com/loyispa/jgc_exporter/actions/workflows/maven.yaml
[codecov-image]: https://codecov.io/gh/loyispa/jgc_exporter/branch/main/graph/badge.svg
[codecov-url]: https://app.codecov.io/gh/loyispa/jgc_exporter
