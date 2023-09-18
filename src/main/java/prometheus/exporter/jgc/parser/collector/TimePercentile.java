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
package prometheus.exporter.jgc.parser.collector;

import io.prometheus.client.SimpleCollector;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;

public class TimePercentile extends SimpleCollector<TimePercentile.Child> {
    private final long durationBetweenRotatesMillis;

    TimePercentile(Builder b) {
        super(b);
        this.durationBetweenRotatesMillis = b.durationBetweenRotatesMillis;
    }

    public double get() {
        return noLabelsChild.get();
    }

    public void add(double duration) {
        noLabelsChild.add(duration);
    }

    public static TimePercentile.Builder build() {
        return new TimePercentile.Builder();
    }

    public static class Builder
            extends SimpleCollector.Builder<TimePercentile.Builder, TimePercentile> {
        private long durationBetweenRotatesMillis = TimeUnit.MINUTES.toMillis(1);

        public Builder durationBetweenRotates(long durationBetweenRotatesMillis) {
            this.durationBetweenRotatesMillis = durationBetweenRotatesMillis;
            return this;
        }

        @Override
        public TimePercentile create() {
            return new TimePercentile(this);
        }
    }

    @Override
    protected Child newChild() {
        return new Child();
    }

    @Override
    public List<MetricFamilySamples> collect() {
        List<MetricFamilySamples.Sample> samples =
                new ArrayList<MetricFamilySamples.Sample>(children.size());
        for (Map.Entry<List<String>, TimePercentile.Child> c : children.entrySet()) {
            samples.add(
                    new MetricFamilySamples.Sample(
                            fullname, labelNames, c.getKey(), c.getValue().get()));
        }
        return familySamplesList(Type.GAUGE, samples);
    }

    public class Child {
        private long lastRotateTimestampMillis;
        private double currValue;
        private double lastValue;

        public synchronized void add(double duration) {
            currValue += duration;
        }

        public synchronized double get() {
            long timeSinceLastRotateMillis = System.currentTimeMillis() - lastRotateTimestampMillis;
            if (timeSinceLastRotateMillis < durationBetweenRotatesMillis) {
                return lastValue;
            }
            lastRotateTimestampMillis = System.currentTimeMillis();
            lastValue = currValue / timeSinceLastRotateMillis;
            currValue = 0;
            return lastValue;
        }
    }
}
