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
package prometheus.exporter.jgc.metric;

import io.prometheus.client.Collector;
import io.prometheus.client.SimpleCollector;
import io.prometheus.client.Supplier;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

public class Metric<C, T extends SimpleCollector<C>> extends Collector {
    private final Supplier<T> supplier;
    private final Map<Object, T> targets;

    private Metric(Supplier<T> supplier) {
        this.supplier = Objects.requireNonNull(supplier);
        this.targets = new ConcurrentHashMap<>();
    }

    public C attach(Object target, String... labels) {
        return targets.computeIfAbsent(Objects.requireNonNull(target), key -> supplier.get())
                .labels(labels);
    }

    public T detach(Object target) {
        return targets.remove(target);
    }

    @Override
    public List<MetricFamilySamples> collect() {

        final Optional<MetricFamilySamples> mfsOpt =
                targets.values().stream()
                        .map(SimpleCollector::collect)
                        .filter(mfs -> !mfs.isEmpty())
                        .map(mfs -> mfs.get(0))
                        .findAny();

        if (mfsOpt.isEmpty()) {
            return Collections.emptyList();
        }

        List<MetricFamilySamples.Sample> samples = new ArrayList<>();
        for (T target : targets.values()) {
            for (MetricFamilySamples m : target.collect()) {
                samples.addAll(m.samples);
            }
        }

        MetricFamilySamples mfs = copyOf(mfsOpt.get(), samples);
        List<MetricFamilySamples> mfsList = new ArrayList<>(1);
        mfsList.add(mfs);
        return mfsList;
    }

    private MetricFamilySamples copyOf(
            MetricFamilySamples mfs, List<MetricFamilySamples.Sample> samples) {
        return new MetricFamilySamples(mfs.name, mfs.unit, mfs.type, mfs.help, samples);
    }

    public static <C, T extends SimpleCollector<C>> Metric<C, T> of(Supplier<T> supplier) {
        return new Metric<>(supplier).register(MetricRegistry.SINGLETON);
    }
}
