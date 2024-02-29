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
package prometheus.exporter.jgc.parser;

import com.microsoft.gctoolkit.io.GCLogFile;
import com.microsoft.gctoolkit.io.LogFileMetadata;
import com.microsoft.gctoolkit.io.SingleLogFileMetadata;
import com.microsoft.gctoolkit.jvm.Diary;
import java.io.IOException;
import java.io.RandomAccessFile;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class GCEventHandlerMatcher extends GCLogFile {
    private static final Logger LOG = LoggerFactory.getLogger(GCEventHandlerMatcher.class);
    private static final int MAX_LINES = 512;
    private final List<String> lines;

    public GCEventHandlerMatcher(Path path) {
        super(path);
        this.lines = readLines();
    }

    @Override
    public LogFileMetadata getMetaData() throws IOException {
        return new SingleLogFileMetadata(path);
    }

    @Override
    public Stream<String> stream() throws IOException {
        return Stream.concat(lines.stream(), Stream.of(endOfData()));
    }

    private List<String> readLines() {
        List<String> lines = new ArrayList<>();
        try (RandomAccessFile raf = new RandomAccessFile(path.toFile(), "r")) {
            for (int i = 0; i < MAX_LINES; ++i) {
                String line = readLine(raf);
                if (line == null) {
                    break;
                } else if (line.isBlank()) {
                    continue;
                }
                lines.add(line);
            }
            return lines;
        } catch (Exception ex) {
            throw new UnsupportedOperationException(ex);
        }
    }

    private String readLine(RandomAccessFile raf) throws IOException {
        StringBuilder input = new StringBuilder();
        int c = -1;
        boolean eol = false;

        while (!eol) {
            switch (c = raf.read()) {
                case -1:
                case '\n':
                    eol = true;
                    break;
                case '\r':
                    eol = true;
                    long cur = raf.getFilePointer();
                    if ((raf.read()) != '\n') {
                        raf.seek(cur);
                    }
                    break;
                default:
                    input.append((char) c);
                    break;
            }
            if (input.length() > 1024) {
                throw new IllegalArgumentException("Too large line: " + path);
            }
        }

        if ((c == -1) && (input.length() == 0)) {
            return null;
        }
        return input.toString();
    }

    public AbstractJVMEventHandler find() {
        try {
            Diary diary = super.diary();
            return matchAny(diary);
        } catch (IOException ioe) {
            LOG.error("Find parser error: {}", path, ioe);
            throw new UnsupportedOperationException(path.toString(), ioe);
        }
    }

    private AbstractJVMEventHandler matchAny(Diary diary) {
        if (diary.isG1GC()) {
            LOG.info("{} is G1", path);
            return new G1GCEventHandler(path.toFile(), diary);
        } else if (diary.isZGC()) {
            LOG.info("{} is ZGC", path);
            return new ZGCEventHandler(path.toFile(), diary);
        } else if (diary.isCMS() || diary.isParNew()) {
            LOG.info("{} is CMS", path);
            return new CMSGCEventHandler(path.toFile(), diary);
        } else if (diary.isDefNew()) {
            LOG.info("{} is defnew", path);
            return new ParallelAndSerialGCEventHandler(path.toFile(), diary);
        } else if (diary.isSerialFull()) {
            LOG.info("{} is serial", path);
            return new ParallelAndSerialGCEventHandler(path.toFile(), diary);
        } else if (diary.isPSOldGen() || diary.isPSYoung()) {
            LOG.info("{} is parallel", path);
            return new ParallelAndSerialGCEventHandler(path.toFile(), diary);
        }
        LOG.info("Unmatched gc log: {}", path);
        throw new UnsupportedOperationException(path.toString());
    }
}
