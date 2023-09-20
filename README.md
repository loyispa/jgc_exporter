# jgc_exporter
[![Build Status][maven-build-image]][maven-build-url]
[![CodeCov][codecov-image]][codecov-url]

JGC_exporter is an exporter that can continuous analysis hotspot jvm garbage collection log files, and automatically detect garbage collectors. The ability of parsing depends on the [gctoolkit](https://github.com/microsoft/gctoolkit), supports most mainstream garbage collections, such as ParNew、CMS、G1、ZGC, etc.
# Running the exporter
JDK require: 11+

Run the exporter as standalone HTTP server:
```shell
java -jar jgc_exporter.jar /path/to/config.yaml
```

A simple `config.yaml` looks like below:
```yaml
hostPort: 127.0.0.1:5898
fileRegexPattern: /path/to/serviceA/gc.*.log,/path/to/serviceB/gc.*.log
```

# Building
```
./mvnw clean package
```

# Contributing
All external contributions are welcome, docs, bugfixes and features.

[maven-build-image]: https://github.com/loyispa/jgc_exporter/workflows/Java%20CI%20with%20Maven/badge.svg
[maven-build-url]: https://github.com/loyispa/jgc_exporter/actions/workflows/maven.yaml
[codecov-image]: https://codecov.io/gh/loyispa/jgc_exporter/branch/main/graph/badge.svg
[codecov-url]: https://app.codecov.io/gh/loyispa/jgc_exporter