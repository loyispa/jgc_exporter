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
package prometheus.exporter.jgc.parser;

import com.microsoft.gctoolkit.aggregator.Aggregates;
import com.microsoft.gctoolkit.aggregator.Aggregator;
import com.microsoft.gctoolkit.aggregator.EventSource;
import com.microsoft.gctoolkit.event.g1gc.G1GCEvent;
import com.microsoft.gctoolkit.event.generational.GenerationalGCEvent;
import com.microsoft.gctoolkit.event.jvm.Safepoint;
import com.microsoft.gctoolkit.event.jvm.SurvivorRecord;
import com.microsoft.gctoolkit.event.zgc.ZGCCycle;

@Aggregates({
    EventSource.CMS_PREUNIFIED,
    EventSource.CMS_UNIFIED,
    EventSource.G1GC,
    EventSource.GENERATIONAL,
    EventSource.ZGC,
    EventSource.SURVIVOR,
    EventSource.SAFEPOINT,
    EventSource.TENURED
})
public class GCEventAggregator extends Aggregator<GCEventAggregation> {

    public GCEventAggregator(GCEventAggregation aggregation) {
        super(aggregation);
        register(GenerationalGCEvent.class, aggregation::collectGenerationalGCEvent);
        register(G1GCEvent.class, aggregation::collectG1GCEvent);
        register(ZGCCycle.class, aggregation::collectZGCEvent);
        register(Safepoint.class, aggregation::collectSafePoint);
        register(SurvivorRecord.class, aggregation::collectSurvivorRecord);
    }
}
