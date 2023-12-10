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
package prometheus.exporter.jgc.collector;

import io.prometheus.client.Collector;
import io.prometheus.client.GaugeMetricFamily;
import java.io.File;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import prometheus.exporter.jgc.collector.parser.GCAggregator;
import prometheus.exporter.jgc.tailer.TailerListener;

public class GCCollector extends Collector implements TailerListener {
    private static final Logger LOG = LoggerFactory.getLogger(GCCollector.class);

    private final Map<File, GCAggregator> registry;

    public GCCollector() {
        this.registry = new ConcurrentHashMap<>();
    }

    @Override
    public List<MetricFamilySamples> collect() {
        GaugeMetricFamily collectFiles =
                new GaugeMetricFamily(
                        "jgc_collect_files", "jgc_collect_files", Arrays.asList("path"));
        registry.forEach(
                (file, aggregator) -> {
                    collectFiles.addMetric(Arrays.asList(file.getPath()), 1);
                });
        return Arrays.asList(collectFiles);
    }

    @Override
    public void onOpen(File file) {
        registry.computeIfAbsent(file, f -> new GCAggregator(file));
        LOG.info("Register gc log: {}", file);
    }

    @Override
    public void onClose(File file) {
        GCAggregator aggregator = registry.remove(file);
        if (aggregator != null) {
            aggregator.close();
        }
        LOG.info("Unregister gc log: {}", file);
    }

    @Override
    public void onRead(File file, String line) {
        LOG.debug("Tailing gc log: {} >>> {}", file, line);
        registry.computeIfPresent(
                file,
                (f, collector) -> {
                    collector.receive(line);
                    return collector;
                });
    }
}
