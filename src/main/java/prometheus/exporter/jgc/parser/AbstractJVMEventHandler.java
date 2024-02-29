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

import static prometheus.exporter.jgc.metric.MetricRegistry.GC_LOG_LINES;

import com.microsoft.gctoolkit.jvm.Diary;
import com.microsoft.gctoolkit.message.DataSourceParser;
import com.microsoft.gctoolkit.message.JVMEventChannel;
import com.microsoft.gctoolkit.message.JVMEventChannelListener;
import java.io.File;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.metric.MetricRegistry;
import prometheus.exporter.jgc.util.OperatingSystem;

public abstract class AbstractJVMEventHandler implements JVMEventChannel {
    private static final Logger LOG = LoggerFactory.getLogger(AbstractJVMEventHandler.class);
    protected final List<DataSourceParser> parsers;
    protected final Diary diary;
    protected final String path;
    protected final String host;

    protected AbstractJVMEventHandler(File file, Diary diary) {
        this.path = file.getPath();
        this.host = OperatingSystem.getLocalHostName();
        this.diary = diary;
        this.parsers = loadParsers();
        initialize();
    }

    protected abstract List<DataSourceParser> loadParsers();

    protected void initialize() {
        for (DataSourceParser parser : parsers) {
            if (!parser.accepts(diary)) {
                throw new UnsupportedOperationException();
            }
            parser.diary(diary);
            parser.publishTo(this);
        }
    }

    public AbstractJVMEventHandler consume(String message) {
        GC_LOG_LINES.attach(this, path, host).inc();
        for (DataSourceParser parser : parsers) {
            try {
                parser.receive(message);
            } catch (Exception ex) {
                LOG.error("{} error: {}", parser.getClass().getSimpleName(), message, ex);
            }
        }
        return this;
    }

    @Override
    public void registerListener(JVMEventChannelListener listener) {
        throw new UnsupportedOperationException();
    }

    @Override
    public void close() {
        MetricRegistry.detach(this);
    }
}
