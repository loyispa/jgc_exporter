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
package prometheus.exporter.jgc.tailer;

import com.google.common.util.concurrent.RateLimiter;
import java.io.*;
import java.nio.file.Files;
import java.util.LinkedList;
import java.util.List;
import java.util.Objects;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Tailer {
    private static final Logger LOG = LoggerFactory.getLogger(Tailer.class);
    private static final byte BYTE_NL = (byte) 10;
    private static final byte BYTE_CR = (byte) 13;
    private static final int NEED_READING = -1;
    private final File file;
    private RandomAccessFile raf;
    private long inode;
    private int bufferPos;
    private int bufferCap;
    private final boolean seekToEnd;
    private final int batchSize;
    private final int bufferSize;
    private final byte[] readBuffer;
    private final LineBuffer lineBuffer;
    private final RateLimiter limiter;
    private long lastModified;

    public Tailer(File file, boolean seekToEnd, int batchSize, int bufferSize, int linesPerSecond) {
        this.file = Objects.requireNonNull(file);
        this.seekToEnd = seekToEnd;
        if (batchSize <= 0) {
            throw new IllegalArgumentException("batchSize");
        }
        if (bufferSize <= 0) {
            throw new IllegalArgumentException("bufferSize");
        }
        if (linesPerSecond <= 0) {
            throw new IllegalArgumentException("linesPerSecond");
        }
        this.batchSize = batchSize;
        this.bufferSize = bufferSize;
        this.readBuffer = new byte[bufferSize];
        this.lineBuffer = new LineBuffer();
        this.limiter = RateLimiter.create(linesPerSecond);
        refresh();
    }

    public List<String> readLines() throws IOException {
        List<String> lines = new LinkedList<>();
        for (int i = 0; i < batchSize; ++i) {
            String line = readLine();
            if (line == null) {
                break;
            }
            lines.add(line);
        }
        return lines;
    }

    public boolean rotated() {
        try {
            // inode changes
            long currInode = getInode(this.file);
            if (currInode != inode) {
                LOG.info("{} changed", this.file);
                return true;
            }

            // truncate occurs
            long fp = raf.getFilePointer();
            long len = raf.length();
            if (fp > len) {
                LOG.info("{} truncated", this.file);
                return true;
            }
        } catch (IOException ex) {
            LOG.error("IO error: {}", this.file, ex.getMessage());
        }
        return false;
    }

    public long lastModified() {
        lastModified = Math.max(lastModified, file.lastModified());
        return lastModified;
    }

    public File getFile() {
        return file;
    }

    public void close() {
        if (this.raf != null) {
            try {
                this.raf.close();
            } catch (IOException ignore) {
            } finally {
                this.raf = null;
            }
        }
    }

    private void refresh() {
        try {
            this.close();
            this.raf = new RandomAccessFile(file, "r");
            this.inode = getInode(file);
            if (seekToEnd) {
                this.raf.seek(this.raf.length());
            } else {
                this.raf.seek(0);
            }
            this.bufferPos = NEED_READING;
            this.bufferCap = 0;
            this.lastModified = file.lastModified();
        } catch (IOException ioe) {
            throw new RuntimeException("Failed init file: " + file);
        }
    }

    private String readLine() throws IOException {
        if (!limiter.tryAcquire()) {
            LOG.warn("Read frequency limit: {}", file);
            return null;
        }
        while (true) {
            if (bufferPos == NEED_READING) {
                if (raf.getFilePointer() < raf.length()) {
                    readFile();
                } else {
                    return null;
                }
            }
            for (int i = bufferPos; i < bufferCap; i++) {
                if (readBuffer[i] == BYTE_NL) {
                    // Don't copy last byte(NEW_LINE)
                    int lineLen = i - bufferPos;

                    lineBuffer.write(readBuffer, bufferPos, lineLen);

                    if (i + 1 < bufferCap) {
                        bufferPos = i + 1;
                    } else {
                        bufferPos = NEED_READING;
                    }

                    return lineBuffer.buildAndReset();
                }
            }

            if (bufferPos != NEED_READING) {
                lineBuffer.write(readBuffer, bufferPos, bufferCap - bufferPos);
            }

            bufferPos = NEED_READING;
        }
    }

    private void readFile() throws IOException {
        bufferCap = raf.read(readBuffer, 0, readBuffer.length);
        bufferPos = 0;
    }

    private long getInode(File file) throws IOException {
        return (long) Files.getAttribute(file.toPath(), "unix:ino");
    }

    class LineBuffer extends ByteArrayOutputStream {

        public LineBuffer() {
            super(bufferSize);
        }

        public String buildAndReset() {

            // For windows, check for CR
            int lineLen = count;
            if (lineLen > 0 && buf[lineLen - 1] == BYTE_CR) {
                lineLen -= 1;
            }
            String line = new String(buf, 0, lineLen);

            // Reset buffer
            if (buf.length > bufferSize) {
                buf = new byte[bufferSize];
            }
            count = 0;

            return line;
        }
    }

    @Override
    public String toString() {
        return "Tailer{"
                + "file='"
                + file
                + '\''
                + ", inode="
                + inode
                + ", lastUpdated="
                + lastModified
                + '}';
    }
}
