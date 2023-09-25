# jgc_exporter
[![Build Status][maven-build-image]][maven-build-url]
[![CodeCov][codecov-image]][codecov-url]

An exporter that can continuously analyze hotspot garbage collection log and automatically detect Garbage collection algorithms. The parsing ability mainly relies on the [gctoolkit](https://github.com/microsoft/gctoolkit), which supports most mainstream garbage collections, such as CMS, G1, ZGC, etc.
# Running the exporter
JDK require: 11+

Run the exporter as standalone HTTP server:
```shell
java -jar jgc_exporter.jar /path/to/config.yml
```

A simple `config.yml` looks like below:
```yaml
fileRegexPattern: /path/to/serviceA/gc.*.log
```

Fetch the metrics:
```agsl
http://0.0.0.0:5898/metrics
```

# Configuration
| Name             | Description                                                                 |
|------------------|-----------------------------------------------------------------------------|
| hostPort         | Host and port that http server binds, default is 0.0.0.0:5898               |
| fileRegexPattern | Regex pattern of gc log file path                                           |
| idleTimeout      | Time (ms) to close idle files, default is 10 minutes(600,000ms)             |
| batchSize        | Maximum number of lines per log read, default is 128                        |   
| analysePeriod    | Minimum time interval between two analyses, default is 10 seconds(10,000ms) |   

# Building
```
./mvnw clean package
```

# Contributing
All contributions are welcome, docs, bugfixes or features.

[maven-build-image]: https://github.com/loyispa/jgc_exporter/workflows/Java%20CI%20with%20Maven/badge.svg
[maven-build-url]: https://github.com/loyispa/jgc_exporter/actions/workflows/maven.yaml
[codecov-image]: https://codecov.io/gh/loyispa/jgc_exporter/branch/main/graph/badge.svg
[codecov-url]: https://app.codecov.io/gh/loyispa/jgc_exporter