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

import static prometheus.exporter.jgc.metric.MetricRegistry.*;

import com.microsoft.gctoolkit.event.MemoryPoolSummary;
import com.microsoft.gctoolkit.event.ReferenceGCSummary;
import com.microsoft.gctoolkit.event.generational.*;
import com.microsoft.gctoolkit.event.jvm.JVMEvent;
import com.microsoft.gctoolkit.jvm.Diary;
import com.microsoft.gctoolkit.message.ChannelName;
import com.microsoft.gctoolkit.message.DataSourceParser;
import com.microsoft.gctoolkit.parser.CMSTenuredPoolParser;
import com.microsoft.gctoolkit.parser.GenerationalHeapParser;
import com.microsoft.gctoolkit.parser.UnifiedGenerationalParser;
import java.io.File;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ClassicGCEventHandler extends AbstractJVMEventHandler {
    private static final Logger LOG = LoggerFactory.getLogger(ClassicGCEventHandler.class);

    public ClassicGCEventHandler(File file, Diary diary) {
        super(file, diary);
    }

    @Override
    protected List<DataSourceParser> loadParsers() {
        if (diary.isUnifiedLogging()) {
            return Collections.singletonList(new UnifiedGenerationalParser());
        } else if (diary.isCMS()) {
            // TODO: remove CMSTenuredPoolParser while gctoolkit release newly
            return Arrays.asList(new CMSTenuredPoolParser(), new GenerationalHeapParser());
        } else {
            return Arrays.asList(new GenerationalHeapParser());
        }
    }

    @Override
    public void publish(ChannelName channel, JVMEvent event) {
        if (event instanceof GenerationalGCEvent) {
            recordGCEvent((GenerationalGCEvent) event);
        } else {
            LOG.warn("{} published an unsupported event: {}", channel, event);
        }
    }

    private void recordGCEvent(GenerationalGCEvent event) {

        String category = parseCategory(event);
        LOG.debug("Collect ClassicGCEvent {}", category);
        GC_EVENT_DURATION.attach(this, path, host, category).observe(event.getDuration());
        GC_EVENT_LAST_MINUTE_DURATION.attach(this, path, host).observe(event.getDuration());

        if (event instanceof GenerationalGCPauseEvent) {
            GC_EVENT_PAUSE_DURATION.attach(this, path, host, category).observe(event.getDuration());
            recordClassicGCPauseEvent((GenerationalGCPauseEvent) event);
        }
    }

    private String parseCategory(GenerationalGCEvent event) {

        if (event instanceof CMSPhase
                || event instanceof ConcurrentModeFailure
                || event instanceof ConcurrentModeInterrupted) {
            return parseCMSCategory(event);
        }

        if (event instanceof GenerationalGCPauseEvent) {
            if (event instanceof FullGC) {
                if (event instanceof PSFullGC) {
                    return "PSFullGC";
                } else if (event instanceof SystemGC) {
                    return "SystemGC";
                }
                return "FullGC";
            } else if (event instanceof ParNew) {
                if (event instanceof ParNewPromotionFailed) {
                    return "ParNewPromotionFailed";
                }
                return "ParNew";
            } else if (event instanceof PSYoungGen) {
                return "PSYoungGC";
            } else if (event instanceof DefNew) {
                return "DefNew";
            } else if (event instanceof YoungGC) {
                return "YoungGC";
            }
        }
        return "Unknown";
    }

    private String parseCMSCategory(GenerationalGCEvent event) {
        if (event instanceof ConcurrentModeFailure) {
            return "CMSConcurrentModeFailure";
        } else if (event instanceof ConcurrentModeInterrupted) {
            return "CMSConcurrentModeInterrupted";
        } else if (event instanceof InitialMark) {
            return "CMSInitialMark";
        } else if (event instanceof CMSRemark) {
            return "CMSRemark";
        } else if (event instanceof AbortablePreClean) {
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
        return "CMSUnknown";
    }

    protected void recordClassicGCPauseEvent(GenerationalGCPauseEvent event) {

        MemoryPoolSummary heapSummary = event.getHeap();
        if (heapSummary != null) {
            HEAP_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(heapSummary.getOccupancyAfterCollection() * 1024);
            HEAP_SIZE_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(heapSummary.getSizeAfterCollection() * 1024);
            HEAP_SIZE_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(heapSummary.getSizeBeforeCollection() * 1024);
            HEAP_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(heapSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary tenuredSummary = event.getTenured();
        if (tenuredSummary != null) {
            OLD_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(tenuredSummary.getOccupancyAfterCollection() * 1024);
            OLD_SIZE_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(tenuredSummary.getSizeAfterCollection() * 1024);
            OLD_SIZE_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(tenuredSummary.getSizeBeforeCollection() * 1024);
            OLD_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(tenuredSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary youngSummary = event.getYoung();
        if (youngSummary != null) {
            YOUNG_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(youngSummary.getOccupancyAfterCollection() * 1024);
            YOUNG_SIZE_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(youngSummary.getSizeAfterCollection() * 1024);
            YOUNG_SIZE_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(youngSummary.getSizeBeforeCollection() * 1024);
            YOUNG_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(youngSummary.getOccupancyBeforeCollection() * 1024);
        }

        MemoryPoolSummary permOrMetaspace = event.getPermOrMetaspace();
        if (permOrMetaspace != null) {
            METASPACE_OCCUPANCY_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(permOrMetaspace.getOccupancyAfterCollection() * 1024);
            METASPACE_SIZE_AFTER_COLLECTION
                    .attach(this, path, host)
                    .set(permOrMetaspace.getSizeAfterCollection() * 1024);
            METASPACE_SIZE_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(permOrMetaspace.getSizeBeforeCollection() * 1024);
            METASPACE_OCCUPANCY_BEFORE_COLLECTION
                    .attach(this, path, host)
                    .set(permOrMetaspace.getOccupancyBeforeCollection() * 1024);
        }

        if (event instanceof CMSPhase) {
            double classUnloadingProcessingTime = event.getClassUnloadingProcessingTime();
            double symbolTableProcessingTime = event.getSymbolTableProcessingTime();
            double stringTableProcessingTime = event.getStringTableProcessingTime();
            double symbolAndStringTableProcessingTime =
                    event.getSymbolAndStringTableProcessingTime();
            if (classUnloadingProcessingTime > 0.0) {
                CMS_CLASS_UNLOADING_PROCESS_TIME
                        .attach(this, path, host)
                        .observe(classUnloadingProcessingTime);
            }
            if (symbolTableProcessingTime > 0.0) {
                CMS_SYMBOL_TABLE_PROCESS_TIME
                        .attach(this, path, host)
                        .observe(symbolTableProcessingTime);
            }
            if (stringTableProcessingTime > 0.0) {
                CMS_STRING_TABLE_PROCESS_TIME
                        .attach(this, path, host)
                        .observe(stringTableProcessingTime);
            }
            if (symbolAndStringTableProcessingTime > 0.0) {
                CMS_SYMBOL_AND_STRING_TABLE_PROCESS_TIME
                        .attach(this, path, host)
                        .observe(symbolAndStringTableProcessingTime);
            }
        }

        ReferenceGCSummary referenceGCSummary = event.getReferenceGCSummary();
        recordReferenceSummary(referenceGCSummary);
    }

    private void recordReferenceSummary(ReferenceGCSummary referenceGCSummary) {
        if (referenceGCSummary != null) {
            int softReferenceCount = referenceGCSummary.getSoftReferenceCount();
            SOFT_REFERENCE_COUNT.attach(this, path, host).set(softReferenceCount);
            double softReferencePauseTime = referenceGCSummary.getSoftReferencePauseTime();
            SOFT_REFERENCE_PAUSE_TIME.attach(this, path, host).observe(softReferencePauseTime);

            int weakReferenceCount = referenceGCSummary.getWeakReferenceCount();
            WEAK_REFERENCE_COUNT.attach(this, path, host).set(weakReferenceCount);
            double weakReferencePauseTime = referenceGCSummary.getWeakReferencePauseTime();
            WEAK_REFERENCE_PAUSE_TIME.attach(this, path, host).observe(weakReferencePauseTime);

            int finalReferenceCount = referenceGCSummary.getFinalReferenceCount();
            FINAL_REFERENCE_COUNT.attach(this, path, host).set(finalReferenceCount);
            double finalReferencePauseTime = referenceGCSummary.getFinalReferencePauseTime();
            FINAL_REFERENCE_PAUSE_TIME.attach(this, path, host).observe(finalReferencePauseTime);

            int phantomReferenceCount = referenceGCSummary.getPhantomReferenceCount();
            PHANTOM_REFERENCE_COUNT.attach(this, path, host).set(phantomReferenceCount);
            int phantomReferenceFreedCount = referenceGCSummary.getPhantomReferenceFreedCount();
            PHANTOM_REFERENCE_FREE_COUNT.attach(this, path, host).set(phantomReferenceFreedCount);
            double phantomReferencePauseTime = referenceGCSummary.getPhantomReferencePauseTime();
            PHANTOM_REFERENCE_PAUSE_TIME
                    .attach(this, path, host)
                    .observe(phantomReferencePauseTime);

            int jniWeakReferenceCount = referenceGCSummary.getJniWeakReferenceCount();
            JNI_WEAK_REFERENCE_COUNT.attach(this, path, host).set(jniWeakReferenceCount);
            double jniWeakReferencePauseTime = referenceGCSummary.getJniWeakReferencePauseTime();
            JNI_WEAK_REFERENCE_PAUSE_TIME
                    .attach(this, path, host)
                    .observe(jniWeakReferencePauseTime);
        }
    }
}
