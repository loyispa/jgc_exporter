/*
 * Copyright (C) 2023 The  jgc_exporter Authors
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
package prometheus.exporter.jgc.tool;

public class Config {
    public static final String DEFAULT_HOST_PORT = "0.0.0.0:5898";
    public static final int DEFAULT_IDLE_TIMEOUT = 600_000;
    public static final int DEFAULT_BATCH_SIZE = 128;
    public static final int DEFAULT_BUFFER_SIZE = 8192;
    public static final int DEFAULT_ANALYSE_PERIOD = 10_000;
    public static final int DEFAULT_INFLIGHT_RECORD_LENGTH = 1024;
    private String fileRegexPattern;
    private String hostPort = DEFAULT_HOST_PORT;
    private int idleTimeout = DEFAULT_IDLE_TIMEOUT;
    private int batchSize = DEFAULT_BATCH_SIZE;
    private int bufferSize = DEFAULT_BUFFER_SIZE;
    private int analysePeriod = DEFAULT_ANALYSE_PERIOD;
    private int inflightRecordLength = DEFAULT_INFLIGHT_RECORD_LENGTH;

    public String getFileRegexPattern() {
        return fileRegexPattern;
    }

    public void setFileRegexPattern(String fileRegexPattern) {
        this.fileRegexPattern = fileRegexPattern;
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

    public int getAnalysePeriod() {
        return analysePeriod;
    }

    public void setAnalysePeriod(int analysePeriod) {
        this.analysePeriod = analysePeriod;
    }

    public int getInflightRecordLength() {
        return inflightRecordLength;
    }

    public void setInflightRecordLength(int inflightRecordLength) {
        this.inflightRecordLength = inflightRecordLength;
    }

    @Override
    public String toString() {
        return "Config{"
                + "fileRegexPattern='"
                + fileRegexPattern
                + '\''
                + ", hostPort="
                + hostPort
                + ", idleTimeout="
                + idleTimeout
                + ", batchSize="
                + batchSize
                + ", bufferSize="
                + bufferSize
                + ", analysePeriod="
                + analysePeriod
                + ", inflightRecordLength="
                + inflightRecordLength
                + '}';
    }
}
