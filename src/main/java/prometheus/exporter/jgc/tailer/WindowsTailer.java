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
package prometheus.exporter.jgc.tailer;

import java.io.File;
import java.io.IOException;
import java.io.RandomAccessFile;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class WindowsTailer extends Tailer {
    private static final Logger LOG = LoggerFactory.getLogger(WindowsTailer.class);
    private long filePointer;

    public WindowsTailer(
            File file, boolean seekToEnd, int batchSize, int bufferSize, int linesPerSecond) {
        super(file, seekToEnd, batchSize, bufferSize, linesPerSecond);
        releaseFile();
    }

    @Override
    public List<String> readLines() throws IOException {
        try {
            holdFile();
            return super.readLines();
        } finally {
            releaseFile();
        }
    }

    @Override
    public boolean rotated() {
        try {
            holdFile();
            return super.rotated();
        } catch (IOException ex) {
            LOG.error("IO error: {}", this.file, ex);
        } finally {
            releaseFile();
        }
        return false;
    }

    private void releaseFile() {
        if (this.raf != null) {
            try {
                this.filePointer = this.raf.getFilePointer();
                this.raf.close();
            } catch (IOException ignore) {
            } finally {
                this.raf = null;
            }
        }
    }

    private void holdFile() throws IOException {
        this.raf = new RandomAccessFile(file, "r");
        this.raf.seek(filePointer);
    }
}
