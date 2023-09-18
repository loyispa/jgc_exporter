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
package prometheus.exporter.jgc.tool;

import com.microsoft.gctoolkit.event.g1gc.*;
import com.microsoft.gctoolkit.event.generational.*;
import com.microsoft.gctoolkit.event.zgc.ZGCCycle;
import io.prometheus.client.Collector;

public class GcEventTool {
    private GcEventTool() {}

    public static String parseGCEventCategory(GenerationalGCEvent event) {

        if (event instanceof GenerationalGCPauseEvent) {

            if (event instanceof FullGC) {
                return "FullGC";
            } else if (event instanceof InitialMark) {
                return "CMSInitialMark";
            } else if (event instanceof CMSRemark) {
                return "CMSRemark";
            } else if (event instanceof DefNew
                    || event instanceof ParNew
                    || event instanceof PSYoungGen
                    || event instanceof YoungGC) {
                return "YoungGC";
            }

        } else if (event instanceof CMSConcurrentEvent) {
            if (event instanceof AbortablePreClean) {
                return "CMSAbortablePreClean";
            } else if (event instanceof ConcurrentMark) {
                return "CMSConcurrentMark";
            } else if (event instanceof ConcurrentSweep) {
                return "CMSConcurrentSweep";
            } else if (event instanceof ConcurrentPreClean) {
                return "CMSConcurrentPreClean";
            } else if (event instanceof ConcurrentReset) {
                return "CMSConcurrentReset";
            }
        }

        return Collector.sanitizeMetricName(event.getGarbageCollectionType().name());
    }

    public static String parseGCEventCategory(G1GCEvent event) {
        if (event instanceof G1GCPauseEvent) {
            if (event instanceof G1Young) {
                if (event instanceof G1YoungInitialMark) {
                    return "G1InitialMark";
                } else if (event instanceof G1Mixed) {
                    return "G1MixedGC";
                }
                return "YoungGC";
            } else if (event instanceof G1Cleanup) {
                return "G1Cleanup";
            } else if (event instanceof G1Remark) {
                return "G1Cleanup";
            } else if (event instanceof G1FullGC) {
                return "FullGC";
            }
        } else if (event instanceof G1GCConcurrentEvent) {
            String category = event.getClass().getSimpleName();
            if (!category.startsWith("G1")) {
                category = "G1" + category;
            }
            return category;
        }
        return Collector.sanitizeMetricName(event.getGarbageCollectionType().name());
    }

    public static String parseGCEventCategory(ZGCCycle event) {
        return "zgc";
    }
}
