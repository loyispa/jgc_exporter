#!/usr/bin/env bash

BASE_DIR=$(dirname $0)/..

NATIVE_MODE="false"

# prefer native executable
if test -e "${BASE_DIR}"/lib/jgc_exporter; then
  NATIVE_MODE="true"
fi

CONSOLE_OUTPUT_FILE="$BASE_DIR"/console.out

JGC_HEAP_OPTS="-Xmx256m"

JGC_NATIVE_HEAP_OPTS="-Xmx64m"

# Which java to use
if [ -z "$JAVA_HOME" ]; then
  JAVA="java"
else
  JAVA="$JAVA_HOME/bin/java"
fi

if [ "x$NATIVE_MODE" = "xtrue" ]; then
  nohup "$BASE_DIR"/lib/jgc_exporter $JGC_NATIVE_HEAP_OPTS "$BASE_DIR"/conf/config.yaml > "$CONSOLE_OUTPUT_FILE" 2>&1 < /dev/null &
else
  nohup "$JAVA" $JGC_HEAP_OPTS -jar "$BASE_DIR"/lib/jgc_exporter.jar "$BASE_DIR"/conf/config.yaml > "$CONSOLE_OUTPUT_FILE" 2>&1 < /dev/null &
fi

