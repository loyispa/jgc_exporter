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
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;

public class SystemCollector extends Collector {
    private final List<Collector.MetricFamilySamples> metricFamilySamples = new ArrayList<>();

    public SystemCollector() {
        GaugeMetricFamily startupTimestamp =
                new GaugeMetricFamily(
                                "jgc_startup_timestamp_seconds",
                                "Timestamp of exporter startup",
                                Collections.emptyList())
                        .addMetric(
                                Collections.emptyList(),
                                System.currentTimeMillis() / MILLISECONDS_PER_SECOND);

        GaugeMetricFamily versionInfo =
                new GaugeMetricFamily(
                        "jgc_exporter_version_info",
                        "The version of jgc_exporter",
                        Arrays.asList("version"));

        Package pkg = this.getClass().getPackage();
        String version = pkg.getImplementationVersion();

        versionInfo.addMetric(Arrays.asList(version != null ? version : "unknown"), 1L);

        metricFamilySamples.add(versionInfo);
        metricFamilySamples.add(startupTimestamp);
    }

    @Override
    public List<MetricFamilySamples> collect() {
        return metricFamilySamples;
    }
}
