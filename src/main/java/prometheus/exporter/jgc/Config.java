/*
 * Copyright (C) 2024 The  jgc_exporter Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package prometheus.exporter.jgc;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class Config {
    public static final String DEFAULT_HOST_PORT = "0.0.0.0:5898";
    public static final int DEFAULT_IDLE_TIMEOUT = 3600_000;
    public static final int DEFAULT_BATCH_SIZE = 1024;
    public static final int DEFAULT_BUFFER_SIZE = 8192;
    public static final int DEFAULT_LINES_PER_SECOND = Integer.MAX_VALUE;
    public static final int DEFAULT_WATCH_INTERVAL = 30000;
    public static final int DEFAULT_READ_INTERVAL = 1_000;
    private String fileRegexPattern;
    private String fileGlobPattern;
    private String hostPort = DEFAULT_HOST_PORT;
    private int idleTimeout = DEFAULT_IDLE_TIMEOUT;
    private int batchSize = DEFAULT_BATCH_SIZE;
    private int bufferSize = DEFAULT_BUFFER_SIZE;
    private int linesPerSecond = DEFAULT_LINES_PER_SECOND;
    private int watchInterval = DEFAULT_WATCH_INTERVAL;
    private int readInterval = DEFAULT_READ_INTERVAL;
    private int scanFilesPerSecond;

    public String getFileRegexPattern() {
        return fileRegexPattern;
    }

    public void setFileRegexPattern(String fileRegexPattern) {
        this.fileRegexPattern = fileRegexPattern;
    }

    public String getFileGlobPattern() {
        return fileGlobPattern;
    }

    public void setFileGlobPattern(String fileGlobPattern) {
        this.fileGlobPattern = fileGlobPattern;
    }

    public String getHostPort() {
        return hostPort;
    }

    public void setHostPort(String hostPort) {
        this.hostPort = hostPort;
    }

    public int getIdleTimeout() {
        return idleTimeout;
    }

    public void setIdleTimeout(int idleTimeout) {
        this.idleTimeout = idleTimeout;
    }

    public int getBatchSize() {
        return batchSize;
    }

    public void setBatchSize(int batchSize) {
        this.batchSize = batchSize;
    }

    public int getBufferSize() {
        return bufferSize;
    }

    public void setBufferSize(int bufferSize) {
        this.bufferSize = bufferSize;
    }

    public int getLinesPerSecond() {
        return linesPerSecond;
    }

    public void setLinesPerSecond(int linesPerSecond) {
        this.linesPerSecond = linesPerSecond;
    }

    public int getWatchInterval() {
        return watchInterval;
    }

    public void setWatchInterval(int watchInterval) {
        this.watchInterval = watchInterval;
    }

    public int getReadInterval() {
        return readInterval;
    }

    public void setReadInterval(int readInterval) {
        this.readInterval = readInterval;
    }

    public int getScanFilesPerSecond() {
        return scanFilesPerSecond;
    }

    public void setScanFilesPerSecond(int scanFilesPerSecond) {
        this.scanFilesPerSecond = scanFilesPerSecond;
    }

    @Override
    public String toString() {
        return "Config{"
                + "fileRegexPattern='"
                + fileRegexPattern
                + '\''
                + ", fileGlobPattern='"
                + fileGlobPattern
                + '\''
                + ", hostPort='"
                + hostPort
                + '\''
                + ", idleTimeout="
                + idleTimeout
                + ", batchSize="
                + batchSize
                + ", bufferSize="
                + bufferSize
                + ", linesPerSecond="
                + linesPerSecond
                + ", watchInterval="
                + watchInterval
                + ", scanFilesPerSecond="
                + scanFilesPerSecond
                + '}';
    }
}
