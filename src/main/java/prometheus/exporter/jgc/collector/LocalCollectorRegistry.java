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

import static io.prometheus.client.Collector.*;

import io.prometheus.client.*;
import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.function.Predicate;

public class LocalCollectorRegistry extends CollectorRegistry {
    public static final LocalCollectorRegistry DEFAULT = new LocalCollectorRegistry();

    private final List<SimpleCollector> collectors;

    private LocalCollectorRegistry() {
        super(true);
        this.collectors = new CopyOnWriteArrayList<>();
    }

    @Override
    public void register(Collector m) {
        super.register(m);
        if (m instanceof SimpleCollector) {
            collectors.add((SimpleCollector) m);
        }
    }

    @Override
    public void unregister(Collector m) {
        super.unregister(m);
        collectors.remove(m);
    }

    public void clean(Predicate<MetricFamilySamples.Sample> predicate) {
        collectors.forEach(
                collector -> {
                    for (MetricFamilySamples samples : collector.collect()) {
                        for (MetricFamilySamples.Sample sample : samples.samples) {
                            if (predicate.test(sample)) {
                                String[] labelValues = sample.labelValues.toArray(new String[] {});
                                collector.remove(labelValues);
                            }
                        }
                    }
                });
    }
}
