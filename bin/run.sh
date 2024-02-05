#!/usr/bin/env bash

BASE_DIR=$(dirname $0)/..

export JGC_EXPORTER_BASE_DIR="${BASE_DIR}"

NATIVE_MODE="false"

# prefer native executable
if test -e "${BASE_DIR}"/native/jgc_exporter; then
  NATIVE_MODE="true"
fi

if [ "$1" == "--jar" ]; then
    NATIVE_MODE="false"
fi

# Which java to use
if [ -z "$JAVA_HOME" ]; then
  JAVA="java"
else
  JAVA="$JAVA_HOME/bin/java"
fi

JGC_NATIVE_HEAP_OPTS="-Xmx64m"

JGC_JAR_HEAP_OPTS="-Xmx256m"

if [ "x$NATIVE_MODE" = "xtrue" ]; then
  echo "Using native executable: ${JGC_NATIVE_HEAP_OPTS}"
  nohup "$BASE_DIR"/native/jgc_exporter $JGC_NATIVE_HEAP_OPTS "$BASE_DIR"/conf/config.yaml > /dev/null 2>&1 &
else
  echo "Using jar executable: ${JGC_JAR_HEAP_OPTS}"
  nohup "$JAVA" $JGC_JAR_HEAP_OPTS -jar "$BASE_DIR"/lib/jgc_exporter.jar "$BASE_DIR"/conf/config.yaml > /dev/null 2>&1 &
fi
