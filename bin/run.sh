#!/usr/bin/env bash

BASE_DIR=$(dirname $0)/..

export JGC_EXPORTER_BASE_DIR="${BASE_DIR}"

# daemon mode
DAEMON_MODE="false"

# prefer native mode
if test -e "${BASE_DIR}"/native/jgc_exporter; then
  NATIVE_MODE="true"
else
  NATIVE_MODE="false"
fi

while [ $# -gt 0 ]; do
  case $1 in
  --daemon)
    DAEMON_MODE="true"
    ;;
  --jar)
    NATIVE_MODE="false"
    ;;
  --native)
    NATIVE_MODE="true"
    ;;
  esac
  shift
done

CONFIG_PATH="$BASE_DIR"/conf/config.yaml

NATIVE_EXECUTABLE="$BASE_DIR"/native/jgc_exporter

JAR_EXECUTABLE="$BASE_DIR"/lib/jgc_exporter.jar

if [ "x$NATIVE_MODE" = "xtrue" ]; then
  JGC_HEAP_OPTS="-Xmx64m"

  echo "Using native executable" $JGC_HEAP_OPTS

  if [ "x$DAEMON_MODE" = "xtrue" ]; then
    nohup "$NATIVE_EXECUTABLE" $JGC_HEAP_OPTS "$CONFIG_PATH" >/dev/null 2>&1 &
  else
    exec "$NATIVE_EXECUTABLE" $JGC_HEAP_OPTS "$CONFIG_PATH"
  fi

else
  JGC_HEAP_OPTS="-Xmx256m"

  echo "Using jar executable" $JGC_HEAP_OPTS

  if [ -z "$JAVA_HOME" ]; then
    JAVA="java"
  else
    JAVA="$JAVA_HOME/bin/java"
  fi

  if [ "x$DAEMON_MODE" = "xtrue" ]; then
    nohup "$JAVA" $JGC_HEAP_OPTS -jar "$JAR_EXECUTABLE" "$CONFIG_PATH" >/dev/null 2>&1 &
  else
    exec "$JAVA" $JGC_HEAP_OPTS -jar "$JAR_EXECUTABLE" "$CONFIG_PATH"
  fi
fi
