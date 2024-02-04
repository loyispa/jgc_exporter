# jgc_exporter
[![Build Status][maven-build-image]][maven-build-url]
[![CodeCov][codecov-image]][codecov-url]

An exporter that can continuously analyze hotspot garbage collection log and automatically detect Garbage collection algorithms. The basic ability relies on the [gctoolkit](https://github.com/microsoft/gctoolkit) library, which supports most mainstream garbage collections, such as CMS, G1, ZGC, etc.
# Running the exporter
JDK require: 11+

OS require: linux

Run the exporter as standalone HTTP server:
```shell
java -jar jgc_exporter.jar /path/to/config.yml
```

A simple `config.yml` looks like below:
```yaml

# glob pattern is allowed in any level of paths
fileGlobPattern: /path/to/some*/*.log
```

Fetch the metrics:
```agsl
http://0.0.0.0:5898/metrics
```

# Configuration
| Name             | Description                                                            |
|------------------|------------------------------------------------------------------------|
| hostPort         | Host and port that http server binds, default is 0.0.0.0:5898          |
| fileGlobPattern  | Glob pattern of gc log file path, separate multiple paths with commas  |
| idleTimeout      | Time (ms) to close idle files, default is 10 minutes(600,000ms)        |

# Metric
| Name                                       | type    | labels         | Description                      |
|--------------------------------------------|---------|----------------|----------------------------------|
| jgc_log_lines_total                        | counter | path           | Number of process log lines      |
| jgc_event_duration_seconds                 | summary | path, category | Duration of GC events            |
| jgc_event_pause_duration_seconds           | summary | path, category | Duration of GC pause events      |
| jgc_heap_occupancy_before_collection_bytes | gauge   | path           | Heap occupancy before collection |
| jgc_heap_occupancy_after_collection_bytes  | gauge   | path           | Heap occupancy after collection  |

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
All contributions are welcome, docs, bugfixes and features.

[maven-build-image]: https://github.com/loyispa/jgc_exporter/workflows/Java%20CI%20with%20Maven/badge.svg
[maven-build-url]: https://github.com/loyispa/jgc_exporter/actions/workflows/maven.yaml
[codecov-image]: https://codecov.io/gh/loyispa/jgc_exporter/branch/main/graph/badge.svg
[codecov-url]: https://app.codecov.io/gh/loyispa/jgc_exporter
