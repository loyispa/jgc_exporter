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
package prometheus.exporter.jgc.parser;

import com.microsoft.gctoolkit.io.GCLogFile;
import com.microsoft.gctoolkit.io.LogFileMetadata;
import com.microsoft.gctoolkit.io.SingleGCLogFile;
import com.microsoft.gctoolkit.jvm.Diary;
import java.io.File;
import java.io.IOException;
import java.util.*;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.stream.Stream;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.tool.Config;

public class ContinousGCLogFile extends GCLogFile {
    private static final Logger LOG = LoggerFactory.getLogger(ContinousGCLogFile.class);
    private final BlockingQueue<String> inflightRecordQueue;
    private final BlockingQueue<String> stream;
    private Optional<GCLogAnalyser> analyser = Optional.empty();
    private final Config config;
    private final long analysePeriod;
    private final int inflightRecordLength;
    private long lastAnalyzed;
    private Diary diary;

    public ContinousGCLogFile(File file, Config config) {
        super(file.toPath());
        this.config = config;
        this.analysePeriod = config.getAnalysePeriod();
        this.inflightRecordLength = config.getInflightRecordLength();
        this.lastAnalyzed = System.currentTimeMillis();
        this.inflightRecordQueue = new LinkedBlockingQueue<>(inflightRecordLength);
        this.stream = new LinkedBlockingQueue<>();
        initDiary();
    }

    public boolean analyze() {
        if (ready()) {
            try {
                inflightRecordQueue.drainTo(stream);
                return getAnalyser().map(GCLogAnalyser::analyze).orElse(false);
            } finally {
                stream.clear();
                lastAnalyzed = System.currentTimeMillis();
            }
        }
        return false;
    }

    private boolean ready() {
        if (inflightRecordQueue.isEmpty()) {
            return false;
        } else if (inflightRecordQueue.size() > inflightRecordLength / 2.0) {
            return true;
        } else if (lastAnalyzed + analysePeriod < System.currentTimeMillis()) {
            return true;
        }
        return false;
    }

    public void append(String line) {
        try {
            inflightRecordQueue.put(line);
        } catch (InterruptedException ie) {
            LOG.error("GCLog append fail.");
        }
    }

    @Override
    public LogFileMetadata getMetaData() throws IOException {
        throw new UnsupportedOperationException("shouldn't call this method");
    }

    @Override
    public Diary diary() {
        return diary;
    }

    private void initDiary() {
        try {
            SingleGCLogFile logFile = new SingleGCLogFile(path);
            this.diary = logFile.diary();
            if (diary.isGenerationalKnown() || diary.isG1GCKnown() || diary.isZGCKnown()) {
                LOG.info("gc diary: {}", diary);
            } else {
                throw new IllegalArgumentException("unsupported gc log file: " + path);
            }
        } catch (IOException ex) {
            throw new IllegalArgumentException("invalid gc log file: " + path, ex);
        }
    }

    @Override
    public Stream<String> stream() {
        return Stream.concat(
                stream.stream()
                        .filter(Objects::nonNull)
                        .filter(line -> !line.isBlank())
                        .map(String::trim)
                        .filter(s -> s.length() > 0),
                Stream.of(endOfData()));
    }

    private Optional<GCLogAnalyser> getAnalyser() {
        if (analyser.isEmpty()) {
            try {
                analyser = Optional.of(new GCLogAnalyser(this));
            } catch (Throwable t) {
                LOG.error("Instance gc parser fail.", t);
            }
        }
        return analyser;
    }
}
