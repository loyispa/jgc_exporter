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
package prometheus.exporter.jgc.collector;

import java.io.File;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.tailer.TailerListener;

public class GCCollectorManager implements TailerListener {
    private static final Logger LOG = LoggerFactory.getLogger(GCCollectorManager.class);

    private final Map<File, GCCollector> registry;

    public GCCollectorManager() {
        this.registry = new ConcurrentHashMap<>();
    }

    @Override
    public void onOpen(File file) {
        registry.computeIfAbsent(file, f -> new GCCollector(file));
        LOG.info("Register file: {}", file);
    }

    @Override
    public void onClose(File file) {
        GCCollector collector = registry.remove(file);
        if (collector != null) {
            collector.close();
        }
        LOG.info("Unregister file: {}", file);
    }

    @Override
    public void onRotate(File file) {
        // not clean metric
        registry.remove(file);
        LOG.info("Rotate file: {}", file);
    }

    @Override
    public void onRead(File file, String line) {
        LOG.debug("Tailing file: {} >>> {}", file, line);
        registry.computeIfPresent(
                file,
                (f, collector) -> {
                    collector.receive(line);
                    return collector;
                });
    }
}
