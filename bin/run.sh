#!/usr/bin/env bash

BASE_DIR=$(dirname $0)/..

EXPORTER_BASE_DIR="${BASE_DIR}"

NATIVE_MODE="false"

# prefer native executable
if test -e "${BASE_DIR}"/lib/jgc_exporter; then
  NATIVE_MODE="true"
fi

JGC_HEAP_OPTS="-Xmx256m"

JGC_NATIVE_HEAP_OPTS="-Xmx64m"

# Which java to use
if [ -z "$JAVA_HOME" ]; then
  JAVA="java"
else
  JAVA="$JAVA_HOME/bin/java"
fi

if [ "x$NATIVE_MODE" = "xtrue" ]; then
  nohup "$BASE_DIR"/lib/jgc_exporter $JGC_NATIVE_HEAP_OPTS "$BASE_DIR"/conf/config.yaml > /dev/null 2>&1 &
else
  nohup "$JAVA" $JGC_HEAP_OPTS -jar "$BASE_DIR"/lib/jgc_exporter.jar "$BASE_DIR"/conf/config.yaml > /dev/null 2>&1 &
fi
